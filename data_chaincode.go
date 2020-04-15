package main

import (
	"encoding/json"
	"errors"
	"strings"
	"fmt"
		
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// DataChainCode example simple Chaincode implementation
type DataChainCode struct {
}

//Struct for location vals
type LocationData struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

// Product - product that is written to the ledger,  Data contains non static type files
type Product struct {
	ID           float64                `json:"id"`
	Gtin         string                 `json:"gtin"`
	Lot          string                 `json:"lot"`
	SerialNumber float64                `json:"serialNo"`
	ExpiryDate   string                 `json:"expirationDate"`
	Event        string                 `json:"event"`
	Gln          string                 `json:"gln"`
	Status       string                 `json:"status"`
	TradeItemDesc string                `json:"tradeItemDesc"`
	Product      string                 `json:"product"`
	TradeName    string                 `json:"tradename"`
	ManufactureDate string              `json:"manufactureDate"`
	Location     string                 `json:"location"`
	ToGln        string                 `json:"toGln"`
	ToLocation   string                 `json:"toLocation"`
	Sender       string                 `json:"sender"`
	Receiver     string                 `json:"receiver"`
	LocationInfo LocationData           `json:"loc_cd"`
	EventDate    string                 `json:"event_dt"`
	Data         map[string]interface{} `json:"-"` // Unknown fields should go here.
}

// ProductKey - this struct represents the product key
type ProductKey struct {
	Key          string
	Gtin         string
	SerialNumber string
	LotNumber    string
	ExpiryDate   string
}

type ReturnVal struct {
	DataHash      string `json:"dataHash"`
	TransactionId string `json:"transactionId"`
}

var logger = shim.NewLogger("data_chaincode-cc")

// ProductObjectType - defines the project object type
const ProductObjectType = "product"

// MaxProductJSONSizeAllowed - defines the max JSON size allowed for an input
const MaxProductJSONSizeAllowed = 2048

// Init is called with the chaincode is instantiated or updated.
// It can be used to initialize data for the chaincode for real products or test
// For this we don't need to pre-populate anything
func (t *DataChainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("Init: enter")
	defer logger.Info("Init: exit")
	return shim.Success(nil)
} // end of init

// Invoke is called per transaction on the chaincode.
func (t *DataChainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("Invoke: enter")
	defer logger.Info("Invoke: exit")

	function, args := stub.GetFunctionAndParameters()
	txID := stub.GetTxID()

	logger.Debug("Invoke: Transaction ID: ", txID)
	logger.Debug("Invoke: function: ", function)
	logger.Debug("Invoke: args count: ", len(args))
	logger.Debug("Invoke: args found: ", args)

	if function == "createProduct" {
		return t.createProduct(stub, args)
	} 

	logger.Error("Invoke: Invalid function = " + function)
	return shim.Error("Invoke: Invalid function = " + function)

} // end of invoke

// ============================================================================================================================
// Create Product - passes 2 arguments first is the key the second is JSON data mapping to the defined structure above
// ============================================================================================================================
// takes a single argument that is JSON of the product to create
func (t *DataChainCode) createProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("createProduct: enter")
	defer logger.Info("createProduct: exit")

	if len(args) != 1 {
		errorString := "createProduct: Invalid number of args, must be exactly 1 argument containing JSON containing data"
		logger.Error(errorString)
		return shim.Error(errorString)
	}

	var productInput = args[0]
	product, err := getProductFromJSON([]byte(productInput))
	if err != nil {
		logger.Error("createProduct: Error with JSON format:", err)
		return shim.Error(err.Error())
	}

	key := getProductKey(product)
	bytes, err := product.toBytes()
	if err != nil {
		logger.Error("createProduct: Error converting input product to bytes:", err)
		return shim.Error(err.Error())
	}
	logger.Info("createProduct: call putState, key = ", key)
	// write it to the ledger
	logger.Debug("createProduct: call putState, key = ", key)
	err = stub.PutState(key, bytes)
	if err != nil {
		logger.Error("createProduct: Error invoking on chaincode:", err)
		return shim.Error(err.Error())
	}
	logger.Info("transaction id", stub.GetTxID())
	logger.Info("createProduct: return successful write")
	//return shim.Success(bytes)
	returnVal := ReturnVal{string(bytes), stub.GetTxID()}
	logger.Info("return val Before Marshal : ", returnVal)
	rv, err := json.Marshal(returnVal)
	logger.Info("return val", rv)
	//return shim.Success([]byte(rv))
	return shim.Success([]byte(stub.GetTxID()))
} // end of createProduct



// ============================================================================================================================
// Get Product - get a product asset from ledger
// ============================================================================================================================
func getProduct(stub shim.ChaincodeStubInterface, key string) (Product, error) {

	logger.Info("getProduct: enter")
	defer logger.Info("getProduct: exit")

	var prod Product

	productAsBytes, err := stub.GetState(key)
	if err != nil { //this seems to always succeed, even if key didn't exist
		return prod, errors.New("getProduct: Failed to find product - " + key)
	}

	if len(productAsBytes) == 0 {
		return prod, errors.New("getProduct: Error processing request, no results found")
	}

	prod, err = getProductFromJSON(productAsBytes)
	if err != nil {
		logger.Error("getProduct: Error with JSON format:", err)
		return prod, errors.New(err.Error())
	}

	//test if product is actually here or just nil
	if len(prod.Gtin) < 1 {
		return prod, errors.New("getProduct:Product does not exist - " + key)
	}

	return prod, nil
}

// Since we have dynamic data we unmarshall into the Data field for everything
// then manually set the know types from the Data then remove them from Data so
// unmarshalling only occurs once
// NOTE: this method just unmashalls and does not validate, so if missing field it return empty string
func getProductFromJSON(incoming []byte) (Product, error) {
	var product Product
	logger.Info("product in getProductFromJSON", product)

	if err := json.Unmarshal([]byte(incoming), &product.Data); err != nil {
		return product, err
	}

	var loc LocationData
	if val, ok := product.Data["loc_cd"]; ok {
		if err := json.Unmarshal([]byte(val), &loc); err != nil {
			return loc, err
		}
		product.LocationInfo = loc
		delete(product.Data, "loc_cd")
	} else {
		product.LocationInfo = loc
	}
	
	if val, ok := product.Data["id"]; ok {
		product.ID = val.(float64)
		delete(product.Data, "id")
	} else {
		product.ID = 0
	}

	if val, ok := product.Data["gtin"]; ok {
		product.Gtin = val.(string)
		delete(product.Data, "gtin")
	} else {
		product.Gtin = ""
	}
	if val, ok := product.Data["serialNo"]; ok {
		product.SerialNumber = val.(float64)
		delete(product.Data, "serialNo")
	} else {
		product.SerialNumber = 0
	}
	if val, ok := product.Data["lot"]; ok {
		product.Lot = val.(string)
		delete(product.Data, "lot")
	} else {
		product.Lot = ""
	}
	if val, ok := product.Data["expirationDate"]; ok {
		product.ExpiryDate = val.(string)
		delete(product.Data, "expirationDate")
	} else {
		product.ExpiryDate = ""
	}
	if val, ok := product.Data["event"]; ok {
		product.Event = val.(string)
		delete(product.Data, "event")
	} else {
		product.Event = ""
	}
	if val, ok := product.Data["event_dt"]; ok {
		product.EventDate = val.(string)
		delete(product.Data, "event_dt")
	} else {
		product.EventDate = ""
	}
	if val, ok := product.Data["gln"]; ok {
		product.Gln = val.(string)
		delete(product.Data, "gln")
	} else {
		product.Gln = ""
	}
	if val, ok := product.Data["status"]; ok {
		product.Status = val.(string)
		delete(product.Data, "status")
	} else {
		product.Status = ""
	}
	if val, ok := product.Data["tradeItemDesc"]; ok {
		product.TradeItemDesc = val.(string)
		delete(product.Data, "tradeItemDesc")
	} else {
		product.TradeItemDesc = ""
	}
	if val, ok := product.Data["product"]; ok {
		product.Product = val.(string)
		delete(product.Data, "product")
	} else {
		product.Product = ""
	}
	if val, ok := product.Data["tradename"]; ok {
		product.TradeName = val.(string)
		delete(product.Data, "tradename")
	} else {
		product.TradeName = ""
	}
	if val, ok := product.Data["manufactureDate"]; ok {
		product.ManufactureDate = val.(string)
		delete(product.Data, "manufactureDate")
	} else {
		product.ManufactureDate = ""
	}
	if val, ok := product.Data["location"]; ok {
		product.Location = val.(string)
		delete(product.Data, "location")
	} else {
		product.Location = ""
	}
	if val, ok := product.Data["toGln"]; ok {
		product.ToGln = val.(string)
		delete(product.Data, "toGln")
	} else {
		product.ToGln = ""
	}
	if val, ok := product.Data["toLocation"]; ok {
		product.ToLocation = val.(string)
		delete(product.Data, "toLocation")
	} else {
		product.ToLocation = ""
	}
	if val, ok := product.Data["sender"]; ok {
		product.Sender = val.(string)
		delete(product.Data, "sender")
	} else {
		product.Sender = ""
	}
	if val, ok := product.Data["receiver"]; ok {
		product.Receiver = val.(string)
		delete(product.Data, "receiver")
	} else {
		product.Receiver = ""
	}
	logger.Info("product in end of getProductFromJSON", product)
	return product, nil

} // end of getProductFromJSON

// Gets the product key from the product
func getProductKey(product Product) string {

	// create the key from the 4 attributes
	//key := strings.ToLower(product.Gtin) + strings.ToLower(product.SerialNumber) + strings.ToLower(product.Lot) + product.ExpiryDate
	key := strings.ToLower(product.Gtin) + fmt.Sprintf("%f", product.SerialNumber) + strings.ToLower(product.Lot) + product.ExpiryDate
	return key

} // end of getProductKey

// Put it back in the original String will have to Marshall it twice
// to handle going from map back to original
func (product Product) toBytes() ([]byte, error) {

	var combinedBytes []byte
	bytesOuter, errOuter := json.Marshal(product)
	if errOuter != nil {
		return nil, errOuter
	}

	if len(product.Data) > 0 {
		bytesInner, errInner := json.Marshal(product.Data)
		if errInner != nil {
			return nil, errInner
		}
		byteSpaceSlice := []byte(" ")
		byteCommaSlice := []byte(",")
		bytesOuter[len(bytesOuter)-1] = byteSpaceSlice[0]
		bytesInner[0] = byteCommaSlice[0]
		combinedBytes = append(bytesOuter, bytesInner...)
	} else {
		combinedBytes = bytesOuter
	}

	return combinedBytes, nil

} // end of toBytes()

func main() {
	err := shim.Start(new(DataChainCode))
	if err != nil {
		logger.Errorf("Error starting Simple chaincode: %s", err)
	}
} // end of main()


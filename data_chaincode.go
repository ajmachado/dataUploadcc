package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"fmt"
		
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
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
	DocType      string                 `json:"docType"`
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


// ProductObjectType - defines the project object type
const ProductObjectType = "product-data"

// MaxProductJSONSizeAllowed - defines the max JSON size allowed for an input
const MaxProductJSONSizeAllowed = 2048

// MaxProductItems - defines the max product items that can be returned
const MaxProductItems = 100

// Init is called with the chaincode is instantiated or updated.
// It can be used to initialize data for the chaincode for real products or test
// For this we don't need to pre-populate anything
func (t *DataChainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Init: enter")
	defer fmt.Println("Init: exit")
	return shim.Success(nil)
} // end of init

// Invoke is called per transaction on the chaincode.
func (t *DataChainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Invoke: enter")
	defer fmt.Println("Invoke: exit")

	function, args := stub.GetFunctionAndParameters()
	txID := stub.GetTxID()

	fmt.Println("Invoke: Transaction ID: ", txID)
	fmt.Println("Invoke: function: ", function)
	fmt.Println("Invoke: args count: ", len(args))
	fmt.Println("Invoke: args found: ", args)

	if function == "createProduct" {
		return t.createProduct(stub, args)
	} else if function == "queryProductsByEvent" {
		return t.queryProductsByEvent(stub, args)
	}

	fmt.Println("Invoke: Invalid function = " + function)
	return shim.Error("Invoke: Invalid function = " + function)

} // end of invoke

// ============================================================================================================================
// Create Product - passes 2 arguments first is the key the second is JSON data mapping to the defined structure above
// ============================================================================================================================
// takes a single argument that is JSON of the product to create
func (t *DataChainCode) createProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("createProduct: enter")
	defer fmt.Println("createProduct: exit")

	if len(args) != 1 {
		errorString := "createProduct: Invalid number of args, must be exactly 1 argument containing JSON containing data"
		fmt.Println(errorString)
		return shim.Error(errorString)
	}

	var productInput = args[0]
	product, err := getProductFromJSON([]byte(productInput))
	if err != nil {
		fmt.Println("createProduct: Error with JSON format:", err)
		return shim.Error(err.Error())
	}

	key := getProductKey(product)
	bytes, err := product.toBytes()
	if err != nil {
		fmt.Println("createProduct: Error converting input product to bytes:", err)
		return shim.Error(err.Error())
	}
	// write it to the ledger
	fmt.Println("createProduct: call putState, key = ", key)
	err = stub.PutState(key, bytes)
	if err != nil {
		fmt.Println("createProduct: Error invoking on chaincode:", err)
		return shim.Error(err.Error())
	}
	fmt.Println("transaction id", stub.GetTxID())
	fmt.Println("createProduct: return successful write")
	return shim.Success(bytes)
	//return shim.Success([]byte(stub.GetTxID()))
} // end of createProduct



// ============================================================================================================================
// Get Product - get a product asset from ledger
// ============================================================================================================================
func getProduct(stub shim.ChaincodeStubInterface, key string) (Product, error) {

	fmt.Println("getProduct: enter")
	defer fmt.Println("getProduct: exit")

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
		fmt.Println("getProduct: Error with JSON format:", err)
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
	fmt.Println("product in getProductFromJSON", product)

	if err := json.Unmarshal([]byte(incoming), &product.Data); err != nil {
		return product, err
	}

	product.DocType = ProductObjectType
	if _, ok := product.Data["docType"]; ok {
		delete(product.Data, "docType")
	}
	
	var temp map[string]interface{}
	var lat float64
	var lon float64
	if val, ok := product.Data["loc_cd"]; ok {
		temp = val.(map[string]interface{})
		lat = temp["lat"].(float64)
		lon = temp["lon"].(float64)
		product.LocationInfo = LocationData{lat, lon}
		delete(product.Data, "loc_cd")
	} else {
		product.LocationInfo = LocationData{lat, lon}
	}
	//fmt.Println("Got Location Data", product.LocationInfo)
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
	fmt.Println("product in end of getProductFromJSON", product)
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

// ===== Example: Parameterized rich query =================================================
// queryProductsByGtin queries for products based on a passed in gtin.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (gtin).
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *DataChainCode) queryProductsByEvent(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println("queryProductsByEvent: enter")
	defer fmt.Println("queryProductsByEvent: exit")

	if len(args) < 1 {
		fmt.Println("queryProductsByGtin: Incorrect number of arguments. Expecting 1, that is an event")
		return shim.Error("queryProductsByEvent: Incorrect number of arguments. Expecting 1, that is an event")
	}

	event := args[0]
	fmt.Println("queryProductsByEvent: passed in event = ", event)

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"product-data\",\"event\":\"%s\"},\"use_index\":[\"_design/eventIndexDoc\",\"eventIndex\"]}", event)

	queryResults, err := getQueryResultForQueryString(stub, queryString, args)
	if err != nil {
		fmt.Println("queryProductsByEvent:, error getting results = ", err)
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)

} // end of queryProductsByGtin


// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string, args []string) ([]byte, error) {
	fmt.Println("getQueryResultForQueryString: enter")
	defer fmt.Println("getQueryResultForQueryString: exit")

	fmt.Println("getQueryResultForQueryString:  queryString:\n", queryString)

	// process additional query parameters such as offset and maxitems
	offset := 1
	maxitems := MaxProductItems

	if len(args) > 1 {
		var errString string
		fmt.Println("get arg1,  arg1 = ", args[1])
		arg1 := args[1]
		i, err := strconv.Atoi(arg1)
		if err != nil {
			errString = "getQueryResultForQueryString:, error passing parameter must be an integer, offset / arg1 = " + arg1 + ", err = " + err.Error()
			fmt.Println(errString)
			return nil, errors.New(errString)
		}
		offset = i
		if offset < 1 {
			errString = "getQueryResultForQueryString:, offset is 1 based and must be >= 1"
			fmt.Println(errString)
			return nil, errors.New(errString)
		}
		if len(args) > 2 {
			arg2 := args[2]
			i, err = strconv.Atoi(arg2)
			if err != nil {
				errString = "getQueryResultForQueryString:, error passing parameter must be an integer, maxitems / arg2 = " + arg2 + ", err = " + err.Error()
				fmt.Println(errString)
				return nil, errors.New(errString)
			}
			maxitems = i
			if maxitems > MaxProductItems {
				errString := "getQueryResultForQueryString: maxitems can not exceed " + strconv.Itoa(MaxProductItems)
				fmt.Println(errString)
				return nil, errors.New(errString)
			}
			if maxitems < 1 {
				errString = "getQueryResultForQueryString:, maxitems must be >= 1"
				fmt.Println(errString)
				return nil, errors.New(errString)
			}
		}

	}
	fmt.Println("getQueryResultForQueryString: offset = ", offset)
	fmt.Println("getQueryResultForQueryString: maxitems = ", maxitems)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("{\"products-data\": [")
	totalCount := 0
	itemsSkipped := 0
	itemsKept := 0
	bArrayMemberAlreadyWritten := false
	preLoopLen := offset - 1

	// execute the loop up to the offset
	for idx := 0; idx < preLoopLen && resultsIterator.HasNext(); idx++ {
		totalCount++
		itemsSkipped++
		_, err := resultsIterator.Next()
		if err != nil {
			return nil, errors.New(err.Error())
		}
	}

	for idx := 0; resultsIterator.HasNext(); idx++ {
		totalCount++
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, errors.New(err.Error())
		}

		if idx < maxitems {
			itemsKept++

			if bArrayMemberAlreadyWritten == true {
				buffer.WriteString(",")
			}
			buffer.WriteString("{\"Key\":")
			buffer.WriteString("\"")
			buffer.WriteString(queryResponse.Key)
			buffer.WriteString("\"")

			buffer.WriteString(", \"Record\":")
			// Record is a JSON object, so we write as-is
			buffer.WriteString(string(queryResponse.Value))
			buffer.WriteString("}")
			bArrayMemberAlreadyWritten = true
			skipString := "idx = " + strconv.Itoa(idx) + ", offset = " + strconv.Itoa(offset) + ", maxiteams = " + strconv.Itoa(maxitems) + "  -- added"
			fmt.Println(skipString)
			fmt.Println("Items added = ", idx)

		} else {
			itemsSkipped++
			skipString := "idx = " + strconv.Itoa(idx) + ", offset = " + strconv.Itoa(offset) + ", maxiteams = " + strconv.Itoa(maxitems) + ", items skipped = " + strconv.Itoa(itemsSkipped) + "  -- ignoring"
			fmt.Println(skipString)

		}
	}
	buffer.WriteString("], ")

	// Add the totalCount and items skipped to returned JSON
	buffer.WriteString("\"totalCount\":")
	buffer.WriteString("\"")
	buffer.WriteString(strconv.Itoa(totalCount))
	buffer.WriteString("\"")

	buffer.WriteString(",\"offset\":")
	buffer.WriteString("\"")
	buffer.WriteString(strconv.Itoa(offset))
	buffer.WriteString("\"")

	buffer.WriteString(",\"maxitems\":")
	buffer.WriteString("\"")
	buffer.WriteString(strconv.Itoa(maxitems))
	buffer.WriteString("\"")

	buffer.WriteString(",\"itemsSkipped\":")
	buffer.WriteString("\"")
	buffer.WriteString(strconv.Itoa(itemsSkipped))
	buffer.WriteString("\"")

	buffer.WriteString(",\"itemsKept\":")
	buffer.WriteString("\"")
	buffer.WriteString(strconv.Itoa(itemsKept))
	buffer.WriteString("\"")

	buffer.WriteString("}")

	fmt.Println("getQueryResultForQueryString: results found\n", buffer.String())

	return buffer.Bytes(), nil

} // end of getQueryResultForQueryString

func main() {
	err := shim.Start(new(DataChainCode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
} // end of main()


/*
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"strings"
	
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// SimpleChaincode example simple Chaincode implementation
type DataChainCode struct {
}

type Product struct {
	ID           int                    `json:"id"`
	Gtin         string                 `json:"gtin"`
	Lot          string                 `json:"lot"`
	SerialNumber string                 `json:"serialNo"`
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
	Sender       string                 `json:"sender"`
	Receiver     string                 `json:"receiver"`
	Data         map[string]interface{} `json:"-"` // Unknown fields should go here.
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(DataChainCode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *DataChainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *DataChainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "createProduct" { //create a new marble
		return t.createProduct(stub, args)
	} 

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

func (t *DataChainCode) createProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("createProduct: enter")
	
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
	//fmt.Println("createProduct: key = ", key)
	bytes, err := product.toBytes()
	if err != nil {
		fmt.Println("createProduct: Error converting input product to bytes:", err)
		return shim.Error(err.Error())
	}

	// write it to the ledger
	//logger.Debug("createProduct: call putState, key = ", key)
	err = stub.PutState(key, bytes)
	if err != nil {
		fmt.Println("createProduct: Error invoking on chaincode:", err)
		return shim.Error(err.Error())
	}

	fmt.Println("createProduct: return successful write")
	return shim.Success(bytes)

} // end of createProduct

// Since we have dynamic data we unmarshall into the Data field for everything
// then manually set the know types from the Data then remove them from Data so
// unmarshalling only occurs once
// NOTE: this method just unmashalls and does not validate, so if missing field it return empty string
func getProductFromJSON(incoming []byte) (Product, error) {
	var product Product
	//logger.Info("product in getProductFromJSON", product)
	if err := json.Unmarshal([]byte(incoming), &product.Data); err != nil {
		return product, err
	}
	
	if val, ok := product.Data["id"]; ok {
		product.ID = val.(int)
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
	if val, ok := product.Data["serialNumber"]; ok {
		product.SerialNumber = val.(string)
		delete(product.Data, "serialNumber")
	} else {
		product.SerialNumber = ""
	}
	if val, ok := product.Data["lot"]; ok {
		product.Lot = val.(string)
		delete(product.Data, "lot")
	} else {
		product.Lot = ""
	}
	if val, ok := product.Data["expiryDate"]; ok {
		product.ExpiryDate = val.(string)
		delete(product.Data, "expiryDate")
	} else {
		product.ExpiryDate = ""
	}
	if val, ok := product.Data["event"]; ok {
		product.Event = val.(string)
		delete(product.Data, "event")
	} else {
		product.Event = ""
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
	//logger.Info("product in end of getProductFromJSON", product)
	return product, nil

} // end of getProductFromJSON

// Gets the product key from the product
func getProductKey(product Product) string {

	// create the key from the 4 attributes
	key := strings.ToLower(product.Gtin) + strings.ToLower(product.SerialNumber) + strings.ToLower(product.Lot) + product.ExpiryDate
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
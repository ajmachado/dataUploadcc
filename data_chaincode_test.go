package main

import (
	"os"
	"testing"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/stretchr/testify/assert"
)

const (
	chainCodeID      = "dataUpload_1"
	
	mockDevJson = `{"id":1,"gtin":"08806555018611","lot":"M036191","serialNo":1936800,"expirationDate":"10/10/2026","event":"commission","gln":"0300060000037","status":"active","tradeItemDesc":"gardasil9 10 pack ","product":"gardasil9","tradename":"Gardasil 9","manufactureDate":"10/10/2019","location":"Wilson, NC","toGln":"0300060000037","toLocation":"Wilson, NC","loc_cd":{"lat":35.721268,"lon":-77.915543},"sender":"manufacturer","receiver":"manufacturer","event_dt":"2019-10-12T04:00:00.000Z"}`
	
)

// If TestMain exists then this will run all the tests,
// there can be a setup before all tests and a clean up after all tests have run
// NOTE: our tests will use the logger declared globally in the chaincode
func TestMain(m *testing.M) {
	//logger.SetLevel(shim.LogWarning)
	// logger.SetLevel(shim.LogDebug)

	//fmt.Println("TestMain: enter")
	//defer fmt.Println("TestMain: exit")
	fmt.Println("TestMain");

	exitCode := m.Run()

	os.Exit(exitCode)

} // end of TestMain

func Test_Init(t *testing.T) {
	simpleCC := new(DataChainCode)
	mockStub := shimtest.NewMockStub("mockstub", simpleCC)
	txId := "mockTxID"

	mockStub.MockTransactionStart(txId)
	response := simpleCC.Init(mockStub)
	mockStub.MockTransactionEnd(txId)
	if s := response.GetStatus(); s != 200 {
		fmt.Println("Init test failed")
		t.FailNow()
	}
}

func TestCreateProducts(t *testing.T) {
	fmt.Println("TestCreateProducts: enter")
	
	stub := shimtest.NewMockStub("mockStub", new(DataChainCode))

	if stub == nil {
		t.Fatalf("MockStub creation failed")
	}
	results := stub.MockInvoke("TestCreateproducts", [][]byte{[]byte("createProduct"), []byte(mockDevJson)})
	var returnCode = int(results.Status)
	assert.Equal(t, 200, returnCode, "Result : Success")

	
} // end of TestCreateProducts

/*func TestCRUD(t *testing.T) {
	fmt.Println("TestCRUD: enter")
	//defer fmt.Println("TestCRUD: exit")

	stub := shim.NewMockStub("mockStub", new(DataChainCode))

	if stub == nil {
		t.Fatalf("MockStub creation failed")
	}

	results := stub.MockInvoke("TestCRUD", [][]byte{[]byte("createProduct"), []byte(mockDevJson)})
	var returnCode = int(results.Status)
	assert.Equal(t, 200, returnCode, "Result : Success")
} // end of TestCRUD
*/
func TestQueryByEvent(t *testing.T) {
	fmt.Println("TestQueryByEvent: enter")
	defer fmt.Println("TestQueryByEvent: exit")

	stub := shimtest.NewMockStub("mockStub", new(DataChainCode))

	if stub == nil {
		t.Fatalf("TestQueryByEvent: MockStub creation failed")
	}

	results := stub.MockInvoke("TestQueryByEvent", [][]byte{[]byte("createProduct"), []byte(mockDevJson)})
	var returnCode = int(results.Status)
	assert.Equal(t, 200, returnCode, "Result : Success, mockProd1GoodCreateJSON")

	// this is not implemented so expect 500
	results = stub.MockInvoke("TestQueryByEvent", [][]byte{[]byte("queryProductsByEvent"), []byte("commission")})
	returnCode = int(results.Status)
	assert.Equal(t, 500, returnCode, "Result : Success, queryProductsByEvent")

	
} // end of TestQueryByEvent
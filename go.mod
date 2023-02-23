module example.com/dataUploadcc

go 1.12

require (
	github.com/hyperledger/fabric-chaincode-go v0.0.0-20190823162523-04390e015b85
	github.com/hyperledger/fabric-protos-go v0.0.0-20190821214336-621b908d5022
	github.com/stretchr/testify v1.4.0
	golang.org/x/text v0.3.8 // indirect
	google.golang.org/grpc v1.25.1 // indirect
	gopkg.in/yaml.v2 v2.2.5 // indirect
)

replace gotest.tools => github.com/gotestyourself/gotest.tools v2.1.0+incompatible

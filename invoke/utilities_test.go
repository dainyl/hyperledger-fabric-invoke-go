package invoke

import (
	"fmt"
	"testing"

	// "github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

func TestSuccess(t *testing.T) {
	status := int32(200)
	payload := "Hello World!"
	expected := pb.Response{
		Status:  status,
		Payload: []byte(payload),
		Message: "",
	}
	actual := Success(status, []byte(payload))

	deepEq(t, fmt.Sprintf("Success(%d, []byte(\"%s\"))", status, payload), expected, actual)
}

func TestError(t *testing.T) {
	status := int32(500)
	message := "Hello World!"
	expected := pb.Response{
		Status:  status,
		Payload: nil,
		Message: message,
	}
	actual := Error(status, message)

	deepEq(t, fmt.Sprintf("Error(%d, \"%s\")", status, message), expected, actual)
}

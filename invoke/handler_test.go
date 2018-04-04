package invoke

import (
	"fmt"
	"testing"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

func TestHandlerUse(t *testing.T) {
	router := NewRouter()
	key := "test"
	stub := shim.NewMockStub("test", new(testCC))
	// this is needed to set the transaction ID
	stub.MockTransactionStart("123")
	// create the transaction context, this is normally done in router.Invoke()
	router.context[stub.GetTxID()] = make(map[string]interface{})
	h := hIntAppender(router, key, 3)
	h = h.use(
		mwIntAppender(router, key, 1),
		mwIntAppender(router, key, 2),
	)

	h(stub, nil)

	expected := []int{1, 2, 3}
	actual := router.GetContext(stub)[key].([]int)
	deepEq(t, fmt.Sprintf("router.GetContext(stub)[%s]", key), expected, actual)
}

func mwIntAppender(router Router, contextKey string, val int) Middleware {
	return func(stub shim.ChaincodeStubInterface, args []string, next Handler) pb.Response {
		if _, ok := router.GetContext(stub)[contextKey]; !ok {
			router.GetContext(stub)[contextKey] = make([]int, 0)
		}
		arr := router.GetContext(stub)[contextKey].([]int)
		arr = append(arr, val)
		router.GetContext(stub)[contextKey] = arr
		return next(stub, args)
	}
}

func hIntAppender(router Router, contextKey string, val int) Handler {
	return func(stub shim.ChaincodeStubInterface, args []string) pb.Response {
		if _, ok := router.GetContext(stub)[contextKey]; !ok {
			router.GetContext(stub)[contextKey] = make([]int, 0)
		}
		arr := router.GetContext(stub)[contextKey].([]int)
		arr = append(arr, val)
		router.GetContext(stub)[contextKey] = arr
		return Success(200, nil)
	}
}

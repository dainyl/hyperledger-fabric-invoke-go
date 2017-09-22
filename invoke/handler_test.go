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
	router.Context[key] = make([]int, 0)
	h := hIntAppender(router, key, 3)
	h = h.use(
		mwIntAppender(router, key, 1),
		mwIntAppender(router, key, 2),
	)

	h(nil, nil)

	expected := []int{1, 2, 3}
	actual := router.Context[key].([]int)
	deepEq(t, fmt.Sprintf("router.Context[%s]", key), expected, actual)
}

func mwIntAppender(router Router, contextKey string, val int) Middleware {
	return func(stub shim.ChaincodeStubInterface, args []string, next Handler) pb.Response {
		arr := router.Context[contextKey].([]int)
		arr = append(arr, val)
		router.Context[contextKey] = arr
		return next(stub, args)
	}
}

func hIntAppender(router Router, contextKey string, val int) Handler {
	return func(stub shim.ChaincodeStubInterface, args []string) pb.Response {
		arr := router.Context[contextKey].([]int)
		arr = append(arr, val)
		router.Context[contextKey] = arr
		return Success(200, nil)
	}
}

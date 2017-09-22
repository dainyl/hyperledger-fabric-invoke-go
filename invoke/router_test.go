package invoke

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

func TestNewRouter(t *testing.T) {
	router := NewRouter()

	notNil(t, "router.Context", router.Context)
	notNil(t, "router.invokeMap", router.invokeMap)
	notNil(t, "router.middlewareChain", router.middlewareChain)
	eq(t, "len(router.Context)", 0, len(router.Context))
	eq(t, "len(router.invokeMap)", 0, len(router.invokeMap))
	eq(t, "len(router.middlewareChain)", 0, len(router.middlewareChain))
}

type testCC struct{}

func (t *testCC) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return Success(200, nil)
}

func (t *testCC) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	return Success(200, nil)
}

func TestRouterUse(t *testing.T) {
	router := NewRouter()
	router.Use(mwIntAppender(router, "test", 1), mwIntAppender(router, "test", 2))

	eq(t, "len(router.middlewareChain)", 2, len(router.middlewareChain))
}

func TestRegisterHandler(t *testing.T) {
	router := NewRouter()
	key := "test"
	endpoint := "endpoint"
	h := router.RegisterHandler(
		endpoint,
		hIntAppender(router, key, 3),
		mwIntAppender(router, key, 1),
		mwIntAppender(router, key, 2),
	)

	eq(t, "len(router.invokeMap)", 1, len(router.invokeMap))
	notNil(t, fmt.Sprintf("router.invokeMap[%s]", endpoint), router.invokeMap[endpoint])
	notNil(t, "h", h)
}

var invokeTests = []struct {
	endpoint    string
	expectedRsp pb.Response
}{
	{"nothing", Error(400, "invalid invoke function \"nothing\"")},
	{"endpoint", Success(200, nil)},
}

func TestInvoke(t *testing.T) {
	router := NewRouter()
	key := "test"
	router.Context[key] = make([]int, 0)
	endpoint := "endpoint"
	router.Use(mwIntAppender(router, key, 1))
	router.RegisterHandler(
		endpoint,
		hIntAppender(router, key, 4),
		mwIntAppender(router, key, 2),
		mwIntAppender(router, key, 3),
	)

	for _, v := range invokeTests {
		stub := shim.NewMockStub("test", new(testCC))
		// this is only needed to set the args, which is the only part the router needs
		stub.MockInvoke("text", [][]byte{[]byte(v.endpoint)})
		rsp := router.Invoke(stub)
		deepEq(t, "invoke response", v.expectedRsp, rsp)
	}
}

func notNil(t *testing.T, name string, val interface{}) {
	if val == nil || reflect.ValueOf(val).IsNil() {
		t.Errorf("%s was unexpectedly nil", name)
	}
}

func eq(t *testing.T, testName string, expected interface{}, actual interface{}) {
	if expected != actual {
		t.Errorf("%s: expected %#v but got %#v", testName, expected, actual)
	}
}

func deepEq(t *testing.T, testName string, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("%s: expected %#v but got %#v", testName, expected, actual)
	}
}

# Hyperledger Fabric Invoke (golang)

Invoke is a convenience library for authoring Hyperledger Fabric chaincode.

The key feature of invoke is the `Router`, which simplifies managing chaincode endpoints and allows the mounting of middleware on handlers.

Invoke also contains a number of common utility functions and middleware functions used in chaincode.

## Basic Router Setup

```go
import "github.ibm.com/bhaesler/hyperledger-fabric-invoke-go/invoke"

type MyChaincode struct {}

// declare the router as a global variable so it can be accessed by the Invoke function
var router invoke.Router

func (m *MyChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
    // manage invocations using the router
    return router.Invoke(stub)
}

func myHandler(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    // handler logic

    return invoke.Success(http.StatusOK, nil)
}

// it's useful to put this in a helper function for writing unit tests on your chaincode
func initRouter() {
    // create the router
    router = invoke.NewRouter()

    // add endpoints to the router
    router.RegisterHandler(
        "myEndpoint",   // the function name that is called by the external client
        myHandler,      // the handler to be run for this endpoint
    )
}

func main() {
    // setup the router
    initRouter()

    if err := shim.Start(new(MyChaincode)); err != nil {
        fmt.Printf("error starting MyChaincode: %s", err)
    }
}
```

## Invoke Middleware

Middleware is intended to reduce the amount of boilerplate code required in handler implementations, reducing handler complexity and increasing readability and maintainability.

### Simple Middleware Function

```go
// argLogger logs all of the arguments passed to an invoke call
func argLogger(stub shim.ChaincodeStubInterface, args []string, next invoke.Handler) pb.Response {
    functionName := stub.GetFunctionAndParameters()
    // log the arguments passed to this endpoint
    fmt.Printf("arguments passed to %s endpoint: %v\n", functionName, args)

    // call the next handler in the chain
    next(stub, args)
}

func initRouter() {
    router = invoke.NewRouter()

    // register global middleware (runs for every endpoint)
    router.Use(argLogger)

    // or, register specific middleware for each endpoint
    router.registerHandler(
        "myEndpoint",   // the function name that is called by the external client
        myHandler,      // the handler to be run for this endpoint
        argLogger,      // middleware function(s) attached to this endpoint
    ) 
}

func main() {
    initRouter()

    if err := shim.Start(new(MyChaincode)); err != nil {
        fmt.Printf("error starting MyChaincode: %s", err)
    }
}
```

### Complex Middleware Function

Sometimes data created in a middleware function needs to passed through to subsequent middleware or the handler. This is achieved by using the router's `Context`, a `map[string]interface{}`

```go
// timestampParser generates a middleware function that will attempt to parse
// the specified argument as a timestamp in the given format, and store the
// result in the context under the key provided.
func timestampParser(argIndex int, timeFormat string, contextKey string) Middleware {
    return func(stub shim.ChaincodeStubInterface, args []string, next invoke.Handler) pb.Response {
        // check index is valid
        if argIndex >= len(args) {
            // return an error here, short-circuiting the execution of other handlers on the endpoint
            err := fmt.Sprintf("argIndex %d was greater than length of args", argIndex)
            fmt.Println(err)
            return invoke.Error(http.StatusInternalServerError, fmt.Sprintf("error parsing time: %s", err))
        }

        // parse timestamp
        var ts time.Time
        var err error
        if ts, err = time.Parse(timeFormat, args[argIndex]); err != nil {
            fmt.Println(err)
            return invoke.Error(http.StatusBadRequest, fmt.Sprintf("error parsing time string: %s", err.Error()))
        }

        // write timestamp to context
        router.Context[contextKey] = ts

        // call next handler
        return next(stub, args)
    }
}

func myHandler(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    // retrieve the timestamp from the Context. A type assertion is required
    // to convert from interface{} to time.Time
    ts := router.Context["timestamp"].(time.Time)

    // use timestamp in handler logic

    return invoke.Success(http.StatusOK, nil)
}

func initRouter() {
    router = invoke.NewRouter()
    
    router.RegisterHandler(
        "myEndpoint",
        myHandler,
        timestampParser(1, time.RFC3339, "timestamp")
    )
}

func main() {
    initRouter()

    if err := shim.Start(new(MyChaincode)); err != nil {
        fmt.Printf("error starting MyChaincode: %s", err)
    }
}
```

### Provided Middleware Functions

`ArgCounter` - Validates number of arguments passed to a function  
`JSONParser` - Parses an argument as json and stores the result in the context  
`TimestampParser` - Parses an argument as time.Time and stores the result in the context  
`TransactionTimestamp` - Extracts the transaction timestamp as time.Time and stores the result in the context. See the `GetTxTimestamp` method in [`ChaincodeStubInterface` godoc](https://godoc.org/github.com/hyperledger/fabric/core/chaincode/shim#ChaincodeStubInterface)

## Utility Functions

### `invoke.Success` and `invoke.Error`

All `Init` and `Invoke` functions return a `pb.Response`
```go
type Response struct {
	// A status code that should follow the HTTP status codes.
	Status int32
	// A message associated with the response code.
	Message string
	// A payload that can be used to include metadata with this response.
	Payload []byte
}
```
`github.com/hyperledger/fabric/core/chaincode/shim` provides convenience functions `Success` and `Error`:
```go
func Success(payload []byte) pb.Response {...}
func Error(msg string) pb.Response {...}
```
These functions default to status codes of 200 and 500 respectively, however these codes often don't correctly reflect the success or error.
Invoke offers two similar functions which allow specifying a response code:
```go
func Success(status int32, payload []byte) pb.Response {...}
func Error(status int32, msg string) pb.Response {...}
```
It is recommended to import `net/http` and use the constant status codes exported by that library.

### `invoke.PutJSON` and `invoke.GetJSON`

Records on the ledger in Hyperledger Fabric are often stored in json format, especially when using CouchDB as the underlying ledger database. `PutJSON` and `GetJSON` combine json marhsal/unmarshal and accessing the ledger. `PutJSON` returns the json encoded byte array for use in `invoke.Success` payloads.

### `invoke.GetQueryResultForQueryString`

 The main advantage of using CouchDB as the underlying peer database is the ability to perform complex queries. `GetQueryResultForQueryString` takes a CouchDB query string and returns a json array of `{ key, value }` pairs, encoded as a byte array for use in `invoke.Success` payloads.

 ### `invoke.GetCreatorCert`

 Extracts and parses the x509 certificate of the creator of the transaction. This is useful for implementing access control on chaincode functions

 ### `invoke.GetCreatorCommonName`

 Extracts the common name field from the x509 certificate of the creator of the transaction. This is useful for implementing access control on chaincode functions

 ## Logging

 Invoke uses a `shim.ChaincodeLogger` for internal logging within functions. It can be disabled by setting the logging level (via code or environment variables), or it can be replaced by setting `invoke.Logger` to another `shim.ChaincodeLogger`
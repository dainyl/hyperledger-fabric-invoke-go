package invoke

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// ArgCounter takes the names of expected arguments to a handler, and returns
// a middleware function that checks for that number of arguments.
func ArgCounter(expected ...string) Middleware {
	return func(stub shim.ChaincodeStubInterface, args []string, next Handler) pb.Response {
		// if there is the wrong number of args
		if len(args) != len(expected) {
			// make a buffer for efficiency
			var res bytes.Buffer
			// write the start of the error
			res.WriteString(fmt.Sprintf("incorrect number of arguments, expected %d", len(expected)))
			// if there were expected args
			if len(expected) > 0 {
				res.WriteString(": ")
				// write each argument
				for i, arg := range expected {
					res.WriteString(arg)
					// add a comma if this is not the last
					if i < len(expected)-1 {
						res.WriteString(", ")
					}
				}
			}

			err := res.String()

			// log and return the error
			Logger.Error(err)
			return Error(http.StatusBadRequest, err)
		}

		// call next handler
		return next(stub, args)
	}
}

// JSONParser creates a middleware that will attempt to parse the string in the
// specified argument position as json and store the result in the context as a pointer.
func JSONParser(router Router, argIndex int, contextKey string, valueType reflect.Type) Middleware {
	return func(stub shim.ChaincodeStubInterface, args []string, next Handler) pb.Response {
		// check index is valid
		if argIndex >= len(args) {
			err := fmt.Sprintf("argIndex %d was greater than length of args", argIndex)
			Logger.Errorf(err)
			return Error(http.StatusInternalServerError, fmt.Sprintf("error unmarshalling json: %s", err))
		}

		// get payload
		b := []byte(args[argIndex])

		// create an object to store the value
		jsonValue := reflect.New(valueType).Interface()

		// try to unmarshal
		if err := json.Unmarshal(b, jsonValue); err != nil {
			Logger.Error(err)
			return Error(http.StatusBadRequest, fmt.Sprintf("error unmarshalling json: %s", err.Error()))
		}

		// store result in context
		router.Context[contextKey] = jsonValue

		// call next handler
		return next(stub, args)
	}
}

// TimestampParser creates a middleware that will attempt to parse the string in
// the specified argument position as a given time format and store the result
// in the context.
func TimestampParser(router Router, argIndex int, timeFormat string, contextKey string) Middleware {
	return func(stub shim.ChaincodeStubInterface, args []string, next Handler) pb.Response {
		// check index is valid
		if argIndex >= len(args) {
			err := fmt.Sprintf("argIndex %d was greater than length of args", argIndex)
			Logger.Error(err)
			return Error(http.StatusInternalServerError, fmt.Sprintf("error parsing time: %s", err))
		}

		// parse timestamp
		var ts time.Time
		var err error
		if ts, err = time.Parse(timeFormat, args[argIndex]); err != nil {
			Logger.Error(err)
			return Error(http.StatusBadRequest, fmt.Sprintf("error parsing time string: %s", err.Error()))
		}

		// write timestamp to context
		router.Context[contextKey] = ts

		// call next handler
		return next(stub, args)
	}
}

// TransactionTimestamp creates a middleware that will extract the transaction
// timestamp and store it in the context under the given key as a time.Time
func TransactionTimestamp(router Router, contextKey string) Middleware {
	return func(stub shim.ChaincodeStubInterface, args []string, next Handler) pb.Response {
		// get the timestamp from the transaction metadata
		ts, err := stub.GetTxTimestamp()
		if err != nil {
			err = fmt.Errorf("error getting transaction timestamp: %s", err.Error())
			Logger.Error(err)
			return Error(http.StatusInternalServerError, err.Error())
		}

		// store the timestamp in the context under the given key
		router.Context[contextKey] = time.Unix(ts.GetSeconds(), int64(ts.GetNanos()))

		// call the next handler
		return next(stub, args)
	}
}

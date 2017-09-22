/*
Copyright IBM Corp. 2017 All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
		 http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package invoke contains convenience tools for writing and managing invoke
// functions in hyperledger-fabric chaincode. Its key feature is the Router,
// which allows the registration of invoke handlers and middleware functions.
package invoke

import (
	"fmt"
	"net/http"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var Logger = shim.NewLogger("invoke")

// Router objects manage handlers and middleware for invoke calls.
type Router struct {
	Context         map[string]interface{}
	invokeMap       map[string]Handler
	middlewareChain []Middleware
}

// NewRouter returns a new router with no handlers or middleware.
func NewRouter() Router {
	return Router{
		Context:         make(map[string]interface{}),
		invokeMap:       make(map[string]Handler),
		middlewareChain: make([]Middleware, 0),
	}
}

// Use adds the given middleware to a list of middleware used on all invoke calls.
func (r *Router) Use(mws ...Middleware) {
	r.middlewareChain = append(r.middlewareChain, mws...)
}

// RegisterHandler adds a new handler to the router, wrapped in any specific middleware provided.
func (r *Router) RegisterHandler(functionName string, h Handler, mws ...Middleware) Handler {
	// attach the middleware
	r.invokeMap[functionName] = h.use(mws...)
	// return the handler with middleware attached
	return r.invokeMap[functionName]
}

// Invoke calls the appropriate handler for this invoke call.
func (r *Router) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	// get arguments to invoke
	function, args := stub.GetFunctionAndParameters()

	// get invoke handler from map
	var fn Handler
	var ok bool
	if fn, ok = r.invokeMap[function]; !ok {
		// if the function was not in the invoke map, return an error
		err := fmt.Errorf("invalid invoke function \"%s\"", function)
		Logger.Error(err.Error())
		return Error(http.StatusBadRequest, err.Error())
	}

	// attach the global middleware chain
	fn = fn.use(r.middlewareChain...)

	// execute invoke function
	result := fn(stub, args)

	// return result
	return result
}

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

package invoke

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// Handler is a function that handles an invoke call.
type Handler func(shim.ChaincodeStubInterface, []string) pb.Response

// Middleware is a function that wraps a handler to perform a specific task,
// and then calls the handler and returns its result
type Middleware func(shim.ChaincodeStubInterface, []string, Handler) pb.Response

// use wraps the handler in each of the provided middleware functions, and
// returns a handler that will execute the middleware in the order provided,
// followed by the handler.
func (h Handler) use(m ...Middleware) Handler {
	// for each middleware, starting with the last listed
	for i := len(m) - 1; i >= 0; i-- {
		// make copies of the middleware and handler for use in the closure
		tmpMW := m[i]
		tmpH := h
		// wrap the handler in the middleware
		h = func(stub shim.ChaincodeStubInterface, args []string) pb.Response {
			return tmpMW(stub, args, tmpH)
		}
	}

	// return the new handler
	return h
}

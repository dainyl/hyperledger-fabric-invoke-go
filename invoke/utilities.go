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
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	mspprotos "github.com/hyperledger/fabric/protos/msp"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// Success is a helper function emulating the behaviour of ChaincodeStubInterface.Success,
// but with a custom status parameter instead of the default 200
func Success(status int32, payload []byte) pb.Response {
	return pb.Response{
		Status:  status,
		Payload: payload,
	}
}

// Error is a helper function emulating the behaviour of ChaincodeStubInterface.Error, but
// with a custom status paramater instead of the default 500
func Error(status int32, message string) pb.Response {
	return pb.Response{
		Status:  status,
		Message: message,
	}
}

// PutJSON marshals the given object to json and writes it to the ledger.
func PutJSON(stub shim.ChaincodeStubInterface, key string, value interface{}) ([]byte, error) {
	// serialise the record as json
	var b []byte
	var err error
	if b, err = json.Marshal(value); err != nil {
		Logger.Error(err.Error())
		return nil, err
	}

	// write the record to the chain
	if err = stub.PutState(key, b); err != nil {
		Logger.Error(err.Error())
		return nil, err
	}

	return b, nil
}

// GetJSON retrieves a value from the ledger and attempts to unmarshal it as json.
func GetJSON(stub shim.ChaincodeStubInterface, key string, valuePtr interface{}) error {
	var b []byte
	var err error
	if b, err = stub.GetState(key); err != nil {
		Logger.Errorf("error getting state of %s from ledger: %s", key, err.Error())
		return err
	}

	if err = json.Unmarshal(b, valuePtr); err != nil {
		Logger.Errorf("error deserialising value of %s as json: %s", b, err.Error())
		return err
	}

	return nil
}

// GetQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
func GetQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	Logger.Debugf("getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	Logger.Debugf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

// GetCreatorCert gets the certificate of the transactor who initiated this transaction.
func GetCreatorCert(stub shim.ChaincodeStubInterface) (*x509.Certificate, error) {
	// get the creator identity from the stub
	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return nil, err
	}

	// deserialise the identity from the protobuf encoding
	var id mspprotos.SerializedIdentity
	if err = proto.Unmarshal(creatorBytes, &id); err != nil {
		return nil, err
	}

	Logger.Debugf("Creator Identity: %#v", id)

	// decode the contents of the .pem file stored in the identity
	block, _ := pem.Decode(id.IdBytes)

	Logger.Debugf("Pem: %#v", block)

	// parse the contents of the .pem as an x509 cert and return the result
	return x509.ParseCertificate(block.Bytes)
}

// GetCreatorCommonName gets the common name from the certificate of the transactor
// who initiated this transaction
func GetCreatorCommonName(stub shim.ChaincodeStubInterface) (string, error) {
	cert, err := GetCreatorCert(stub)
	if err != nil {
		return "", err
	}

	return cert.Subject.CommonName, nil
}
package main

import (
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
	b64 "encoding/base64"
	"github.com/hyperledger/fabric/core/util"
)

const KVS_HANLDER_KEY = "KVS_HANDLER_KEY"

type UserManagement struct {
}

type UserKey struct {
	BIC    string  `json:"bic"`
	Login  string  `json:"login"`
}

type PermissionAccountKey struct {
	Type         string  `json:"type"`
        Holder       string  `json:"holder"`
	Owner        string  `json:"owner"`
	Currency     string  `json:"currency"`
	AccountType  string  `json:"accountType"`
}

type Permission struct {
	Key      PermissionAccountKey  `json:"accountKey"`
	Access   string  `json:"access"`
}

type UserDetails struct {
	Password     string  `json:"password"`
	Permissions  []Permission  `json:"permissions"`
}

type Response struct {
	Status     string  `json:"status"`
        Message    string  `json:"message"`
        AuthToken  string  `json:"authToken"`
}

func (t *UserManagement) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. KVS chaincode id is expected");
	}
	stub.PutState(KVS_HANLDER_KEY, []byte(args[0]))

	return nil, nil
}

func (t *UserManagement) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	switch function {
	case "create":
		if len(args) != 3 {
			return nil, errors.New("Incorrect number of arguments. 3 parameters are expected: bic,login,permissions");
		}

		userKey := &UserKey{ BIC: args[0], Login: args[1] }
		userDetails, _ := b64.StdEncoding.DecodeString(string(args[2]))

		state, _ := stub.GetState(KVS_HANLDER_KEY)
		mapId := string(state);
		jsonUserKey, _ := json.Marshal(userKey)
		invokeArgs := util.ToChaincodeArgs("put", string(jsonUserKey), string(userDetails))

		stub.InvokeChaincode(mapId, invokeArgs)
		return nil, nil
	default:
		return nil, errors.New("Unsupported operation")
	}
}

func (t *UserManagement) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	switch function {
	case "login":
		if len(args) != 3 {
			return nil, errors.New("Incorrect number of arguments. 3 parameters are expected: bic,login,permissions");
		}

		userKey := &UserKey{ BIC: args[0], Login: args[1] }
		state, _ := stub.GetState(KVS_HANLDER_KEY)
		mapId := string(state);
		jsonUserKey, _ := json.Marshal(userKey)
		queryArgs := util.ToChaincodeArgs("function", string(jsonUserKey))

		queryResult, _ := stub.QueryChaincode(mapId, queryArgs)
		var userDetails UserDetails
		if err := json.Unmarshal(queryResult, &userDetails); err != nil {
			panic(err)
		}

		if (userDetails.Password != string(args[2])) {
			return nil, errors.New("BIC code or login or password you entered is incorrect.");
		}
		token := b64.StdEncoding.EncodeToString(jsonUserKey)

		authToken := &Response {
			Status: "OK",
			AuthToken: token,
		}
		return json.Marshal(authToken)
	case "userDetails":
		if len(args) != 1 {
			return nil, errors.New("Incorrect number of arguments. Expecting 1 argument: authToken");
		}

		jsonUserKey, _ := b64.StdEncoding.DecodeString(string(args[0]))
		state, _ := stub.GetState(KVS_HANLDER_KEY)
		mapId := string(state);
		queryArgs := util.ToChaincodeArgs("function", string(jsonUserKey))

		queryResult, _ := stub.QueryChaincode(mapId, queryArgs)
		var userDetails UserDetails
		if err := json.Unmarshal(queryResult, &userDetails); err != nil {
			panic(err)
		}
		userDetails.Password = "removed"

		return json.Marshal(userDetails)
	default:
		return nil, errors.New("Unsupported operation")
	}
}

func main() {
	err := shim.Start(new(UserManagement))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
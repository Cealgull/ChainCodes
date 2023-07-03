package chaincode_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"userprofile/chaincode"
	"userprofile/chaincode/mocks"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"

	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	// _ "github.com/maxbrunsfeld/counterfeiter/v6"
	"github.com/stretchr/testify/require"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o mocks/transaction.go -fake-name TransactionContext . transactionContext
type transactionContext interface {
	contractapi.TransactionContextInterface
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o mocks/chaincodestub.go -fake-name ChaincodeStub . chaincodeStub
type chaincodeStub interface {
	shim.ChaincodeStubInterface
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o mocks/statequeryiterator.go -fake-name StateQueryIterator . stateQueryIterator
type stateQueryIterator interface {
	shim.StateQueryIteratorInterface
}

func TestInitLedger(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	userprofile := chaincode.SmartContract{}
	err := userprofile.InitLedger(transactionContext)
	require.NoError(t, err)

	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = userprofile.InitLedger(transactionContext)
	require.EqualError(t, err, "failed to put to world state. failed inserting key")
}

func TestCreateUser(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	userprofile := chaincode.SmartContract{}
	err := userprofile.CreateUser(transactionContext, "", "", "", "")
	require.NoError(t, err)

	chaincodeStub.GetStateReturns([]byte{}, nil)
	err = userprofile.CreateUser(transactionContext, "user1", "", "", "")
	require.EqualError(t, err, "the user user1 already exists")

	chaincodeStub.GetStateReturns(nil, fmt.Errorf("unable to retrieve asset"))
	err = userprofile.CreateUser(transactionContext, "user1", "", "", "")
	require.EqualError(t, err, "failed to read from world state: unable to retrieve asset")
}

func TestReadUser(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	expectedAsset := &chaincode.UserProfile{IdentityId: "user1"}
	bytes, err := json.Marshal(expectedAsset)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	userprofile := chaincode.SmartContract{}
	asset, err := userprofile.ReadUser(transactionContext, "")
	require.NoError(t, err)
	require.Equal(t, expectedAsset, asset)

	chaincodeStub.GetStateReturns(nil, fmt.Errorf("unable to retrieve asset"))
	_, err = userprofile.ReadUser(transactionContext, "")
	require.EqualError(t, err, "failed to read from world state: unable to retrieve asset")

	chaincodeStub.GetStateReturns(nil, nil)
	asset, err = userprofile.ReadUser(transactionContext, "user1")
	require.EqualError(t, err, "the user user1 does not exist")
	require.Nil(t, asset)
}

func TestUpdateUser(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	expectedAsset := &chaincode.UserProfile{IdentityId: "user1"}
	bytes, err := json.Marshal(expectedAsset)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	userprofile := chaincode.SmartContract{}
	err = userprofile.UpdateUser(transactionContext, "", "", "", "")
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(nil, nil)
	err = userprofile.UpdateUser(transactionContext, "user1", "", "", "")
	require.EqualError(t, err, "the user user1 does not exist")

	chaincodeStub.GetStateReturns(nil, fmt.Errorf("unable to retrieve asset"))
	err = userprofile.UpdateUser(transactionContext, "user1", "", "", "")
	require.EqualError(t, err, "failed to read from world state: unable to retrieve asset")
}

func TestAssignRole(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	expectedAsset := &chaincode.UserProfile{IdentityId: "user1"}
	bytes, err := json.Marshal(expectedAsset)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	userprofile := chaincode.SmartContract{}
	err = userprofile.AssignRole(transactionContext, "user1", "Admin")
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(nil, nil)
	err = userprofile.AssignRole(transactionContext, "user1", "")
	require.EqualError(t, err, "the user user1 does not exist")

	chaincodeStub.GetStateReturns(nil, fmt.Errorf("unable to retrieve asset"))
	err = userprofile.AssignRole(transactionContext, "", "")
	require.EqualError(t, err, "failed to read from world state: unable to retrieve asset")
}

func TestRemoveRole(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	expectedUser := &chaincode.UserProfile{IdentityId: "user1", Roles: []string{"Admin"}}
	bytes, err := json.Marshal(expectedUser)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	userprofile := chaincode.SmartContract{}
	err = userprofile.RemoveRole(transactionContext, "user1", "Admin")
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(nil, nil)
	err = userprofile.RemoveRole(transactionContext, "user1", "")
	require.EqualError(t, err, "the user user1 does not exist")

	chaincodeStub.GetStateReturns(nil, fmt.Errorf("unable to retrieve asset"))
	err = userprofile.RemoveRole(transactionContext, "", "")
	require.EqualError(t, err, "failed to read from world state: unable to retrieve asset")
}

func TestAssignBadge(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	expectedAsset := &chaincode.UserProfile{IdentityId: "user1"}
	bytes, err := json.Marshal(expectedAsset)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	userprofile := chaincode.SmartContract{}
	err = userprofile.AssignBadge(transactionContext, "user1", "Admin")
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(nil, nil)
	err = userprofile.AssignBadge(transactionContext, "user1", "")
	require.EqualError(t, err, "the user user1 does not exist")

	chaincodeStub.GetStateReturns(nil, fmt.Errorf("unable to retrieve asset"))
	err = userprofile.AssignBadge(transactionContext, "", "")
	require.EqualError(t, err, "failed to read from world state: unable to retrieve asset")
}

func TestRemoveBadge(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	expectedUser := &chaincode.UserProfile{IdentityId: "user1", Badge: []string{"Admin"}}
	bytes, err := json.Marshal(expectedUser)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	userprofile := chaincode.SmartContract{}
	err = userprofile.RemoveBadge(transactionContext, "user1", "Admin")
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(nil, nil)
	err = userprofile.RemoveBadge(transactionContext, "user1", "")
	require.EqualError(t, err, "the user user1 does not exist")

	chaincodeStub.GetStateReturns(nil, fmt.Errorf("unable to retrieve asset"))
	err = userprofile.RemoveBadge(transactionContext, "", "")
	require.EqualError(t, err, "failed to read from world state: unable to retrieve asset")
}

func TestGetAllUsers(t *testing.T) {
	asset := &chaincode.UserProfile{IdentityId: "user1"}
	bytes, err := json.Marshal(asset)
	require.NoError(t, err)

	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, false)
	iterator.NextReturns(&queryresult.KV{Value: bytes}, nil)

	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	chaincodeStub.GetStateByRangeReturns(iterator, nil)
	userprofile := &chaincode.SmartContract{}
	assets, err := userprofile.GetAllUsers(transactionContext)
	require.NoError(t, err)
	require.Equal(t, []*chaincode.UserProfile{asset}, assets)

	iterator.HasNextReturns(true)
	iterator.NextReturns(nil, fmt.Errorf("failed retrieving next item"))
	assets, err = userprofile.GetAllUsers(transactionContext)
	require.EqualError(t, err, "failed retrieving next item")
	require.Nil(t, assets)

	chaincodeStub.GetStateByRangeReturns(nil, fmt.Errorf("failed retrieving all assets"))
	assets, err = userprofile.GetAllUsers(transactionContext)
	require.EqualError(t, err, "failed retrieving all assets")
	require.Nil(t, assets)
}

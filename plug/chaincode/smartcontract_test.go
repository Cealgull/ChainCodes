package chaincode_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"plug/chaincode"
	"plug/chaincode/mocks"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"

	// _ "github.com/maxbrunsfeld/counterfeiter/v6"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
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

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o mocks/clientIdentity.go -fake-name ClientIdentity . clientIdentity
type clientIdentity interface {
	cid.ClientIdentity
}

const myOrg1Msp = "Org1Testmsp"
const myOrg1Clientid = "myOrg1Userid"
const myOrg1PrivCollection = "Org1TestmspPrivateCollection"
const myOrg2Msp = "Org2Testmsp"
const myOrg2Clientid = "myOrg2Userid"
const myOrg2PrivCollection = "Org2TestmspPrivateCollection"

var sampleTag = &chaincode.Tag{
	Name:        "tag1",
	Creator:     "1",
	Description: "tag1",
}

var sampleInput1, _ = json.Marshal(sampleTag)

func TestCreateTag(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	tag := chaincode.SmartContract{}

	err := tag.CreateTag(transactionContext, string(sampleInput1))
	require.NoError(t, err)

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	err = tag.CreateTag(transactionContext, string(sampleInput1))
	require.EqualError(t, err, "failed to read from world state: failure")

	expectedTag := &chaincode.Tag{Name: "1"}
	bytes, err := json.Marshal(expectedTag)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	err = tag.CreateTag(transactionContext, string(sampleInput1))
	require.EqualError(t, err, "the tag tag1 already exists")

	chaincodeStub.GetStateReturns(nil, nil)
	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = tag.CreateTag(transactionContext, string(sampleInput1))
	require.EqualError(t, err, "failed to put to world state: failed inserting key")
}

func TestReadTag(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	tag := chaincode.SmartContract{}

	tmpTag := &chaincode.Tag{Name: "1"}
	bytes, _ := json.Marshal(tmpTag)
	chaincodeStub.GetStateReturns(bytes, nil)
	_, err := tag.ReadTag(transactionContext, "1")
	require.NoError(t, err)

	err = tag.CreateTag(transactionContext, "sad")
	require.EqualError(t, err, "invalid character 's' looking for beginning of value")

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	_, err = tag.ReadTag(transactionContext, "1")
	require.EqualError(t, err, "failed to read from world state: failure")

	chaincodeStub.GetStateReturns(nil, nil)
	_, err = tag.ReadTag(transactionContext, "1")
	require.EqualError(t, err, "the tag 1 does not exist")
}

func TestUpdateTag(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	tag := chaincode.SmartContract{}

	err := tag.UpdateTag(transactionContext, string(sampleInput1))
	require.EqualError(t, err, "the tag tag1 does not exist")

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	err = tag.UpdateTag(transactionContext, string(sampleInput1))
	require.EqualError(t, err, "failed to read from world state: failure")

	tmpTag := &chaincode.Tag{Name: "1", Creator: "1"}
	bytes, _ := json.Marshal(tmpTag)
	chaincodeStub.GetStateReturns(bytes, nil)

	err = tag.UpdateTag(transactionContext, string(sampleInput1))
	require.NoError(t, err)

	err = tag.UpdateTag(transactionContext, "sad")
	require.EqualError(t, err, "invalid character 's' looking for beginning of value")

	tmpTag = &chaincode.Tag{Name: "1", Creator: "1"}
	bytes, _ = json.Marshal(tmpTag)
	chaincodeStub.GetStateReturns(bytes, nil)

	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = tag.UpdateTag(transactionContext, string(sampleInput1))
	require.EqualError(t, err, "failed to put to world state: failed inserting key")
}

func TestGetAllTags(t *testing.T) {
	asset := &chaincode.Tag{Name: "user1"}
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
	assets, err := userprofile.GetAllTags(transactionContext)
	require.NoError(t, err)
	require.Equal(t, []*chaincode.Tag{asset}, assets)

	iterator.HasNextReturns(true)
	iterator.NextReturns(nil, fmt.Errorf("failed retrieving next item"))
	assets, err = userprofile.GetAllTags(transactionContext)
	require.EqualError(t, err, "failed retrieving next item")
	require.Nil(t, assets)

	chaincodeStub.GetStateByRangeReturns(nil, fmt.Errorf("failed retrieving all assets"))
	assets, err = userprofile.GetAllTags(transactionContext)
	require.EqualError(t, err, "failed retrieving all assets")
	require.Nil(t, assets)
}

var sampleCategory = &chaincode.Category{
	Name:              "category1",
	Color:             "1",
	CategoryGroupName: "1",
}

var sampleInput2, _ = json.Marshal(sampleCategory)

func TestCreateCategory(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	category := chaincode.SmartContract{}

	err := category.CreateCategory(transactionContext, string(sampleInput2))
	require.NoError(t, err)

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	err = category.CreateCategory(transactionContext, string(sampleInput2))
	require.EqualError(t, err, "failed to read from world state: failure")

	expectedCategory := &chaincode.Category{Name: "1"}
	bytes, err := json.Marshal(expectedCategory)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	err = category.CreateCategory(transactionContext, string(sampleInput2))
	require.EqualError(t, err, "the category category1 already exists")

	chaincodeStub.GetStateReturns(nil, nil)
	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = category.CreateCategory(transactionContext, string(sampleInput2))
	require.EqualError(t, err, "failed to put to world state: failed inserting key")
}

func TestReadCategory(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	category := chaincode.SmartContract{}

	tmpCategory := &chaincode.Category{Name: "1"}
	bytes, _ := json.Marshal(tmpCategory)
	chaincodeStub.GetStateReturns(bytes, nil)
	_, err := category.ReadCategory(transactionContext, "1")
	require.NoError(t, err)

	err = category.CreateCategory(transactionContext, "sad")
	require.EqualError(t, err, "invalid character 's' looking for beginning of value")

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	_, err = category.ReadCategory(transactionContext, "1")
	require.EqualError(t, err, "failed to read from world state: failure")

	chaincodeStub.GetStateReturns(nil, nil)
	_, err = category.ReadCategory(transactionContext, "1")
	require.EqualError(t, err, "the category 1 does not exist")
}

func TestUpdateCategory(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	category := chaincode.SmartContract{}

	err := category.UpdateCategory(transactionContext, string(sampleInput2))
	require.EqualError(t, err, "the category category1 does not exist")

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	err = category.UpdateCategory(transactionContext, string(sampleInput2))
	require.EqualError(t, err, "failed to read from world state: failure")

	tmpCategory := &chaincode.Category{Name: "1"}
	bytes, _ := json.Marshal(tmpCategory)
	chaincodeStub.GetStateReturns(bytes, nil)

	err = category.UpdateCategory(transactionContext, string(sampleInput2))
	require.NoError(t, err)

	err = category.UpdateCategory(transactionContext, "sad")
	require.EqualError(t, err, "invalid character 's' looking for beginning of value")

	tmpCategory = &chaincode.Category{Name: "1"}
	bytes, _ = json.Marshal(tmpCategory)
	chaincodeStub.GetStateReturns(bytes, nil)

	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = category.UpdateCategory(transactionContext, string(sampleInput2))
	require.EqualError(t, err, "failed to put to world state: failed inserting key")
}

func TestGetAllCategorys(t *testing.T) {
	asset := &chaincode.Category{Name: "user1"}
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
	assets, err := userprofile.GetAllCategorys(transactionContext)
	require.NoError(t, err)
	require.Equal(t, []*chaincode.Category{asset}, assets)

	iterator.HasNextReturns(true)
	iterator.NextReturns(nil, fmt.Errorf("failed retrieving next item"))
	assets, err = userprofile.GetAllCategorys(transactionContext)
	require.EqualError(t, err, "failed retrieving next item")
	require.Nil(t, assets)

	chaincodeStub.GetStateByRangeReturns(nil, fmt.Errorf("failed retrieving all assets"))
	assets, err = userprofile.GetAllCategorys(transactionContext)
	require.EqualError(t, err, "failed retrieving all assets")
	require.Nil(t, assets)
}

var sampleCategoryGroup = &chaincode.CategoryGroup{
	Name:  "categoryGroup1",
	Color: "1",
}

var sampleInput3, _ = json.Marshal(sampleCategoryGroup)

func TestCreateCategoryGroup(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	categoryGroup := chaincode.SmartContract{}

	err := categoryGroup.CreateCategoryGroup(transactionContext, string(sampleInput3))
	require.NoError(t, err)

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	err = categoryGroup.CreateCategoryGroup(transactionContext, string(sampleInput3))
	require.EqualError(t, err, "failed to read from world state: failure")

	expectedCategoryGroup := &chaincode.CategoryGroup{Name: "1"}
	bytes, err := json.Marshal(expectedCategoryGroup)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	err = categoryGroup.CreateCategoryGroup(transactionContext, string(sampleInput3))
	require.EqualError(t, err, "the categoryGroup categoryGroup1 already exists")

	chaincodeStub.GetStateReturns(nil, nil)
	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = categoryGroup.CreateCategoryGroup(transactionContext, string(sampleInput3))
	require.EqualError(t, err, "failed to put to world state: failed inserting key")
}

func TestReadCategoryGroup(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	categoryGroup := chaincode.SmartContract{}

	tmpCategoryGroup := &chaincode.CategoryGroup{Name: "1"}
	bytes, _ := json.Marshal(tmpCategoryGroup)
	chaincodeStub.GetStateReturns(bytes, nil)
	_, err := categoryGroup.ReadCategoryGroup(transactionContext, "1")
	require.NoError(t, err)

	err = categoryGroup.CreateCategoryGroup(transactionContext, "sad")
	require.EqualError(t, err, "invalid character 's' looking for beginning of value")

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	_, err = categoryGroup.ReadCategoryGroup(transactionContext, "1")
	require.EqualError(t, err, "failed to read from world state: failure")

	chaincodeStub.GetStateReturns(nil, nil)
	_, err = categoryGroup.ReadCategoryGroup(transactionContext, "1")
	require.EqualError(t, err, "the categoryGroup 1 does not exist")
}

func TestUpdateCategoryGroup(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	categoryGroup := chaincode.SmartContract{}

	err := categoryGroup.UpdateCategoryGroup(transactionContext, string(sampleInput3))
	require.EqualError(t, err, "the categoryGroup categoryGroup1 does not exist")

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	err = categoryGroup.UpdateCategoryGroup(transactionContext, string(sampleInput3))
	require.EqualError(t, err, "failed to read from world state: failure")

	tmpCategoryGroup := &chaincode.CategoryGroup{Name: "1"}
	bytes, _ := json.Marshal(tmpCategoryGroup)
	chaincodeStub.GetStateReturns(bytes, nil)

	err = categoryGroup.UpdateCategoryGroup(transactionContext, string(sampleInput3))
	require.NoError(t, err)

	err = categoryGroup.UpdateCategoryGroup(transactionContext, "sad")
	require.EqualError(t, err, "invalid character 's' looking for beginning of value")

	tmpCategoryGroup = &chaincode.CategoryGroup{Name: "1"}
	bytes, _ = json.Marshal(tmpCategoryGroup)
	chaincodeStub.GetStateReturns(bytes, nil)

	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = categoryGroup.UpdateCategoryGroup(transactionContext, string(sampleInput3))
	require.EqualError(t, err, "failed to put to world state: failed inserting key")
}

func TestGetAllCategoryGroups(t *testing.T) {
	asset := &chaincode.CategoryGroup{Name: "user1"}
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
	assets, err := userprofile.GetAllCategoryGroups(transactionContext)
	require.NoError(t, err)
	require.Equal(t, []*chaincode.CategoryGroup{asset}, assets)

	iterator.HasNextReturns(true)
	iterator.NextReturns(nil, fmt.Errorf("failed retrieving next item"))
	assets, err = userprofile.GetAllCategoryGroups(transactionContext)
	require.EqualError(t, err, "failed retrieving next item")
	require.Nil(t, assets)

	chaincodeStub.GetStateByRangeReturns(nil, fmt.Errorf("failed retrieving all assets"))
	assets, err = userprofile.GetAllCategoryGroups(transactionContext)
	require.EqualError(t, err, "failed retrieving all assets")
	require.Nil(t, assets)
}

func prepMocksAsOrg1() (*mocks.TransactionContext, *mocks.ChaincodeStub) {
	return prepMocks(myOrg1Msp, myOrg1Clientid)
}
func prepMocksAsOrg2() (*mocks.TransactionContext, *mocks.ChaincodeStub) {
	return prepMocks(myOrg2Msp, myOrg2Clientid)
}
func prepMocks(orgMSP, clientId string) (*mocks.TransactionContext, *mocks.ChaincodeStub) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	clientIdentity := &mocks.ClientIdentity{}
	clientIdentity.GetMSPIDReturns(orgMSP, nil)
	clientIdentity.GetIDReturns(base64.StdEncoding.EncodeToString([]byte(clientId)), nil)
	// set matching msp ID using peer shim env variable
	os.Setenv("CORE_PEER_LOCALMSPID", orgMSP)
	transactionContext.GetClientIdentityReturns(clientIdentity)
	return transactionContext, chaincodeStub
}

func prepMocksIllegalId() (*mocks.TransactionContext, *mocks.ChaincodeStub) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	clientIdentity := &mocks.ClientIdentity{}
	clientIdentity.GetMSPIDReturns(myOrg1Msp, nil)
	// clientIdentity.GetIDReturns("illegal", nil)
	clientIdentity.GetIDReturns("", fmt.Errorf("failure"))
	// set matching msp ID using peer shim env variable
	os.Setenv("CORE_PEER_LOCALMSPID", myOrg1Msp)
	transactionContext.GetClientIdentityReturns(clientIdentity)
	return transactionContext, chaincodeStub
}

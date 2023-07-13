package chaincode_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"topic/chaincode"
	"topic/chaincode/mocks"

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

func TestInitLedger(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	topic := chaincode.SmartContract{}
	err := topic.InitLedger(transactionContext)
	require.NoError(t, err)

	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = topic.InitLedger(transactionContext)
	require.EqualError(t, err, "failed to put to world state. failed inserting key")
}

func TestGetSubmittingClientIdentity(t *testing.T) {
	transactionContext, _ := prepMocksAsOrg1()
	topic := chaincode.SmartContract{}

	_, err := topic.GetSubmittingClientIdentity(transactionContext)
	require.NoError(t, err)

	tr := &mocks.TransactionContext{}
	clientIdentity := &mocks.ClientIdentity{}
	clientIdentity.GetIDReturns("", fmt.Errorf("failure"))
	tr.GetClientIdentityReturns(clientIdentity)
	_, err = topic.GetSubmittingClientIdentity(tr)
	require.EqualError(t, err, "failed to read clientID: failure")
	clientIdentity.GetIDReturns("aa", nil)
	_, err = topic.GetSubmittingClientIdentity(tr)
	require.EqualError(t, err, "failed to base64 decode clientID: illegal base64 data at input byte 0")
}

func TestCreateTopic(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	topic := chaincode.SmartContract{}

	err := topic.CreateTopic(transactionContext, "1", "1", "1", "1", []string{"1", "2"}, []string{"1", "2"})
	require.NoError(t, err)

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	err = topic.CreateTopic(transactionContext, "1", "1", "1", "1", []string{"1", "2"}, []string{"1", "2"})
	require.EqualError(t, err, "failed to read from world state: failure")

	expectedTopic := &chaincode.Topic{Id: "1"}
	bytes, err := json.Marshal(expectedTopic)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	err = topic.CreateTopic(transactionContext, "1", "1", "1", "1", []string{"1", "2"}, []string{"1", "2"})
	require.EqualError(t, err, "the topic 1 already exists")

	transactionContext, _ = prepMocksIllegalId()
	err = topic.CreateTopic(transactionContext, "1", "1", "1", "1", []string{"1", "2"}, []string{"1", "2"})
	require.EqualError(t, err, "failed to read clientID: failure")
}

func TestReadTopic(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	topic := chaincode.SmartContract{}

	tmpTopic := &chaincode.Topic{Id: "1"}
	bytes, _ := json.Marshal(tmpTopic)
	chaincodeStub.GetStateReturns(bytes, nil)
	_, err := topic.ReadTopic(transactionContext, "1")
	require.NoError(t, err)

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	_, err = topic.ReadTopic(transactionContext, "1")
	require.EqualError(t, err, "failed to read from world state: failure")

	chaincodeStub.GetStateReturns(nil, nil)
	_, err = topic.ReadTopic(transactionContext, "1")
	require.EqualError(t, err, "the topic 1 does not exist")
}

func TestUpdateTopic(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	topic := chaincode.SmartContract{}

	err := topic.UpdateTopic(transactionContext, "1", "1", "1", "1", []string{"1", "2"}, []string{"1", "2"})
	require.EqualError(t, err, "the topic 1 does not exist")

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	err = topic.UpdateTopic(transactionContext, "1", "1", "1", "1", []string{"1", "2"}, []string{"1", "2"})
	require.EqualError(t, err, "failed to read from world state: failure")

	tmpTopic := &chaincode.Topic{Id: "1", Creator: myOrg1Clientid}
	bytes, _ := json.Marshal(tmpTopic)
	chaincodeStub.GetStateReturns(bytes, nil)

	err = topic.UpdateTopic(transactionContext, "1", "1", "1", "1", []string{"1", "2"}, []string{"1", "2"})
	require.NoError(t, err)

	tmpTopic = &chaincode.Topic{Id: "1", Creator: myOrg2Clientid}
	bytes, _ = json.Marshal(tmpTopic)
	chaincodeStub.GetStateReturns(bytes, nil)

	err = topic.UpdateTopic(transactionContext, "1", "1", "1", "1", []string{"1", "2"}, []string{"1", "2"})
	require.EqualError(t, err, "the topic 1 can only be updated by its creator")

	transactionContext, chaincodeStub = prepMocksIllegalId()
	tmpTopic = &chaincode.Topic{Id: "1"}
	bytes, _ = json.Marshal(tmpTopic)
	chaincodeStub.GetStateReturns(bytes, nil)

	err = topic.UpdateTopic(transactionContext, "1", "1", "1", "1", []string{"1", "2"}, []string{"1", "2"})
	require.EqualError(t, err, "failed to read clientID: failure")
}

func TestGetAllTopics(t *testing.T) {
	asset := &chaincode.Topic{Id: "user1"}
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
	assets, err := userprofile.GetAllTopics(transactionContext)
	require.NoError(t, err)
	require.Equal(t, []*chaincode.Topic{asset}, assets)

	iterator.HasNextReturns(true)
	iterator.NextReturns(nil, fmt.Errorf("failed retrieving next item"))
	assets, err = userprofile.GetAllTopics(transactionContext)
	require.EqualError(t, err, "failed retrieving next item")
	require.Nil(t, assets)

	chaincodeStub.GetStateByRangeReturns(nil, fmt.Errorf("failed retrieving all assets"))
	assets, err = userprofile.GetAllTopics(transactionContext)
	require.EqualError(t, err, "failed retrieving all assets")
	require.Nil(t, assets)
}

func TestQueryTopicsByTitle(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	topic := chaincode.SmartContract{}

	chaincodeStub.GetQueryResultReturns(nil, fmt.Errorf("failure"))
	_, err := topic.QueryTopicsByTitle(transactionContext, "1")
	require.EqualError(t, err, "failure")
}
func TestQueryTopicsByCreator(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	topic := chaincode.SmartContract{}

	chaincodeStub.GetQueryResultReturns(nil, fmt.Errorf("failure"))
	_, err := topic.QueryTopicsByCreator(transactionContext, "1")
	require.EqualError(t, err, "failure")

	tmpTopic := &chaincode.Topic{Id: "user1", Creator: myOrg1Clientid}
	bytes, _ := json.Marshal(tmpTopic)

	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, false)
	iterator.NextReturns(&queryresult.KV{Value: bytes}, nil)

	chaincodeStub.GetQueryResultReturns(iterator, nil)
	topics, err := topic.QueryTopicsByCreator(transactionContext, "1")
	require.NoError(t, err)
	require.Equal(t, []*chaincode.Topic{tmpTopic}, topics)

	iterator.NextReturns(nil, fmt.Errorf("failed retrieving next item"))
	iterator.HasNextReturnsOnCall(2, true)
	iterator.HasNextReturnsOnCall(3, true)
	chaincodeStub.GetQueryResultReturns(iterator, nil)
	topics, err = topic.QueryTopicsByCreator(transactionContext, "1")
	require.EqualError(t, err, "failed retrieving next item")
	require.Nil(t, topics)
}

func TestQueryTopicsByCategory(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	topic := chaincode.SmartContract{}

	chaincodeStub.GetQueryResultReturns(nil, fmt.Errorf("failure"))
	_, err := topic.QueryTopicsByCategory(transactionContext, "1")
	require.EqualError(t, err, "failure")
}

func TestQueryTopicsByTag(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	topic := chaincode.SmartContract{}

	chaincodeStub.GetQueryResultReturns(nil, fmt.Errorf("failure"))
	_, err := topic.QueryTopicsByTag(transactionContext, "1")
	require.EqualError(t, err, "failure")
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

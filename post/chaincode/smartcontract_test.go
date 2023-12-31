package chaincode_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"post/chaincode"
	"post/chaincode/mocks"

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

var samplePost = &chaincode.Post{
	Hash:     "1",
	Creator:  "1",
	CID:      "1",
	ReplyTo:  "1",
	BelongTo: "1",
}

var sampleInput, _ = json.Marshal(samplePost)

func TestCreatePost(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	post := chaincode.SmartContract{}

	err := post.CreatePost(transactionContext, string(sampleInput))
	require.NoError(t, err)

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	err = post.CreatePost(transactionContext, string(sampleInput))
	require.EqualError(t, err, "failed to read from world state: failure")

	expectedPost := &chaincode.Post{Hash: "1"}
	bytes, err := json.Marshal(expectedPost)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	err = post.CreatePost(transactionContext, string(sampleInput))
	require.EqualError(t, err, "the post 1 already exists")

	err = post.CreatePost(transactionContext, "sad")
	require.EqualError(t, err, "invalid character 's' looking for beginning of value")

	chaincodeStub.GetStateReturns(nil, nil)
	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = post.CreatePost(transactionContext, string(sampleInput))
	require.EqualError(t, err, "failed to put to world state: failed inserting key")

}

func TestDeletePost(t *testing.T) {
	deleteRequest := &chaincode.Delete{Hash: "1", Creator: "1"}
	deleteInput, _ := json.Marshal(deleteRequest)

	transactionContext, chaincodeStub := prepMocksAsOrg1()
	post := chaincode.SmartContract{}

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	err := post.DeletePost(transactionContext, string(deleteInput))
	require.EqualError(t, err, "failed to read from world state: failure")

	expectedPost := &chaincode.Post{Hash: "1", Creator: "1"}
	bytes, err := json.Marshal(expectedPost)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(bytes, nil)
	err = post.DeletePost(transactionContext, string(deleteInput))
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(nil, nil)
	err = post.DeletePost(transactionContext, string(deleteInput))
	require.EqualError(t, err, "the post 1 does not exist")

	expectedPost = &chaincode.Post{Hash: "1", Creator: "2"}
	bytes, err = json.Marshal(expectedPost)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(bytes, nil)
	err = post.DeletePost(transactionContext, string(deleteInput))
	require.EqualError(t, err, "the post 1 is not created by 1")

	expectedPost = &chaincode.Post{Hash: "1", Creator: "1"}
	bytes, err = json.Marshal(expectedPost)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(bytes, nil)
	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = post.DeletePost(transactionContext, string(deleteInput))
	require.EqualError(t, err, "failed to put to world state: failed inserting key")

	err = post.DeletePost(transactionContext, "sad")
	require.EqualError(t, err, "invalid character 's' looking for beginning of value")
}

func TestReadPost(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	post := chaincode.SmartContract{}

	tmpPost := &chaincode.Post{Hash: "1"}
	bytes, _ := json.Marshal(tmpPost)
	chaincodeStub.GetStateReturns(bytes, nil)
	_, err := post.ReadPost(transactionContext, "1")
	require.NoError(t, err)

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	_, err = post.ReadPost(transactionContext, "1")
	require.EqualError(t, err, "failed to read from world state: failure")

	chaincodeStub.GetStateReturns(nil, nil)
	_, err = post.ReadPost(transactionContext, "1")
	require.EqualError(t, err, "the post 1 does not exist")
}

func TestUpdatePost(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	post := chaincode.SmartContract{}

	err := post.UpdatePost(transactionContext, string(sampleInput))
	require.EqualError(t, err, "the post 1 does not exist")

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	err = post.UpdatePost(transactionContext, string(sampleInput))
	require.EqualError(t, err, "failed to read from world state: failure")

	tmpPost := &chaincode.Post{Hash: "1", Creator: "1"}
	bytes, _ := json.Marshal(tmpPost)
	chaincodeStub.GetStateReturns(bytes, nil)

	err = post.UpdatePost(transactionContext, string(sampleInput))
	require.NoError(t, err)

	err = post.UpdatePost(transactionContext, "sad")
	require.EqualError(t, err, "invalid character 's' looking for beginning of value")

	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = post.UpdatePost(transactionContext, string(sampleInput))
	require.EqualError(t, err, "failed to put to world state: failed inserting key")

}

var upvoteInput, _ = json.Marshal(&chaincode.Upvote{Hash: "1", Creator: "1"})

func TestUpvotePost(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	post := chaincode.SmartContract{}

	err := post.UpvotePost(transactionContext, string(upvoteInput))
	require.EqualError(t, err, "the post 1 does not exist")

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	err = post.UpvotePost(transactionContext, string(upvoteInput))
	require.EqualError(t, err, "failed to read from world state: failure")

	tmpPost := &chaincode.Post{Hash: "1", Creator: "1"}
	bytes, _ := json.Marshal(tmpPost)
	chaincodeStub.GetStateReturns(bytes, nil)

	err = post.UpvotePost(transactionContext, string(upvoteInput))
	require.NoError(t, err)

	tmpPost = &chaincode.Post{Hash: "1", Creator: "1", Upvotes: []string{"1"}}
	bytes, _ = json.Marshal(tmpPost)
	chaincodeStub.GetStateReturns(bytes, nil)
	err = post.UpvotePost(transactionContext, string(upvoteInput))
	require.NoError(t, err)

	tmpPost = &chaincode.Post{Hash: "1", Creator: "1", Downvotes: []string{"1"}}
	bytes, _ = json.Marshal(tmpPost)
	chaincodeStub.GetStateReturns(bytes, nil)
	err = post.UpvotePost(transactionContext, string(upvoteInput))
	require.NoError(t, err)

	err = post.UpvotePost(transactionContext, string(upvoteInput))
	require.NoError(t, err)

	err = post.UpvotePost(transactionContext, "sad")
	require.EqualError(t, err, "invalid character 's' looking for beginning of value")

	tmpPost = &chaincode.Post{Hash: "1", Creator: myOrg2Clientid}
	bytes, _ = json.Marshal(tmpPost)
	chaincodeStub.GetStateReturns(bytes, nil)

	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = post.UpvotePost(transactionContext, string(upvoteInput))
	require.EqualError(t, err, "failed to put to world state: failed inserting key")
}

func TestDownvotePost(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	post := chaincode.SmartContract{}

	err := post.DownvotePost(transactionContext, string(upvoteInput))
	require.EqualError(t, err, "the post 1 does not exist")

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	err = post.DownvotePost(transactionContext, string(upvoteInput))
	require.EqualError(t, err, "failed to read from world state: failure")

	tmpPost := &chaincode.Post{Hash: "1", Creator: "1"}
	bytes, _ := json.Marshal(tmpPost)
	chaincodeStub.GetStateReturns(bytes, nil)

	err = post.DownvotePost(transactionContext, string(upvoteInput))
	require.NoError(t, err)

	tmpPost = &chaincode.Post{Hash: "1", Creator: "1", Downvotes: []string{"1"}}
	bytes, _ = json.Marshal(tmpPost)
	chaincodeStub.GetStateReturns(bytes, nil)
	err = post.DownvotePost(transactionContext, string(upvoteInput))
	require.NoError(t, err)

	tmpPost = &chaincode.Post{Hash: "1", Creator: "1", Upvotes: []string{"1"}}
	bytes, _ = json.Marshal(tmpPost)
	chaincodeStub.GetStateReturns(bytes, nil)
	err = post.DownvotePost(transactionContext, string(upvoteInput))
	require.NoError(t, err)

	err = post.DownvotePost(transactionContext, "sad")
	require.EqualError(t, err, "invalid character 's' looking for beginning of value")

	tmpPost = &chaincode.Post{Hash: "1", Creator: myOrg2Clientid}
	bytes, _ = json.Marshal(tmpPost)
	chaincodeStub.GetStateReturns(bytes, nil)

	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = post.DownvotePost(transactionContext, string(upvoteInput))
	require.EqualError(t, err, "failed to put to world state: failed inserting key")
}

var emojiInput, _ = json.Marshal(&chaincode.Emoji{Hash: "1", Creator: "1", Code: "1"})

func TestAddEmojiPost(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	post := chaincode.SmartContract{}

	err := post.AddEmojiPost(transactionContext, string(emojiInput))
	require.EqualError(t, err, "the post 1 does not exist")

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	err = post.AddEmojiPost(transactionContext, string(emojiInput))
	require.EqualError(t, err, "failed to read from world state: failure")

	tmpPost := &chaincode.Post{Hash: "1", Creator: "1"}
	bytes, _ := json.Marshal(tmpPost)
	chaincodeStub.GetStateReturns(bytes, nil)

	err = post.AddEmojiPost(transactionContext, string(emojiInput))
	require.NoError(t, err)

	err = post.AddEmojiPost(transactionContext, "sad")
	require.EqualError(t, err, "invalid character 's' looking for beginning of value")

	tmpPost = &chaincode.Post{Hash: "1", Creator: myOrg2Clientid}
	bytes, _ = json.Marshal(tmpPost)
	chaincodeStub.GetStateReturns(bytes, nil)

	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = post.AddEmojiPost(transactionContext, string(emojiInput))
	require.EqualError(t, err, "failed to put to world state: failed inserting key")
}

func TestRemoveEmojiPost(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	post := chaincode.SmartContract{}

	err := post.RemoveEmojiPost(transactionContext, string(emojiInput))
	require.EqualError(t, err, "the post 1 does not exist")

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	err = post.RemoveEmojiPost(transactionContext, string(emojiInput))
	require.EqualError(t, err, "failed to read from world state: failure")

	tmpPost := &chaincode.Post{Hash: "1", Creator: "1"}
	bytes, _ := json.Marshal(tmpPost)
	chaincodeStub.GetStateReturns(bytes, nil)

	err = post.RemoveEmojiPost(transactionContext, string(emojiInput))
	require.NoError(t, err)

	err = post.RemoveEmojiPost(transactionContext, "sad")
	require.EqualError(t, err, "invalid character 's' looking for beginning of value")

	tmpPost = &chaincode.Post{Hash: "1", Creator: myOrg2Clientid}
	bytes, _ = json.Marshal(tmpPost)
	chaincodeStub.GetStateReturns(bytes, nil)

	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = post.RemoveEmojiPost(transactionContext, string(emojiInput))
	require.EqualError(t, err, "failed to put to world state: failed inserting key")

	testRemoveEmojiPost := &chaincode.Post{Hash: "1", Creator: "1", Emojis: make(map[string][]string)}
	testRemoveEmojiPost.Emojis["1"] = []string{"1"}
	bytes, _ = json.Marshal(testRemoveEmojiPost)
	chaincodeStub.GetStateReturns(bytes, nil)
	err = post.RemoveEmojiPost(transactionContext, string(emojiInput))
	require.EqualError(t, err, "failed to put to world state: failed inserting key")
}

func TestGetAllPosts(t *testing.T) {
	asset := &chaincode.Post{Hash: "user1"}
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
	assets, err := userprofile.GetAllPosts(transactionContext)
	require.NoError(t, err)
	require.Equal(t, []*chaincode.Post{asset}, assets)

	iterator.HasNextReturns(true)
	iterator.NextReturns(nil, fmt.Errorf("failed retrieving next item"))
	assets, err = userprofile.GetAllPosts(transactionContext)
	require.EqualError(t, err, "failed retrieving next item")
	require.Nil(t, assets)

	chaincodeStub.GetStateByRangeReturns(nil, fmt.Errorf("failed retrieving all assets"))
	assets, err = userprofile.GetAllPosts(transactionContext)
	require.EqualError(t, err, "failed retrieving all assets")
	require.Nil(t, assets)
}

func TestQueryPostsByCreator(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	Post := chaincode.SmartContract{}

	chaincodeStub.GetQueryResultReturns(nil, fmt.Errorf("failure"))
	_, err := Post.QueryPostsByCreator(transactionContext, "1")
	require.EqualError(t, err, "failure")

	tmpPost := &chaincode.Post{Hash: "user1", Creator: myOrg1Clientid}
	bytes, _ := json.Marshal(tmpPost)

	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, false)
	iterator.NextReturns(&queryresult.KV{Value: bytes}, nil)

	chaincodeStub.GetQueryResultReturns(iterator, nil)
	Posts, err := Post.QueryPostsByCreator(transactionContext, "1")
	require.NoError(t, err)
	require.Equal(t, []*chaincode.Post{tmpPost}, Posts)

	iterator.NextReturns(nil, fmt.Errorf("failed retrieving next item"))
	iterator.HasNextReturnsOnCall(2, true)
	iterator.HasNextReturnsOnCall(3, true)
	chaincodeStub.GetQueryResultReturns(iterator, nil)
	Posts, err = Post.QueryPostsByCreator(transactionContext, "1")
	require.EqualError(t, err, "failed retrieving next item")
	require.Nil(t, Posts)
}

func TestQueryPostsByBelongTo(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	Post := chaincode.SmartContract{}

	chaincodeStub.GetQueryResultReturns(nil, fmt.Errorf("failure"))
	_, err := Post.QueryPostsByBelongTo(transactionContext, "1")
	require.EqualError(t, err, "failure")
}

func TestQueryPostsByReplyTo(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	Post := chaincode.SmartContract{}

	chaincodeStub.GetQueryResultReturns(nil, fmt.Errorf("failure"))
	_, err := Post.QueryPostsByReplyTo(transactionContext, "1")
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

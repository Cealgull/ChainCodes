package chaincode

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Post struct {
	Id         string    `json:"id"`
	CId        string    `json:"cid"`
	Creator    string    `json:"creator"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
	BelongTo   string    `json:"belongTo"`
	ReplyTo    string    `json:"replyTo,omitempty" metadata:"replyTo,optional" `
	Images     []string  `json:"images,omitempty" metadata:"images,optional" `
}

const TimeFormat = "2006-01-02 15:04:05" // deprecated: use time.Time instead of string

// InitLedger adds a base set of posts to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	txntmsp, _ := ctx.GetStub().GetTxTimestamp()
	timestamp := txntmsp.AsTime()
	posts := []Post{
		{Id: "1", CId: "c1", Creator: "user1", CreateTime: timestamp, UpdateTime: timestamp, BelongTo: "1", ReplyTo: "1", Images: []string{"1.jpg", "2.jpg"}},
		{Id: "2", CId: "c2", Creator: "user2", CreateTime: timestamp, UpdateTime: timestamp, BelongTo: "2", ReplyTo: "2", Images: []string{"3.jpg", "4.jpg"}},
		{Id: "3", CId: "c3", Creator: "user3", CreateTime: timestamp, UpdateTime: timestamp, BelongTo: "3", ReplyTo: "3"},
	}

	for _, post := range posts {
		assetJSON, _ := json.Marshal(post)

		err := ctx.GetStub().PutState(post.Id, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// Get client identity which submit the transaction
func (s *SmartContract) GetSubmittingClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {

	b64ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("failed to read clientID: %v", err)
	}
	decodeID, err := base64.StdEncoding.DecodeString(b64ID)
	if err != nil {
		return "", fmt.Errorf("failed to base64 decode clientID: %v", err)
	}
	return string(decodeID), nil
}

// CreatePost creates a post.
func (s *SmartContract) CreatePost(ctx contractapi.TransactionContextInterface, postId string, cid string, operator string, belongTo string, replyTo string, imagesString string) error {
	exists, err := s.PostExists(ctx, postId)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the post %s already exists", postId)
	}

	txntmsp, _ := ctx.GetStub().GetTxTimestamp()
	timestamp := txntmsp.AsTime()
	images := strings.Split(imagesString, "-")
	post := Post{
		Id:         postId,
		CId:        cid,
		Creator:    operator,
		CreateTime: timestamp,
		UpdateTime: timestamp,
		BelongTo:   belongTo,
		ReplyTo:    replyTo,
		Images:     images,
	}

	postJSON, _ := json.Marshal(post)
	return ctx.GetStub().PutState(postId, postJSON)
}

// PostExists returns true when post with given ID exists in world state
func (s *SmartContract) PostExists(ctx contractapi.TransactionContextInterface, postId string) (bool, error) {
	postJSON, err := ctx.GetStub().GetState(postId)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return postJSON != nil, nil
}

// ReadPost returns the post stored in the world state with given id.
func (s *SmartContract) ReadPost(ctx contractapi.TransactionContextInterface, postId string) (*Post, error) {
	postJSON, err := ctx.GetStub().GetState(postId)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if postJSON == nil {
		return nil, fmt.Errorf("the post %s does not exist", postId)
	}

	var post Post
	json.Unmarshal(postJSON, &post)

	return &post, nil
}

// UpdatePost updates an existing post in the world state with provided parameters.
func (s *SmartContract) UpdatePost(ctx contractapi.TransactionContextInterface, postId string, cid string, operator string, imagesString string) error {
	post, err := s.ReadPost(ctx, postId)
	if err != nil {
		return err
	}

	if post.Creator != operator {
		return fmt.Errorf("the post %s can only be updated by its creator", postId)
	}

	txntmsp, _ := ctx.GetStub().GetTxTimestamp()
	timestamp := txntmsp.AsTime()

	post.CId = cid
	post.UpdateTime = timestamp
	images := strings.Split(imagesString, "-")
	post.Images = images
	postJSON, _ := json.Marshal(post)

	return ctx.GetStub().PutState(postId, postJSON)
}

// GetAllPosts returns all posts found in world state
func (s *SmartContract) GetAllPosts(ctx contractapi.TransactionContextInterface) ([]*Post, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var posts []*Post
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var post Post
		json.Unmarshal(queryResponse.Value, &post)
		posts = append(posts, &post)
	}

	return posts, nil
}

func (s *SmartContract) QueryPostsByCreator(ctx contractapi.TransactionContextInterface, creator string) ([]*Post, error) {
	queryString := fmt.Sprintf(`{"selector":{"creator":"%s"}}`, creator)
	return getQueryResultForQueryString(ctx, queryString)
}

func (s *SmartContract) QueryPostsByBelongTo(ctx contractapi.TransactionContextInterface, belongTo string) ([]*Post, error) {
	queryString := fmt.Sprintf(`{"selector":{"belongTo":"%s"}}`, belongTo)
	return getQueryResultForQueryString(ctx, queryString)
}

func (s *SmartContract) QueryPostsByReplyTo(ctx contractapi.TransactionContextInterface, replyTo string) ([]*Post, error) {
	queryString := fmt.Sprintf(`{"selector":{"replyTo":"%s"}}`, replyTo)
	return getQueryResultForQueryString(ctx, queryString)
}

// getQueryResultForQueryString executes the passed in query string.
// The result set is built and returned as a byte array containing the JSON results.
func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*Post, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

// constructQueryResponseFromIterator constructs a slice of posts from the resultsIterator
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*Post, error) {
	var posts []*Post
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var post Post
		json.Unmarshal(queryResult.Value, &post)
		posts = append(posts, &post)
	}

	return posts, nil
}

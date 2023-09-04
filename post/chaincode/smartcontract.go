package chaincode

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Post struct {
	Hash     string    `json:"hash"`
	Creator  string    `json:"creator"`
	CID      string    `json:"cid"`
	CreateAt time.Time `json:"createAt"`
	UpdateAt time.Time `json:"updateAt"`
	ReplyTo  string    `json:"replyTo"`
	BelongTo string    `json:"belongTo"`
	Assets   []string  `json:"assets,omitempty"`
}

// CreatePost creates a post.
func (s *SmartContract) CreatePost(ctx contractapi.TransactionContextInterface, payload string) error {

	post := Post{}
	err := json.Unmarshal([]byte(payload), &post)

	if err != nil {
		return err
	}

	exists, err := s.PostExists(ctx, post.Hash)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the post %s already exists", post.Hash)
	}

	err = ctx.GetStub().PutState(post.Hash, []byte(payload))
	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("CreatePost", []byte(payload))
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
func (s *SmartContract) UpdatePost(ctx contractapi.TransactionContextInterface, payload string) error {
	next := Post{}
	err := json.Unmarshal([]byte(payload), &next)

	if err != nil {
		return err
	}

	exists, err := s.PostExists(ctx, next.Hash)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the post %s does not exist", next.Hash)
	}

	prev, _ := s.ReadPost(ctx, next.Hash)

	x := reflect.ValueOf(&next).Elem()
	y := reflect.ValueOf(prev).Elem()

	// use reflection package to dynamically update non-zero value
	for i := 0; i < x.NumField(); i++ {
		name := x.Type().Field(i).Name
		yf := y.FieldByName(name)
		xf := x.FieldByName(name)
		if name != "Hash" && yf.CanSet() && !xf.IsZero() {
			yf.Set(xf)
		}
	}

	// overwriting original post with new post
	yJSON, _ := json.Marshal(prev)
	err = ctx.GetStub().PutState(prev.Hash, yJSON)

	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("UpdatePost", yJSON)
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

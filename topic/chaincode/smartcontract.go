package chaincode

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Topic struct {
	Id         string    `json:"id"`
	CId        string    `json:"cid"`
	Title      string    `json:"title"`
	Creator    string    `json:"creator"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
	Category   string    `json:"category"`
	Tags       []string  `json:"tags,omitempty" metadata:"tags,optional" `
}

const TimeFormat = "2006-01-02 15:04:05" // deprecated: use time.Time instead of string

// InitLedger adds a base set of topics to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	txntmsp, _ := ctx.GetStub().GetTxTimestamp()
	timestamp := txntmsp.AsTime()
	topics := []Topic{
		{Id: "1", CId: "c1", Title: "title1", Creator: "user1", CreateTime: timestamp, UpdateTime: timestamp, Category: "category1", Tags: []string{"tag1", "tag2"}},
		{Id: "2", CId: "c2", Title: "title2", Creator: "user2", CreateTime: timestamp, UpdateTime: timestamp, Category: "category2", Tags: []string{"tag1", "tag2", "tag3"}},
		{Id: "3", CId: "c3", Title: "title3", Creator: "user3", CreateTime: timestamp, UpdateTime: timestamp, Category: "category3", Tags: []string{"tag1", "tag2", "tag3", "tag4"}},
	}

	for _, topic := range topics {
		assetJSON, _ := json.Marshal(topic)

		err := ctx.GetStub().PutState(topic.Id, assetJSON)
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

// CreateTopic creates a topic.
func (s *SmartContract) CreateTopic(ctx contractapi.TransactionContextInterface, topicId string, cid string, title string, category string, tags []string) error {
	operator, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	exists, err := s.TopicExists(ctx, topicId)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the topic %s already exists", topicId)
	}

	txntmsp, _ := ctx.GetStub().GetTxTimestamp()
	timestamp := txntmsp.AsTime()
	topic := Topic{
		Id:         topicId,
		CId:        cid,
		Title:      title,
		Creator:    operator,
		CreateTime: timestamp,
		UpdateTime: timestamp,
		Category:   category,
		Tags:       tags,
	}

	topicJSON, _ := json.Marshal(topic)
	return ctx.GetStub().PutState(topicId, topicJSON)
}

// TopicExists returns true when topic with given ID exists in world state
func (s *SmartContract) TopicExists(ctx contractapi.TransactionContextInterface, topicId string) (bool, error) {
	topicJSON, err := ctx.GetStub().GetState(topicId)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return topicJSON != nil, nil
}

// ReadTopic returns the topic stored in the world state with given id.
func (s *SmartContract) ReadTopic(ctx contractapi.TransactionContextInterface, topicId string) (*Topic, error) {
	topicJSON, err := ctx.GetStub().GetState(topicId)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if topicJSON == nil {
		return nil, fmt.Errorf("the topic %s does not exist", topicId)
	}

	var topic Topic
	json.Unmarshal(topicJSON, &topic)

	return &topic, nil
}

// UpdateTopic updates an existing topic in the world state with provided parameters.
func (s *SmartContract) UpdateTopic(ctx contractapi.TransactionContextInterface, topicId string, cid string, title string, category string, tags []string) error {
	topic, err := s.ReadTopic(ctx, topicId)
	if err != nil {
		return err
	}

	operator, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if topic.Creator != operator {
		return fmt.Errorf("the topic %s can only be updated by its creator", topicId)
	}

	txntmsp, _ := ctx.GetStub().GetTxTimestamp()
	timestamp := txntmsp.AsTime()

	topic.CId = cid
	topic.Title = title
	topic.UpdateTime = timestamp
	topic.Category = category
	topic.Tags = tags
	topicJSON, _ := json.Marshal(topic)

	return ctx.GetStub().PutState(topicId, topicJSON)
}

// GetAllTopics returns all topics found in world state
func (s *SmartContract) GetAllTopics(ctx contractapi.TransactionContextInterface) ([]*Topic, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var topics []*Topic
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var topic Topic
		json.Unmarshal(queryResponse.Value, &topic)
		topics = append(topics, &topic)
	}

	return topics, nil
}

func (s *SmartContract) QueryTopicsByTitle(ctx contractapi.TransactionContextInterface, title string) ([]*Topic, error) {
	queryString := fmt.Sprintf(`{"selector":{"title":"%s"}}`, title)
	return getQueryResultForQueryString(ctx, queryString)
}

func (s *SmartContract) QueryTopicsByCreator(ctx contractapi.TransactionContextInterface, creator string) ([]*Topic, error) {
	queryString := fmt.Sprintf(`{"selector":{"creator":"%s"}}`, creator)
	return getQueryResultForQueryString(ctx, queryString)
}

func (s *SmartContract) QueryTopicsByCategory(ctx contractapi.TransactionContextInterface, category string) ([]*Topic, error) {
	queryString := fmt.Sprintf(`{"selector":{"category":"%s"}}`, category)
	return getQueryResultForQueryString(ctx, queryString)
}

func (s *SmartContract) QueryTopicsByTag(ctx contractapi.TransactionContextInterface, tag string) ([]*Topic, error) {
	queryString := fmt.Sprintf(`{"selector":{"tags":{"$elemMatch":{"$eq":"%s"}}}}`, tag)
	return getQueryResultForQueryString(ctx, queryString)
}

// getQueryResultForQueryString executes the passed in query string.
// The result set is built and returned as a byte array containing the JSON results.
func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*Topic, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

// constructQueryResponseFromIterator constructs a slice of topics from the resultsIterator
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*Topic, error) {
	var topics []*Topic
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var topic Topic
		json.Unmarshal(queryResult.Value, &topic)
		topics = append(topics, &topic)
	}

	return topics, nil
}

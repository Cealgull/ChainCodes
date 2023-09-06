package chaincode

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Topic struct {
	Hash     string   `json:"hash"`
	Title    string   `json:"title"`
	Creator  string   `json:"creator"`
	CID      string   `json:"cid"`
	Category string     `json:"category"`
	Tags     []string   `json:"tags"`
	Images   []string `json:"images"`

	Upvotes   []string          `json:"upvotes"`
	Downvotes []string          `json:"downvotes"`
	Emojis    map[uint][]string `json:"emojis"`
}

type Upvote struct {
	Hash    string `json:"hash"`
	Creator string `json:"creator"`
}

type Downvote struct {
	Hash    string `json:"hash"`
	Creator string `json:"creator"`
}

type Emoji struct {
	Hash    string `json:"hash"`
	Creator string `json:"creator"`
	Code    uint   `json:"code"`
}

// CreateTopic creates a topic.
func (s *SmartContract) CreateTopic(ctx contractapi.TransactionContextInterface, payload string) error {

	topic := Topic{}
	err := json.Unmarshal([]byte(payload), &topic)

	if err != nil {
		return err
	}

	exists, err := s.TopicExists(ctx, topic.Hash)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the topic %s already exists", topic.Hash)
	}

	err = ctx.GetStub().PutState(topic.Hash, []byte(payload))

	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("CreateTopic", []byte(payload))
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
func (s *SmartContract) UpdateTopic(ctx contractapi.TransactionContextInterface, payload string) error {
	next := Topic{}
	err := json.Unmarshal([]byte(payload), &next)

	if err != nil {
		return err
	}

	exists, err := s.TopicExists(ctx, next.Hash)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the topic %s does not exist", next.Hash)
	}

	prev, _ := s.ReadTopic(ctx, next.Hash)

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

	// overwriting original topic with new topic
	yJSON, _ := json.Marshal(prev)
	err = ctx.GetStub().PutState(prev.Hash, yJSON)

	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("UpdateTopic", []byte(payload))
}

func (s *SmartContract) UpvoteTopic(ctx contractapi.TransactionContextInterface, payload string) error {
	upvote := Upvote{}
	err := json.Unmarshal([]byte(payload), &upvote)
	if err != nil {
		return err
	}

	exists, err := s.TopicExists(ctx, upvote.Hash)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the topic %s does not exist", upvote.Hash)
	}

	topic, _ := s.ReadTopic(ctx, upvote.Hash)
	topic.Upvotes = append(topic.Upvotes, upvote.Creator)
	topicJSON, _ := json.Marshal(topic)
	err = ctx.GetStub().PutState(upvote.Hash, topicJSON)

	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("UpvoteTopic", []byte(payload))
}

func (s *SmartContract) DownvoteTopic(ctx contractapi.TransactionContextInterface, payload string) error {
	downvote := Downvote{}
	err := json.Unmarshal([]byte(payload), &downvote)
	if err != nil {
		return err
	}

	exists, err := s.TopicExists(ctx, downvote.Hash)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the topic %s does not exist", downvote.Hash)
	}

	topic, _ := s.ReadTopic(ctx, downvote.Hash)
	topic.Downvotes = append(topic.Downvotes, downvote.Creator)
	topicJSON, _ := json.Marshal(topic)

	err = ctx.GetStub().PutState(downvote.Hash, topicJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("DownvoteTopic", []byte(payload))
}

func (s *SmartContract) AddEmojiTopic(ctx contractapi.TransactionContextInterface, payload string) error {
	emoji := Emoji{}
	err := json.Unmarshal([]byte(payload), &emoji)
	if err != nil {
		return err
	}

	exists, err := s.TopicExists(ctx, emoji.Hash)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the topic %s does not exist", emoji.Hash)
	}

	topic, _ := s.ReadTopic(ctx, emoji.Hash)
	if topic.Emojis == nil {
		topic.Emojis = make(map[uint][]string)
	}
	topic.Emojis[emoji.Code] = append(topic.Emojis[emoji.Code], emoji.Creator)
	topicJSON, _ := json.Marshal(topic)

	err = ctx.GetStub().PutState(emoji.Hash, topicJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("AddEmojiTopic", []byte(payload))
}

func (s *SmartContract) RemoveEmojiTopic(ctx contractapi.TransactionContextInterface, payload string) error {
	emoji := Emoji{}
	err := json.Unmarshal([]byte(payload), &emoji)
	if err != nil {
		return err
	}

	exists, err := s.TopicExists(ctx, emoji.Hash)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the topic %s does not exist", emoji.Hash)
	}

	topic, _ := s.ReadTopic(ctx, emoji.Hash)
	if topic.Emojis[emoji.Code] != nil {
		for i, v := range topic.Emojis[emoji.Code] {
			if v == emoji.Creator {
				topic.Emojis[emoji.Code] = append(topic.Emojis[emoji.Code][:i], topic.Emojis[emoji.Code][i+1:]...)
				break
			}
		}
		if len(topic.Emojis[emoji.Code]) == 0 {
			delete(topic.Emojis, emoji.Code)
		}
	}

	topicJSON, _ := json.Marshal(topic)

	err = ctx.GetStub().PutState(emoji.Hash, topicJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("RemoveEmojiTopic", []byte(payload))
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

func (s *SmartContract) QueryTopicsByCategory(ctx contractapi.TransactionContextInterface, category uint) ([]*Topic, error) {
	queryString := fmt.Sprintf(`{"selector":{"category":"%d"}}`, category)
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

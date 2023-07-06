package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// golang keeps the order when marshal to json but doesn't order automatically
type UserProfile struct {
	IdentityId string   `json:"identityId"`
	Username   string   `json:"username"`
	Avatar     string   `json:"avatar"`
	Signature  string   `json:"signature"`
	Roles      []string `json:"roles"`
	Badge      []string `json:"badge"`
}

// InitLedger adds a base set of users to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	users := []UserProfile{
		{IdentityId: "user1", Username: "clever1", Avatar: "avatar1", Signature: "signature1", Roles: []string{"role1", "role2"}, Badge: []string{"badge1", "badge2"}},
		{IdentityId: "user2", Username: "clever2", Avatar: "avatar2", Signature: "signature2", Roles: []string{"role1", "role2"}, Badge: []string{"badge1", "badge2"}},
		{IdentityId: "user3", Username: "clever3", Avatar: "avatar3", Signature: "signature3", Roles: []string{"role1", "role2"}, Badge: []string{"badge1", "badge2"}},
	}

	for _, user := range users {
		assetJSON, _ := json.Marshal(user)

		err := ctx.GetStub().PutState(user.IdentityId, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	// return nil
	usersJson, _ := json.Marshal(users)
	return ctx.GetStub().SetEvent("InitLedger", usersJson)
}

// CreateUser creates a new user on the ledger with given details.
func (s *SmartContract) CreateUser(ctx contractapi.TransactionContextInterface, userId string, username string, avatar string, signature string) error {
	exists, err := s.UserExists(ctx, userId)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the user %s already exists", userId)
	}

	user := UserProfile{
		IdentityId: userId,
		Username:   username,
		Avatar:     avatar,
		Signature:  signature,
	}
	userJSON, _ := json.Marshal(user)

	// return ctx.GetStub().PutState(userId, userJSON)
	err = ctx.GetStub().PutState(userId, userJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	userJson, _ := json.Marshal(user)
	return ctx.GetStub().SetEvent("CreateUser", userJson)
}

// ReadUser returns the user stored in the world state with given id.
func (s *SmartContract) ReadUser(ctx contractapi.TransactionContextInterface, userId string) (*UserProfile, error) {
	userJSON, err := ctx.GetStub().GetState(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if userJSON == nil {
		return nil, fmt.Errorf("the user %s does not exist", userId)
	}

	var asset UserProfile
	json.Unmarshal(userJSON, &asset)

	// return &asset, nil
	return &asset, ctx.GetStub().SetEvent("ReadUser", userJSON)
}

func (s *SmartContract) UpdateUser(ctx contractapi.TransactionContextInterface, userId string, username string, avatar string, signature string) error {
	exists, err := s.UserExists(ctx, userId)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the user %s does not exist", userId)
	}

	// overwriting original user with new user
	user := UserProfile{
		IdentityId: userId,
		Username:   username,
		Avatar:     avatar,
		Signature:  signature,
	}
	userJSON, _ := json.Marshal(user)

	// return ctx.GetStub().PutState(userId, userJSON)
	err = ctx.GetStub().PutState(userId, userJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("UpdateUser", userJSON)
}

func (s *SmartContract) AssignRole(ctx contractapi.TransactionContextInterface, userId string, role string) error {
	user, err := s.ReadUser(ctx, userId)
	if err != nil {
		return err
	}

	user.Roles = append(user.Roles, role)

	userJSON, _ := json.Marshal(user)

	// return ctx.GetStub().PutState(userId, userJSON)
	err = ctx.GetStub().PutState(userId, userJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("AssignRole", userJSON)
}

func (s *SmartContract) RemoveRole(ctx contractapi.TransactionContextInterface, userId string, role string) error {
	user, err := s.ReadUser(ctx, userId)
	if err != nil {
		return err
	}

	for i, r := range user.Roles {
		if r == role {
			user.Roles = append(user.Roles[:i], user.Roles[i+1:]...)
			break
		}
	}

	userJSON, _ := json.Marshal(user)

	// return ctx.GetStub().PutState(userId, userJSON)
	err = ctx.GetStub().PutState(userId, userJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("RemoveRole", userJSON)
}

func (s *SmartContract) AssignBadge(ctx contractapi.TransactionContextInterface, userId string, badge string) error {
	user, err := s.ReadUser(ctx, userId)
	if err != nil {
		return err
	}

	user.Badge = append(user.Badge, badge)

	userJSON, _ := json.Marshal(user)

	// return ctx.GetStub().PutState(userId, userJSON)
	err = ctx.GetStub().PutState(userId, userJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("AssignBadge", userJSON)
}

func (s *SmartContract) RemoveBadge(ctx contractapi.TransactionContextInterface, userId string, badge string) error {
	user, err := s.ReadUser(ctx, userId)
	if err != nil {
		return err
	}

	for i, b := range user.Badge {
		if b == badge {
			user.Badge = append(user.Badge[:i], user.Badge[i+1:]...)
			break
		}
	}

	userJSON, _ := json.Marshal(user)

	// return ctx.GetStub().PutState(userId, userJSON)
	err = ctx.GetStub().PutState(userId, userJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("RemoveBadge", userJSON)
}

// UserExists returns true when asset with given ID exists in world state
func (s *SmartContract) UserExists(ctx contractapi.TransactionContextInterface, userId string) (bool, error) {
	userJSON, err := ctx.GetStub().GetState(userId)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return userJSON != nil, nil
}

// GetAllUsers returns all users found in world state
func (s *SmartContract) GetAllUsers(ctx contractapi.TransactionContextInterface) ([]*UserProfile, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*UserProfile
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset UserProfile
		json.Unmarshal(queryResponse.Value, &asset)
		assets = append(assets, &asset)
	}

	// return assets, nil
	assetsJSON, _ := json.Marshal(assets)
	return assets, ctx.GetStub().SetEvent("GetAllUsers", assetsJSON)
}

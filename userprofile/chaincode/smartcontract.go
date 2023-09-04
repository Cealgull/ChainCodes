package chaincode

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// golang keeps the order when marshal to json but doesn't order automatically
type Profile struct {
	Username  string `json:"username"`
	Wallet    string `json:"wallet"`
	Avatar    string `json:"avatar"`
	Signature string `json:"signature"`
	Muted     bool   `json:"muted"`
	Banned    bool   `json:"banned"`

	Balance     int  `json:"balance"`
	Credibility uint `json:"credibility"`

	ActiveRole     uint   `json:"activeRole"`
	RolesAssigned  []uint `json:"rolesAssigned"`
	ActiveBadge    uint   `json:"activeBadge"`
	BadgesReceived []uint `json:"badgesReceived"`
}

// CreateUser creates a new user on the ledger with given details.
func (s *SmartContract) CreateUser(ctx contractapi.TransactionContextInterface, payload string) error {

	user := Profile{}

	err := json.Unmarshal([]byte(payload), &user)

	if err != nil {
		return err
	}

	exists, err := s.UserExists(ctx, user.Wallet)

	if exists {
		return fmt.Errorf("the user wallet %s already exists", user.Wallet)
	}

	// return ctx.GetStub().PutState(userId, userJSON)
	err = ctx.GetStub().PutState(user.Wallet, []byte(payload))

	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	userJson, _ := json.Marshal(user)
	return ctx.GetStub().SetEvent("CreateUser", userJson)
}

// ReadUser returns the user stored in the world state with given id.
func (s *SmartContract) ReadUser(ctx contractapi.TransactionContextInterface, wallet string) (*Profile, error) {
	userJSON, err := ctx.GetStub().GetState(wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if userJSON == nil {
		return nil, fmt.Errorf("the user %s does not exist", wallet)
	}

	var asset Profile

	json.Unmarshal(userJSON, &asset)

	return &asset, nil
}

func (s *SmartContract) UpdateUser(ctx contractapi.TransactionContextInterface, payload string) error {

	next := Profile{}

	err := json.Unmarshal([]byte(payload), &next)

	if err != nil {
		return err
	}

	if next.Wallet == "" {
		return errors.New("wallet is required for user updating")
	}

	exists, err := s.UserExists(ctx, next.Wallet)

	if !exists {
		return fmt.Errorf("the user %s does not exist", next.Wallet)
	}

	prev, _ := s.ReadUser(ctx, next.Wallet)

	x := reflect.ValueOf(&next).Elem()
	y := reflect.ValueOf(prev).Elem()

	// use reflection package to dynamically update non-zero value
	for i := 0; i < x.NumField(); i++ {
		name := x.Type().Field(i).Name
		yf := y.FieldByName(name)
		xf := x.FieldByName(name)
		if name != "Wallet" && yf.CanSet() && !xf.IsZero() {
			yf.Set(xf)
		}
	}

	// overwriting original user with new user

	yJSON, _ := json.Marshal(prev)
	err = ctx.GetStub().PutState(prev.Wallet, yJSON)

	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("UpdateUser", []byte(payload))
}

func (s *SmartContract) AssignRole(ctx contractapi.TransactionContextInterface, wallet string, role uint) error {

	user, err := s.ReadUser(ctx, wallet)
	if err != nil {
		return err
	}

	user.RolesAssigned = append(user.RolesAssigned, role)

	userJSON, _ := json.Marshal(user)

	err = ctx.GetStub().PutState(wallet, userJSON)

	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("AssignRole", userJSON)
}

func (s *SmartContract) RemoveRole(ctx contractapi.TransactionContextInterface, wallet string, role uint) error {

	user, err := s.ReadUser(ctx, wallet)

	if err != nil {
		return err
	}

	for i, r := range user.RolesAssigned {
		if r == role {
			user.RolesAssigned = append(user.RolesAssigned[:i], user.RolesAssigned[i+1:]...)
			break
		}
	}

	userJSON, _ := json.Marshal(user)

	// return ctx.GetStub().PutState(userId, userJSON)
	err = ctx.GetStub().PutState(wallet, userJSON)

	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("RemoveRole", userJSON)

}

func (s *SmartContract) AssignBadge(ctx contractapi.TransactionContextInterface, wallet string, badge uint) error {

	user, err := s.ReadUser(ctx, wallet)
	if err != nil {
		return err
	}

	user.BadgesReceived = append(user.BadgesReceived, badge)

	userJSON, _ := json.Marshal(user)

	// return ctx.GetStub().PutState(userId, userJSON)
	err = ctx.GetStub().PutState(wallet, userJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("AssignBadge", userJSON)
}

func (s *SmartContract) RemoveBadge(ctx contractapi.TransactionContextInterface, wallet string, badge uint) error {
	user, err := s.ReadUser(ctx, wallet)
	if err != nil {
		return err
	}

	for i, b := range user.BadgesReceived {
		if b == badge {
			user.BadgesReceived = append(user.BadgesReceived[:i], user.BadgesReceived[i+1:]...)
			break
		}
	}

	userJSON, _ := json.Marshal(user)

	// return ctx.GetStub().PutState(userId, userJSON)
	err = ctx.GetStub().PutState(wallet, userJSON)
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
func (s *SmartContract) GetAllUsers(ctx contractapi.TransactionContextInterface) ([]*Profile, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Profile
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Profile
		json.Unmarshal(queryResponse.Value, &asset)
		assets = append(assets, &asset)
	}

	// return assets, nil
	assetsJSON, _ := json.Marshal(assets)
	return assets, ctx.GetStub().SetEvent("GetAllUsers", assetsJSON)
}

package chaincode

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Tag struct {
	Name        string `json:"name"`
	CreatorID   uint   `json:"creatorID"`
	Description string `json:"description"`
}

type Category struct {
	CategoryGroupID uint   `json:"categoryGroupID"`
	Color           uint   `json:"color"`
	Name            string `json:"name"`
}

type CategoryGroup struct {
	Name       string    `json:"name"`
	Color      uint      `json:"color"`
	Categories []uint    `json:"categories"`
	CreateAt   time.Time `json:"createAt"`
}

// CreateTag creates a tag.
func (s *SmartContract) CreateTag(ctx contractapi.TransactionContextInterface, payload string) error {

	tag := Tag{}
	err := json.Unmarshal([]byte(payload), &tag)

	if err != nil {
		return err
	}

	exists, err := s.TagExists(ctx, tag.Name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the tag %s already exists", tag.Name)
	}

	err = ctx.GetStub().PutState(tag.Name, []byte(payload))

	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("CreateTag", []byte(payload))
}

// TagExists returns true when tag with given name exists in world state
func (s *SmartContract) TagExists(ctx contractapi.TransactionContextInterface, tagName string) (bool, error) {
	tagJSON, err := ctx.GetStub().GetState(tagName)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return tagJSON != nil, nil
}

// ReadTag returns the tag stored in the world state with given name.
func (s *SmartContract) ReadTag(ctx contractapi.TransactionContextInterface, tagName string) (*Tag, error) {
	tagJSON, err := ctx.GetStub().GetState(tagName)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if tagJSON == nil {
		return nil, fmt.Errorf("the tag %s does not exist", tagName)
	}

	var tag Tag
	json.Unmarshal(tagJSON, &tag)

	return &tag, nil
}

// UpdateTag updates an existing tag in the world state with provided parameters.
func (s *SmartContract) UpdateTag(ctx contractapi.TransactionContextInterface, payload string) error {
	next := Tag{}
	err := json.Unmarshal([]byte(payload), &next)

	if err != nil {
		return err
	}

	exists, err := s.TagExists(ctx, next.Name)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the tag %s does not exist", next.Name)
	}

	prev, _ := s.ReadTag(ctx, next.Name)

	x := reflect.ValueOf(&next).Elem()
	y := reflect.ValueOf(prev).Elem()

	// use reflection package to dynamically update non-zero value
	for i := 0; i < x.NumField(); i++ {
		name := x.Type().Field(i).Name
		yf := y.FieldByName(name)
		xf := x.FieldByName(name)
		if yf.CanSet() && !xf.IsZero() {
			yf.Set(xf)
		}
	}

	// overwriting original tag with new tag

	err = ctx.GetStub().PutState(prev.Name, []byte(payload))

	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("UpdateTag", []byte(payload))
}

// GetAllTags returns all tags found in world state
func (s *SmartContract) GetAllTags(ctx contractapi.TransactionContextInterface) ([]*Tag, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var tags []*Tag
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var tag Tag
		json.Unmarshal(queryResponse.Value, &tag)
		tags = append(tags, &tag)
	}

	return tags, nil
}

// CreateCategory creates a category.
func (s *SmartContract) CreateCategory(ctx contractapi.TransactionContextInterface, payload string) error {

	category := Category{}
	err := json.Unmarshal([]byte(payload), &category)

	if err != nil {
		return err
	}

	exists, err := s.CategoryExists(ctx, category.Name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the category %s already exists", category.Name)
	}

	err = ctx.GetStub().PutState(category.Name, []byte(payload))

	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("CreateCategory", []byte(payload))
}

// CategoryExists returns true when category with given name exists in world state
func (s *SmartContract) CategoryExists(ctx contractapi.TransactionContextInterface, categoryName string) (bool, error) {
	categoryJSON, err := ctx.GetStub().GetState(categoryName)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return categoryJSON != nil, nil
}

// ReadCategory returns the category stored in the world state with given name.
func (s *SmartContract) ReadCategory(ctx contractapi.TransactionContextInterface, categoryName string) (*Category, error) {
	categoryJSON, err := ctx.GetStub().GetState(categoryName)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if categoryJSON == nil {
		return nil, fmt.Errorf("the category %s does not exist", categoryName)
	}

	var category Category
	json.Unmarshal(categoryJSON, &category)

	return &category, nil
}

// UpdateCategory updates an existing category in the world state with provided parameters.
func (s *SmartContract) UpdateCategory(ctx contractapi.TransactionContextInterface, payload string) error {
	next := Category{}
	err := json.Unmarshal([]byte(payload), &next)

	if err != nil {
		return err
	}

	exists, err := s.CategoryExists(ctx, next.Name)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the category %s does not exist", next.Name)
	}

	prev, _ := s.ReadCategory(ctx, next.Name)

	x := reflect.ValueOf(&next).Elem()
	y := reflect.ValueOf(prev).Elem()

	// use reflection package to dynamically update non-zero value
	for i := 0; i < x.NumField(); i++ {
		name := x.Type().Field(i).Name
		yf := y.FieldByName(name)
		xf := x.FieldByName(name)
		if yf.CanSet() && !xf.IsZero() {
			yf.Set(xf)
		}
	}

	// overwriting original category with new category

	err = ctx.GetStub().PutState(prev.Name, []byte(payload))

	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("UpdateCategory", []byte(payload))
}

// GetAllCategorys returns all categorys found in world state
func (s *SmartContract) GetAllCategorys(ctx contractapi.TransactionContextInterface) ([]*Category, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var categorys []*Category
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var category Category
		json.Unmarshal(queryResponse.Value, &category)
		categorys = append(categorys, &category)
	}

	return categorys, nil
}

// CreateCategoryGroup creates a categoryGroup.
func (s *SmartContract) CreateCategoryGroup(ctx contractapi.TransactionContextInterface, payload string) error {

	categoryGroup := CategoryGroup{}
	err := json.Unmarshal([]byte(payload), &categoryGroup)

	if err != nil {
		return err
	}

	exists, err := s.CategoryGroupExists(ctx, categoryGroup.Name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the categoryGroup %s already exists", categoryGroup.Name)
	}

	err = ctx.GetStub().PutState(categoryGroup.Name, []byte(payload))

	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("CreateCategoryGroup", []byte(payload))
}

// CategoryGroupExists returns true when categoryGroup with given name exists in world state
func (s *SmartContract) CategoryGroupExists(ctx contractapi.TransactionContextInterface, categoryGroupName string) (bool, error) {
	categoryGroupJSON, err := ctx.GetStub().GetState(categoryGroupName)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return categoryGroupJSON != nil, nil
}

// ReadCategoryGroup returns the categoryGroup stored in the world state with given name.
func (s *SmartContract) ReadCategoryGroup(ctx contractapi.TransactionContextInterface, categoryGroupName string) (*CategoryGroup, error) {
	categoryGroupJSON, err := ctx.GetStub().GetState(categoryGroupName)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if categoryGroupJSON == nil {
		return nil, fmt.Errorf("the categoryGroup %s does not exist", categoryGroupName)
	}

	var categoryGroup CategoryGroup
	json.Unmarshal(categoryGroupJSON, &categoryGroup)

	return &categoryGroup, nil
}

// UpdateCategoryGroup updates an existing categoryGroup in the world state with provided parameters.
func (s *SmartContract) UpdateCategoryGroup(ctx contractapi.TransactionContextInterface, payload string) error {
	next := CategoryGroup{}
	err := json.Unmarshal([]byte(payload), &next)

	if err != nil {
		return err
	}

	exists, err := s.CategoryGroupExists(ctx, next.Name)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the categoryGroup %s does not exist", next.Name)
	}

	prev, _ := s.ReadCategoryGroup(ctx, next.Name)

	x := reflect.ValueOf(&next).Elem()
	y := reflect.ValueOf(prev).Elem()

	// use reflection package to dynamically update non-zero value
	for i := 0; i < x.NumField(); i++ {
		name := x.Type().Field(i).Name
		yf := y.FieldByName(name)
		xf := x.FieldByName(name)
		if yf.CanSet() && !xf.IsZero() {
			yf.Set(xf)
		}
	}

	// overwriting original categoryGroup with new categoryGroup

	err = ctx.GetStub().PutState(prev.Name, []byte(payload))

	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("UpdateCategoryGroup", []byte(payload))
}

// GetAllCategoryGroups returns all categoryGroups found in world state
func (s *SmartContract) GetAllCategoryGroups(ctx contractapi.TransactionContextInterface) ([]*CategoryGroup, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var categoryGroups []*CategoryGroup
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var categoryGroup CategoryGroup
		json.Unmarshal(queryResponse.Value, &categoryGroup)
		categoryGroups = append(categoryGroups, &categoryGroup)
	}

	return categoryGroups, nil
}

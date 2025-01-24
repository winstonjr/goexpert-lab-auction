package auction

import (
	"context"
	"fmt"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"os"
	"testing"
	"time"
)

func TestCreateAuction_FindAuctionById(t *testing.T) {
	databaseName := "auctions"

	err := os.Setenv("MONGODB_DB", databaseName)
	assert.Nil(t, err)
	defer os.Clearenv()

	mTestDB := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mTestDB.Run("FindAuctionById", func(mt *mtest.T) {
		inputData := auction_entity.Auction{
			Id:          "prod_1",
			ProductName: "Prod 1",
			Category:    "Cat 1",
			Description: "Prod 1 Cat 1",
			Condition:   auction_entity.New,
			Status:      auction_entity.Active,
			Timestamp:   time.Now(),
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(
			1,
			fmt.Sprintf("%s.auctions", databaseName),
			mtest.FirstBatch,
			convertEntityToBson(inputData)))

		ar := NewAuctionRepository(mt.DB)

		result, err := ar.FindAuctionById(context.Background(), inputData.Id)
		assert.Nil(t, err)
		assert.Equal(t, inputData.Id, result.Id)
		assert.Equal(t, inputData.ProductName, result.ProductName)
		assert.Equal(t, inputData.Category, result.Category)
		assert.Equal(t, inputData.Description, result.Description)
		assert.Equal(t, inputData.Condition, result.Condition)
		assert.Equal(t, inputData.Status, result.Status)
		assert.Equal(t, inputData.Timestamp.Unix(), result.Timestamp.Unix())
	})

	mTestDB.Run("find auctions returning nil", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(0, fmt.Sprintf("%s.auctions", databaseName), mtest.FirstBatch))

		ar := NewAuctionRepository(mt.DB)

		result, err := ar.FindAuctionById(context.Background(), "any_invalid_id")
		assert.Nil(t, result)

		assert.NotNil(t, err)
		assert.Equal(t, "Error trying to find auction by id", err.Error())

		operationCalled := false
		for _, event := range mt.GetAllStartedEvents() {
			if event.CommandName == "find" { // Replace with the operation you're checking
				operationCalled = true
				filter := event.Command.Lookup("filter")
				expectedFilter := bson.D{{"_id", "any_invalid_id"}}
				expectedFilterRaw, err := convertMorDToRaw(expectedFilter)
				assert.Nil(t, err)
				assert.Equal(mt, expectedFilterRaw, filter.Document(), "Unexpected filter in FindOne operation")
				break
			}
		}
		assert.True(mt, operationCalled, "Expected operation was not called")
	})
}

func TestCreateAuction_FindAuctions(t *testing.T) {
	databaseName := "auctions"

	err := os.Setenv("MONGODB_DB", databaseName)
	assert.Nil(t, err)
	defer os.Clearenv()

	mTestDB := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mTestDB.Run("find auctions with no parameters", func(mt *mtest.T) {
		cursorResponse, endCursor, inputData := createMockAuctionList(databaseName)

		mt.AddMockResponses(cursorResponse, endCursor)

		ar := NewAuctionRepository(mt.DB)

		results, err := ar.FindAuctions(context.Background(), -1, "", "")
		assert.Nil(t, err)
		for i, result := range results {
			assert.Equal(t, inputData[i].Id, result.Id)
			assert.Equal(t, inputData[i].ProductName, result.ProductName)
			assert.Equal(t, inputData[i].Category, result.Category)
			assert.Equal(t, inputData[i].Description, result.Description)
			assert.Equal(t, inputData[i].Condition, result.Condition)
			assert.Equal(t, inputData[i].Status, result.Status)
			assert.Equal(t, inputData[i].Timestamp.Unix(), result.Timestamp.Unix())
		}
	})

	mTestDB.Run("find auctions with all filters", func(mt *mtest.T) {
		cursorResponse, endCursor, inputData := createMockAuctionList(databaseName)
		mt.AddMockResponses(cursorResponse, endCursor)

		ar := NewAuctionRepository(mt.DB)

		statusFilter := auction_entity.Active
		categoryFilter := "blah"
		productNameFilter := "flueh"
		results, err := ar.FindAuctions(context.Background(), statusFilter, categoryFilter, productNameFilter)
		assert.Nil(t, err)
		for i, result := range results {
			assert.Equal(t, inputData[i].Id, result.Id)
			assert.Equal(t, inputData[i].ProductName, result.ProductName)
			assert.Equal(t, inputData[i].Category, result.Category)
			assert.Equal(t, inputData[i].Description, result.Description)
			assert.Equal(t, inputData[i].Condition, result.Condition)
			assert.Equal(t, inputData[i].Status, result.Status)
			assert.Equal(t, inputData[i].Timestamp.Unix(), result.Timestamp.Unix())
		}
	})
}

func convertMorDToRaw[T bson.D | bson.M](doc T) (bson.Raw, error) {
	data, err := bson.Marshal(doc)
	if err != nil {
		return nil, err
	}
	raw := bson.Raw(data)

	return raw, nil
}

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

func TestAuctionRepository_CreateAuction(t *testing.T) {
	databaseName := "auctions"

	err := os.Setenv("MONGODB_DB", databaseName)
	assert.Nil(t, err)
	defer os.Clearenv()

	mto := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mto.Run("successful auction creation", func(mt *mtest.T) {
		ar := NewAuctionRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		auction := &auction_entity.Auction{
			Id:          "123",
			ProductName: "Test Product",
			Category:    "Test Category",
			Description: "Test Description",
			Condition:   auction_entity.New,
			Status:      auction_entity.Active,
			Timestamp:   time.Now(),
		}

		err := ar.CreateAuction(context.Background(), auction)
		assert.Nil(t, err)
	})

	mto.Run("database error during insertion", func(mt *mtest.T) {
		ar := NewAuctionRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    11000,
			Message: "duplicate key error",
		}))

		auction := &auction_entity.Auction{
			Id:          "123",
			ProductName: "Test Product",
			Timestamp:   time.Now(),
		}

		err := ar.CreateAuction(context.Background(), auction)
		assert.NotNil(t, err)
		assert.Equal(t, "Error trying to insert auction", err.Error())
	})

	mto.Run("context cancellation", func(mt *mtest.T) {
		ar := NewAuctionRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel the context immediately

		auction := &auction_entity.Auction{
			Id:          "123",
			ProductName: "Test Product",
			Timestamp:   time.Now(),
		}

		err := ar.CreateAuction(ctx, auction)
		assert.NotNil(t, err)
		assert.Equal(t, "Error trying to insert auction", err.Error())
	})
}

func TestAuctionRepository_CompleteExpiredAuctions(t *testing.T) {
	databaseName := "auctions"

	err := os.Setenv("MONGODB_DB", databaseName)
	assert.Nil(t, err)
	defer os.Clearenv()

	mto := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mto.Run("find auctions with no parameters", func(mt *mtest.T) {
		cursorResponse, endCursor, _ := createMockAuctionList(databaseName)

		mt.AddMockResponses(cursorResponse, endCursor,
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}, bson.E{Key: "nModified", Value: 1}))

		ar := NewAuctionRepository(mt.DB)

		err := ar.CompleteExpiredAuctions(context.Background())
		assert.Nil(t, err)

		updateOneCount := 0
		for _, event := range mt.GetAllStartedEvents() {
			if event.CommandName == "update" {
				updateOneCount++
				break
			}
		}
		assert.Equal(t, 1, updateOneCount, "Expected UpdateOne was called")
	})
}

func convertEntityListToBsonList(auctions []auction_entity.Auction) []bson.D {
	retVal := make([]bson.D, len(auctions))

	for i, auction := range auctions {
		retVal[i] = convertEntityToBson(auction)
	}

	return retVal
}

func convertEntityToBson(auction auction_entity.Auction) bson.D {
	return bson.D{
		{Key: "_id", Value: auction.Id},
		{Key: "product_name", Value: auction.ProductName},
		{Key: "category", Value: auction.Category},
		{Key: "description", Value: auction.Description},
		{Key: "condition", Value: auction.Condition},
		{Key: "status", Value: auction.Status},
		{Key: "timestamp", Value: auction.Timestamp.Unix()},
	}
}

func createMockAuctionList(databaseName string) (bson.D, bson.D, []auction_entity.Auction) {
	inputData := make([]auction_entity.Auction, 3)
	inputData[2] = auction_entity.Auction{
		Id:          "prod_1",
		ProductName: "Prod 1",
		Category:    "Cat 1",
		Description: "Prod 1 Cat 1",
		Condition:   auction_entity.New,
		Status:      auction_entity.Active,
		Timestamp:   time.Now().Add(-48 * time.Hour),
	}
	inputData[1] = auction_entity.Auction{
		Id:          "prod_2",
		ProductName: "Prod 1",
		Category:    "Cat 1",
		Description: "Prod 1 Cat 1",
		Condition:   auction_entity.Used,
		Status:      auction_entity.Active,
		Timestamp:   time.Now().Add(2 * time.Hour),
	}
	inputData[0] = auction_entity.Auction{
		Id:          "prod_3",
		ProductName: "Prod 1",
		Category:    "Cat 1",
		Description: "Prod 1 Cat 1",
		Condition:   auction_entity.Refurbished,
		Status:      auction_entity.Active,
		Timestamp:   time.Now().Add(3 * time.Hour),
	}

	cursorResponse := mtest.CreateCursorResponse(
		1,
		fmt.Sprintf("%s.auctions", databaseName),
		mtest.FirstBatch,
		convertEntityListToBsonList(inputData)...)
	endCursor := mtest.CreateCursorResponse(0, "dbName.auctions", mtest.NextBatch)

	return cursorResponse, endCursor, inputData
}

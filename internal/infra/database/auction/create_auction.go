package auction

import (
	"context"
	"fmt"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"go.mongodb.org/mongo-driver/bson"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}
type AuctionRepository struct {
	Collection                *mongo.Collection
	auctionExpirationInterval time.Duration
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	return &AuctionRepository{
		Collection:                database.Collection("auctions"),
		auctionExpirationInterval: getAuctionInterval(),
	}
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}
	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	return nil
}

func (ar *AuctionRepository) CompleteExpiredAuctions(ctx context.Context) *internal_error.InternalError {
	auctions, err := ar.FindAuctions(ctx, auction_entity.Active, "", "")
	if err != nil {
		logger.Error("Error trying to find auctions", err)
		return internal_error.NewInternalServerError("Error trying to find auctions")
	}
	var wg sync.WaitGroup
	wg.Add(len(auctions))

	bla := 0
	for _, a := range auctions {
		go func(auction *auction_entity.Auction) {
			bla = bla + 1
			defer wg.Done()
			auctionEndTime := auction.Timestamp.Add(ar.auctionExpirationInterval)
			now := time.Now()
			if now.After(auctionEndTime) {
				auctionEntityMongo := &AuctionEntityMongo{
					Id:          auction.Id,
					ProductName: auction.ProductName,
					Category:    auction.Category,
					Description: auction.Description,
					Condition:   auction.Condition,
					Status:      auction_entity.Completed,
					Timestamp:   auction.Timestamp.Unix(),
				}

				filter := bson.M{"_id": auction.Id}
				data := bson.D{{"$set", auctionEntityMongo}}
				if _, err := ar.Collection.UpdateOne(ctx, filter, data); err != nil {
					logger.Error("Error trying to complete expired auction", err)
					return
				}
			}
		}(&a)
	}
	fmt.Println(bla)
	wg.Wait()
	return nil
}

func getAuctionInterval() time.Duration {
	auctionInterval := os.Getenv("AUCTION_INTERVAL")
	duration, err := time.ParseDuration(auctionInterval)
	if err != nil {
		return time.Minute * 5
	}

	return duration
}

package auction

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/israelmiranda/go-auction/internal/entity/auction_entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupTestDatabase(t *testing.T) (*mongo.Database, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURL := os.Getenv("MONGODB_URL")
	if mongoURL == "" {
		// For local testing with authentication
		mongoURL = "mongodb://admin:admin@localhost:27017/?authSource=admin"
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to ping MongoDB: %v", err)
	}

	// Use the auctions database for testing
	db := client.Database("auctions")

	// Cleanup function
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Delete all auctions in test collection
		collection := db.Collection("auctions")
		_, err := collection.DeleteMany(ctx, bson.M{})
		if err != nil {
			t.Logf("Warning: failed to clean test database: %v", err)
		}

		// Close connection
		err = client.Disconnect(ctx)
		if err != nil {
			t.Logf("Warning: failed to disconnect from MongoDB: %v", err)
		}
	}

	return db, cleanup
}

// TestAuctionAutomaticClosure tests that an auction automatically closes after the configured interval
func TestAuctionAutomaticClosure(t *testing.T) {
	// Set a short auction interval for testing (3 seconds instead of default 5 minutes)
	os.Setenv("AUCTION_INTERVAL", "3s")
	defer os.Unsetenv("AUCTION_INTERVAL")

	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	// Create auction repository
	repo := NewAuctionRepository(db)

	// Create a test auction
	ctx := context.Background()
	testAuction, err := auction_entity.CreateAuction(
		"Test Product",
		"Electronics",
		"This is a test product for automated testing",
		auction_entity.New,
	)
	if err != nil {
		t.Fatalf("Failed to create auction entity: %v", err)
	}

	// Insert the auction
	err = repo.CreateAuction(ctx, testAuction)
	if err != nil {
		t.Fatalf("Failed to create auction in database: %v", err)
	}

	// Verify the auction was created with Active status
	auctionAfterCreate, err := repo.FindAuctionById(ctx, testAuction.Id)
	if err != nil {
		t.Fatalf("Failed to find auction after creation: %v", err)
	}

	if auctionAfterCreate.Status != auction_entity.Active {
		t.Fatalf("Expected auction status to be Active (0), got %d", auctionAfterCreate.Status)
	}

	// Wait for the auction interval + buffer
	time.Sleep(4 * time.Second)

	// Verify the auction status changed to Completed
	auctionAfterInterval, err := repo.FindAuctionById(ctx, testAuction.Id)
	if err != nil {
		t.Fatalf("Failed to find auction after interval: %v", err)
	}

	if auctionAfterInterval.Status != auction_entity.Completed {
		t.Fatalf("Expected auction status to be Completed (1), got %d after interval", auctionAfterInterval.Status)
	}
}

// TestAuctionAutomaticClosureWithCustomInterval tests auction closure with different time intervals
func TestAuctionAutomaticClosureWithCustomInterval(t *testing.T) {
	// Set a very short auction interval for faster testing (2 seconds)
	os.Setenv("AUCTION_INTERVAL", "2s")
	defer os.Unsetenv("AUCTION_INTERVAL")

	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewAuctionRepository(db)
	ctx := context.Background()

	// Create and insert auction
	testAuction, _ := auction_entity.CreateAuction(
		"Another Test Product",
		"Books",
		"This is another test product for interval testing",
		auction_entity.Used,
	)

	err := repo.CreateAuction(ctx, testAuction)
	if err != nil {
		t.Fatalf("Failed to create auction: %v", err)
	}

	// Wait for closure with buffer
	time.Sleep(3 * time.Second)

	// Verify closure
	auction, err := repo.FindAuctionById(ctx, testAuction.Id)
	if err != nil {
		t.Fatalf("Failed to find auction: %v", err)
	}

	if auction.Status != auction_entity.Completed {
		t.Errorf("Expected auction to be Completed, got status: %d", auction.Status)
	}
}

// TestMultipleAuctionsAutomaticClosure tests that multiple auctions close independently
func TestMultipleAuctionsAutomaticClosure(t *testing.T) {
	os.Setenv("AUCTION_INTERVAL", "2s")
	defer os.Unsetenv("AUCTION_INTERVAL")

	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewAuctionRepository(db)
	ctx := context.Background()

	// Create multiple auctions
	auction1, _ := auction_entity.CreateAuction(
		"Product 1",
		"Category1",
		"Description for product 1",
		auction_entity.New,
	)
	auction2, _ := auction_entity.CreateAuction(
		"Product 2",
		"Category2",
		"Description for product 2",
		auction_entity.Refurbished,
	)

	// Insert both auctions
	err1 := repo.CreateAuction(ctx, auction1)
	err2 := repo.CreateAuction(ctx, auction2)

	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to create auctions")
	}

	// Wait for closure
	time.Sleep(3 * time.Second)

	// Verify both auctions are closed
	a1, _ := repo.FindAuctionById(ctx, auction1.Id)
	a2, _ := repo.FindAuctionById(ctx, auction2.Id)

	if a1.Status != auction_entity.Completed {
		t.Errorf("Auction 1 should be Completed, got %d", a1.Status)
	}
	if a2.Status != auction_entity.Completed {
		t.Errorf("Auction 2 should be Completed, got %d", a2.Status)
	}
}

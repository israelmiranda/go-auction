# Go Auction

A RESTful API built with Go for managing online auctions with automatic closure functionality. Auctions automatically close after a configured time interval without requiring manual intervention.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Project Setup](#project-setup)
- [Environment Variables](#environment-variables)
- [Running the Project](#running-the-project)
- [Running Tests](#running-tests)
- [API Endpoints](#api-endpoints)
- [Architecture](#architecture)
- [Automatic Auction Closure](#automatic-auction-closure)

## Prerequisites

Before running this project, ensure you have the following installed:

- **Go**: Version 1.26 or higher ([Download](https://golang.org/dl/))
- **Docker**: Latest version with Docker Compose ([Download](https://www.docker.com/products/docker-desktop))
- **MongoDB**: Version 5.0 or higher (installed via Docker Compose)

## Environment Variables

### Create the .env File

The project requires environment variables to configure MongoDB connection and auction behavior.

1. Copy the example environment file:

```bash
cp cmd/auction/.env.example cmd/auction/.env
```

2. Edit `cmd/auction/.env` with your configuration:

```env
# Batch Configuration
BATCH_INSERT_INTERVAL=20s
MAX_BATCH_SIZE=4

# Auction Configuration
# Time after which an auction automatically closes
# Format: Go duration (5m = 5 minutes, 10s = 10 seconds, 2h30m = 2 hours 30 minutes)
AUCTION_INTERVAL=20s

# MongoDB Configuration
MONGO_INITDB_ROOT_USERNAME=admin
MONGO_INITDB_ROOT_PASSWORD=admin
MONGODB_URL=mongodb://admin:admin@mongodb:27017/auctions?authSource=admin
MONGODB_DB=auctions
```

### Environment Variables Description

| Variable | Description | Example | Required |
|----------|-------------|---------|----------|
| `BATCH_INSERT_INTERVAL` | Interval for batch operations | `20s` | No |
| `MAX_BATCH_SIZE` | Maximum batch size for operations | `4` | No |
| `AUCTION_INTERVAL` | Time until auction auto-closes | `5m`, `10s`, `2h` | Yes |
| `MONGO_INITDB_ROOT_USERNAME` | MongoDB admin username | `admin` | Yes |
| `MONGO_INITDB_ROOT_PASSWORD` | MongoDB admin password | `admin` | Yes |
| `MONGODB_URL` | MongoDB connection string | `mongodb://user:pass@host:port/db?authSource=admin` | Yes |
| `MONGODB_DB` | MongoDB database name | `auctions` | Yes |

## Running the Project

### Option 1: Using Docker Compose (Recommended)

The easiest way to run the entire project with all dependencies:

```bash
docker compose up
```

This command will:
- Build the Go application image
- Start the MongoDB container
- Run the application on `http://localhost:8080`
- Configure all environment variables automatically

To run in the background:

```bash
docker compose up -d
```

To stop the application:

```bash
docker compose down
```

To view logs:

```bash
docker compose logs -f app
```

### Option 2: Local Development (without Docker)

If you prefer running directly on your machine:

1. **Start MongoDB** (using Docker):

```bash
docker run -d \
  --name mongodb \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=admin \
  mongo:latest
```

2. **Run the application**:

```bash
go run cmd/auction/main.go
```

The API will be available at `http://localhost:8080`

3. **Stop MongoDB**:

```bash
docker stop mongodb && docker rm mongodb
```

## Running Tests

### Prerequisites for Testing

Ensure MongoDB is running before executing tests. Use Docker to start it:

```bash
docker run -d --name mongodb-test -p 27017:27017 mongo:latest
```

### Run All Tests

Execute all tests in the project:

```bash
go test ./...
```

This will test all packages and display coverage information.

### Run Specific Test Suite

Test only the auction repository (including automatic closure tests):

```bash
go test -v ./internal/infra/database/auction/...
```

### Run Specific Test Case

Test the automatic auction closure feature:

```bash
go test -v -run TestAuctionAutomaticClosure ./internal/infra/database/auction/...
```

### Test Cases

The test suite includes the following test cases for automatic auction closure:

#### 1. **TestAuctionAutomaticClosure**
Tests the basic automatic closure functionality:
- Creates an auction with a 3-second auto-close interval
- Verifies the auction starts with `Active` status
- Waits for the interval to elapse
- Confirms the status changes to `Completed` without manual intervention

```bash
go test -v -run TestAuctionAutomaticClosure ./internal/infra/database/auction/...
```

#### 2. **TestAuctionAutomaticClosureWithCustomInterval**
Tests closure with different time intervals:
- Creates an auction with a 2-second interval
- Verifies automatic closure works with custom durations

```bash
go test -v -run TestAuctionAutomaticClosureWithCustomInterval ./internal/infra/database/auction/...
```

#### 3. **TestMultipleAuctionsAutomaticClosure**
Tests that multiple auctions close independently:
- Creates two auctions simultaneously
- Verifies both automatically close after the interval without interfering with each other

```bash
go test -v -run TestMultipleAuctionsAutomaticClosure ./internal/infra/database/auction/...
```

### Stop Test MongoDB

After testing, stop the MongoDB container:

```bash
docker stop mongodb-test && docker rm mongodb-test
```

## API Endpoints

### Auction Endpoints

#### Create Auction
```http
POST /auction
Content-Type: application/json

{
  "product_name": "iPhone 13",
  "category": "Electronics",
  "description": "New iPhone 13 in excellent condition",
  "condition": 0
}
```

**Condition Values:**
- `0` = New
- `1` = Used
- `2` = Refurbished

**Response:**
```http
HTTP/1.1 200 OK
```

#### Get All Auctions
```http
GET /auction?status=0&category=Electronics&product_name=iPhone
```

**Query Parameters:**
- `status` (optional): `0` = Active, `1` = Completed
- `category` (optional): Filter by category
- `product_name` (optional): Filter by product name

**Response:**
```json
[
  {
    "id": "uuid",
    "product_name": "iPhone 13",
    "category": "Electronics",
    "description": "New iPhone 13",
    "condition": 0,
    "status": 0,
    "timestamp": "2024-01-15 10:30:45"
  }
]
```

#### Get Auction by ID
```http
GET /auction/:auctionId
```

**Response:**
```json
{
  "id": "uuid",
  "product_name": "iPhone 13",
  "category": "Electronics",
  "description": "New iPhone 13",
  "condition": 0,
  "status": 0,
  "timestamp": "2024-01-15 10:30:45"
}
```

#### Get Winning Bid
```http
GET /auction/winner/:auctionId
```

**Response:**
```json
{
  "auction": {
    "id": "uuid",
    "product_name": "iPhone 13",
    "category": "Electronics",
    "description": "New iPhone 13",
    "condition": 0,
    "status": 1,
    "timestamp": "2024-01-15 10:30:45"
  },
  "bid": {
    "id": "bid-uuid",
    "auction_id": "uuid",
    "user_id": "user-uuid",
    "value": 900.50,
    "timestamp": "2024-01-15 10:45:30"
  }
}
```

### Bid Endpoints

#### Create Bid
```http
POST /bid
Content-Type: application/json

{
  "auction_id": "uuid",
  "user_id": "user-uuid",
  "value": 850.00
}
```

**Response:**
```http
HTTP/1.1 200 OK
```

#### Get Bids by Auction
```http
GET /bid/:auctionId
```

**Response:**
```json
[
  {
    "id": "bid-uuid",
    "auction_id": "uuid",
    "user_id": "user-uuid",
    "value": 850.00,
    "timestamp": "2024-01-15 10:45:30"
  }
]
```

### User Endpoints

#### Get User by ID
```http
GET /user/:userId
```

**Response:**
```json
{
  "id": "user-uuid",
  "name": "John Doe",
  "email": "john@example.com"
}
```

## Architecture

The project follows a **Clean Architecture** pattern with clear separation of concerns:

```
├── cmd/auction/
│   ├── main.go              # Application entry point
│   └── .env                 # Environment configuration
├── configuration/           # Config setup
│   ├── database/
│   │   └── mongodb/         # MongoDB connection
│   └── logger/              # Logging configuration
├── internal/                # Internal packages
│   ├── entity/              # Domain entities
│   │   ├── auction_entity/
│   │   ├── bid_entity/
│   │   └── user_entity/
│   ├── infra/               # Infrastructure layer
│   │   ├── api/
│   │   │   └── web/         # REST controllers & validation
│   │   └── database/        # Repository implementations
│   ├── usecase/             # Use cases / business logic
│   │   ├── auction_usecase/
│   │   ├── bid_usecase/
│   │   └── user_usecase/
│   └── internal_error/      # Error handling
```

### Layers

1. **Entity Layer**: Domain models (Auction, Bid, User)
2. **UseCase Layer**: Business logic and rules
3. **Infrastructure Layer**: Database operations and API controllers
4. **Configuration Layer**: External service setup

## Automatic Auction Closure

### How It Works

The automatic auction closure feature ensures that auctions close automatically after a configured time interval without requiring manual intervention.

#### Implementation Details

1. **Goroutine-Based Timer**: When an auction is created, a background goroutine starts a timer
2. **Configurable Duration**: The closure time is controlled by the `AUCTION_INTERVAL` environment variable
3. **Status Update**: After the timer expires, the auction status automatically changes from `Active` (0) to `Completed` (1)
4. **Database Update**: The change is persisted immediately to MongoDB

#### Example Flow

```
1. User creates auction with AUCTION_INTERVAL=5m
2. Goroutine starts timer for 5 minutes
3. Timer runs in background (non-blocking)
4. After 5 minutes: status automatically changed to Completed
5. Subsequent API calls show Completed status
6. No manual action required
```

#### Code Location

See [create_auction.go](internal/infra/database/auction/create_auction.go#L48-L68) for implementation details.

### Testing Automatic Closure

To verify the automatic closure functionality works correctly:

```bash
# Run all closure tests
go test -v ./internal/infra/database/auction/...

# Test basic closure
go test -v -run TestAuctionAutomaticClosure ./internal/infra/database/auction/

# Test with multiple auctions
go test -v -run TestMultipleAuctionsAutomaticClosure ./internal/infra/database/auction/
```

### Monitoring Auction Status

You can check an auction's status at any time:

```bash
# Check if auction is still active
curl http://localhost:8080/auction/{auctionId}

# Get winning bid (only works for completed auctions)
curl http://localhost:8080/auction/winner/{auctionId}
```

## Troubleshooting

### MongoDB Connection Error

**Problem**: `Error trying to connect to mongodb database`

**Solution**:
1. Ensure MongoDB is running: `docker ps | grep mongo`
2. Check the `MONGODB_URL` in `.env` is correct
3. Verify MongoDB credentials match the connection string
4. Restart MongoDB: `docker restart mongodb`

### Tests Fail to Connect to MongoDB

**Problem**: Tests timeout or cannot connect

**Solution**:
1. Ensure test MongoDB is running: `docker run -d --name mongodb-test -p 27017:27017 mongo:latest`
2. Check port 27017 is available: `lsof -i :27017`
3. Clear MongoDB test database: `docker exec mongodb-test mongosh -u admin -p admin --eval "use test; db.dropDatabase()"`

### Auction Not Closing Automatically

**Problem**: Auction status remains Active after time interval

**Solution**:
1. Verify `AUCTION_INTERVAL` is set correctly in `.env`
2. Check application logs for errors: `docker compose logs app`
3. Ensure MongoDB is accessible and running
4. Check database permissions for the configured user

### Build Error: Module Not Found

**Problem**: `cannot find module`

**Solution**:
```bash
go mod tidy
go mod download
```

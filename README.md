# Internal Transfers System

A Go-based API service for handling internal financial transfers between accounts.

## Features

- **Account Management**: Create accounts and query balances
- **Atomic Transactions**: Secure transaction processing between accounts with database-level atomicity
- **Connection Pooling**: Optimized database connection management for high concurrency
- **PostgreSQL Integration**: Robust transaction log and account state management
- **RESTful API**: Clean, well-documented HTTP endpoints
- **Error Handling**: Comprehensive error handling with appropriate HTTP status codes
- **Docker Support**: Complete containerization with simplified setup

## Architecture

### Database Transaction Pattern

The system uses a two-tier approach for database operations:

- **Standalone Operations**: For individual account operations (create, read)
- **Transaction-Aware Operations**: For atomic operations requiring consistency (transfers)

```go
// Standalone operations (no transaction context needed)
CreateAccount(ctx, accountID, initialBalance)
GetAccount(ctx, accountID)

// Transaction-aware operations (used within database transactions)
GetAccountWithTx(ctx, tx, accountID)
UpdateBalanceWithTx(ctx, tx, accountID, newBalance)
CreateTransactionWithTx(ctx, tx, transaction)
```

### Connection Pooling

The system implements efficient database connection pooling with configurable settings:

- **MaxOpenConns**: 25 (default) - Maximum concurrent database connections
- **MaxIdleConns**: 5 (default) - Connections kept in pool when idle
- **ConnMaxLifetime**: 30 minutes (default) - Connection recycling interval

### Docker Architecture

The application is containerized using Docker Compose with:

- **API Container**: Go application with connection pooling
- **Database Container**: PostgreSQL with automatic migration execution
- **Network**: Internal container communication
- **Volumes**: Persistent database storage
- **Health Checks**: Service readiness monitoring

## Prerequisites

### For Local Development
- Go 1.22 or later
- PostgreSQL 12 or later
- `psql` command-line tool (for database setup)

### For Docker (Recommended)
- Docker 20.10 or later
- Docker Compose 2.0 or later

## Setup

### Option 1: Docker Setup (Recommended)

1. **Clone the repository**:
   ```bash
   git clone https://github.com/khamiruf/internal_transfers_system_go.git
   cd internal_transfers_system_go
   ```

2. **Start the application**:
   ```bash
   # Using Makefile (recommended)
   make docker-up
   
   # Or using docker-compose directly
   docker compose up -d
   ```

3. **Verify the setup**:
   ```bash
   # Check container status
   make docker-status
   
   # View logs
   make docker-logs
   
   # Test the API
   curl http://localhost:8080/health
   ```

### Option 2: Local Development Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/khamiruf/internal_transfers_system_go.git
   cd internal_transfers_system_go
   ```

2. **Set up the database**:
   ```bash
   # Create the database
   createdb transfers

   # Apply migrations
   psql -d transfers -f migrations/001_init.sql
   ```

3. **Configure environment variables** (optional):
   ```bash
   export DATABASE_URL="postgres://username:password@localhost:5432/transfers?sslmode=disable"
   export PORT=8080
   export MAX_DB_CONNECTIONS=25
   export MAX_IDLE_CONNECTIONS=5
   export CONN_MAX_LIFETIME_MINUTES=30
   ```

4. **Build and run the application**:
   ```bash
   go build -o transfers_api ./cmd/api
   ./transfers_api
   ```

## Docker Management

### Quick Commands

```bash
# Start the application
make docker-up

# Stop all services
make docker-down

# View logs
make docker-logs

# Check status
make docker-status
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_DB` | `transfers` | Database name |
| `POSTGRES_USER` | `transfers_user` | Database user |
| `POSTGRES_PASSWORD` | `transfers_password` | Database password |
| `POSTGRES_HOST` | `db` | Database host (use `db` for Docker) |
| `POSTGRES_PORT` | `5432` | Database port |
| `DATABASE_URL` | Auto-generated | Full database connection string |
| `PORT` | `8080` | API server port |
| `MAX_DB_CONNECTIONS` | `25` | Maximum database connections |
| `MAX_IDLE_CONNECTIONS` | `5` | Maximum idle connections |
| `CONN_MAX_LIFETIME_MINUTES` | `30` | Connection lifetime in minutes |
| `LOG_LEVEL` | `debug` | Logging level |

## API Endpoints

### Health Check
- **GET** `/health`
- Returns service health status
- Used by Docker health checks

### Create Account
- **POST** `/accounts`
- Creates a new account with the specified ID and initial balance
- Request body:
  ```json
  {
    "account_id": 123,
    "initial_balance": "100.23344"
  }
  ```
- Response: `201 Created` on success

### Get Account Balance
- **GET** `/accounts/{account_id}`
- Retrieves account information including current balance
- Response:
  ```json
  {
    "account_id": 123,
    "balance": "100.23344"
  }
  ```

### Create Transaction
- **POST** `/transactions`
- Processes an atomic transfer between two accounts
- All operations (balance checks, updates, transaction recording) are performed within a single database transaction
- Request body:
  ```json
  {
    "source_account_id": 123,
    "destination_account_id": 456,
    "amount": "100.12345"
  }
  ```
- Response: `201 Created` on success

## Database Schema

### Accounts Table
```sql
CREATE TABLE accounts (
    account_id BIGINT PRIMARY KEY,
    balance DECIMAL(20,5) NOT NULL CHECK (balance >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Transactions Table
```sql
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    source_account_id BIGINT NOT NULL,
    destination_account_id BIGINT NOT NULL,
    amount DECIMAL(20,5) NOT NULL CHECK (amount > 0),
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (source_account_id) REFERENCES accounts(account_id),
    FOREIGN KEY (destination_account_id) REFERENCES accounts(account_id)
);
```

## Error Handling

The API returns appropriate HTTP status codes and structured error responses:

- **400 Bad Request**: Invalid input data (negative amounts, same account transfer)
- **404 Not Found**: Account not found
- **409 Conflict**: Account already exists
- **422 Unprocessable Entity**: Insufficient balance
- **500 Internal Server Error**: Database or system errors

Error response format:
```json
{
  "error": {
    "code": "insufficient_balance",
    "message": "insufficient balance"
  }
}
```

## Development

### Running Tests
```bash
# Run all tests with cache cleaning and coverage
make test

# Run specific test packages
go test ./internal/repository/...
go test ./internal/service/...
go test ./internal/api/handlers/...
```

### Project Structure
```
├── cmd/api/                 # Application entry point
├── internal/
│   ├── api/                # HTTP handlers and server setup
│   ├── config/             # Configuration management
│   ├── errors/             # Domain-specific errors
│   ├── logger/             # Logging utilities
│   ├── models/             # Domain models
│   ├── repository/         # Database access layer
│   └── service/            # Business logic layer
├── migrations/             # Database schema migrations
├── Dockerfile              # Application container
├── docker-compose.yml      # Docker orchestration
├── .env.example            # Environment template
├── Makefile                # Development commands
└── postman/               # API documentation and test collections
```

### Key Design Patterns

1. **Repository Pattern**: Clean separation between data access and business logic
2. **Service Layer**: Business logic encapsulation with transaction management
3. **Dependency Injection**: Loose coupling between components
4. **Error Handling**: Consistent error types and HTTP status codes
5. **Connection Pooling**: Efficient database resource management
6. **Containerization**: Docker-based deployment and development

## Performance Considerations

- **Connection Pooling**: Reduces connection establishment overhead by 5-10x
- **Atomic Transactions**: Ensures data consistency during transfers
- **Indexed Queries**: Optimized database indexes for fast lookups
- **Prepared Statements**: Efficient query execution with parameterized queries
- **Container Optimization**: Multi-stage builds and Alpine Linux for minimal image size

## Troubleshooting

### Docker Issues

1. **Container won't start**: Check logs with `make docker-logs`
2. **Database connection failed**: Verify `.env` file configuration
3. **Port conflicts**: Change ports in `.env` file
4. **Permission issues**: Ensure Docker has proper permissions

### Common Commands

```bash
# Rebuild containers
docker compose build --no-cache

# Reset database
make docker-down
make docker-up

# View container resources
docker stats

# Access database directly
docker compose exec db psql -U transfers_user -d transfers
```

## Security

### SQL Injection Protection
- All database queries use parameterized statements
- No string concatenation in SQL queries
- Input validation at multiple layers

### Best Practices
- Non-root Docker containers
- Environment variable configuration
- Connection pooling with limits
- Graceful error handling
- Comprehensive input validation

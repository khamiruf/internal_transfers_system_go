package repository

import (
	"context"
	"database/sql"

	"github.com/khamiruf/internal_transfers_system_go/internal/models"
	"github.com/shopspring/decimal"
)

// AccountRepository defines the interface for account-related database operations
//
// Transaction Pattern:
//   - Standalone operations (CreateAccount, GetAccount) are used for individual account operations
//   - Transaction-aware operations (GetAccountWithTx, UpdateBalanceWithTx) are used within database transactions
//     for operations that require atomicity (like transfers between accounts)
type AccountRepository interface {
	// CreateAccount creates a new account with the given ID and initial balance
	// This is a standalone operation that doesn't require transaction context
	CreateAccount(ctx context.Context, accountID int64, initialBalance decimal.Decimal) error

	// GetAccount retrieves an account by its ID
	// This is a standalone operation for reading account data
	GetAccount(ctx context.Context, accountID int64) (*models.Account, error)

	// Transaction-aware methods - used within database transactions for atomic operations

	// GetAccountWithTx retrieves an account by its ID within a transaction
	// Used when account data is needed as part of a larger atomic operation
	GetAccountWithTx(ctx context.Context, tx *sql.Tx, accountID int64) (*models.Account, error)

	// UpdateBalanceWithTx updates an account's balance within a transaction
	// Used for balance updates that must be atomic (e.g., during transfers)
	UpdateBalanceWithTx(ctx context.Context, tx *sql.Tx, accountID int64, newBalance decimal.Decimal) error
}

// TransactionRepository defines the interface for transaction-related database operations
//
// Transaction Pattern:
// - GetTransactionsByAccount is a standalone read operation
// - CreateTransactionWithTx is used within database transactions for atomic transaction recording
type TransactionRepository interface {
	// GetTransactionsByAccount retrieves all transactions for a given account
	// This is a standalone read operation that doesn't require transaction context
	GetTransactionsByAccount(ctx context.Context, accountID int64) ([]*models.Transaction, error)

	// Transaction-aware methods - used within database transactions for atomic operations

	// CreateTransactionWithTx creates a transaction record within a database transaction
	// Used when recording transactions as part of a larger atomic operation (e.g., during transfers)
	// Returns the created transaction with the generated ID and timestamp
	CreateTransactionWithTx(ctx context.Context, tx *sql.Tx, transaction *models.Transaction) (*models.Transaction, error)
}

package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/khamiruf/internal_transfers_system_go/internal/errors"
	"github.com/khamiruf/internal_transfers_system_go/internal/logger"
	"github.com/khamiruf/internal_transfers_system_go/internal/models"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

var (
	ErrAccountNotFound     = errors.ErrAccountNotFound
	ErrInsufficientBalance = errors.ErrInsufficientBalance
)

type PostgresAccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *PostgresAccountRepository {
	return &PostgresAccountRepository{db: db}
}

// CreateAccount creates a new account with the given ID and initial balance
func (r *PostgresAccountRepository) CreateAccount(ctx context.Context, accountID int64, initialBalance decimal.Decimal) error {
	logger.Info("Creating account in database: account_id=%d, initial_balance=%s", accountID, initialBalance.String())

	// Validate initial balance
	if initialBalance.IsNegative() {
		logger.Warn("Invalid initial balance for account %d: %s (negative amount)", accountID, initialBalance.String())
		return errors.ErrInvalidAmount
	}

	query := `
		INSERT INTO accounts (account_id, balance)
		VALUES ($1, $2)
	`
	_, err := r.db.ExecContext(ctx, query, accountID, initialBalance)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				logger.Warn("Account already exists in database: %d", accountID)
				return errors.ErrAccountAlreadyExists
			case "check_constraint_violation":
				logger.Warn("Check constraint violation for account %d: %s", accountID, initialBalance.String())
				return errors.ErrInvalidAmount
			}
		}
		logger.Error("Database error creating account %d: %v", accountID, err)
		return fmt.Errorf("failed to create account: %w", err)
	}

	logger.Info("Successfully created account in database: account_id=%d", accountID)
	return nil
}

// GetAccount retrieves an account by its ID
func (r *PostgresAccountRepository) GetAccount(ctx context.Context, accountID int64) (*models.Account, error) {
	logger.Info("Retrieving account from database: account_id=%d", accountID)

	query := `
		SELECT account_id, balance
		FROM accounts
		WHERE account_id = $1
	`
	var account models.Account
	err := r.db.QueryRowContext(ctx, query, accountID).Scan(&account.AccountID, &account.Balance)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warn("Account not found in database: %d", accountID)
			return nil, errors.ErrAccountNotFound
		}
		logger.Error("Database error retrieving account %d: %v", accountID, err)
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	logger.Info("Successfully retrieved account from database: account_id=%d, balance=%s", accountID, account.Balance.String())
	return &account, nil
}

// GetAccountWithTx retrieves an account by its ID within a transaction
func (r *PostgresAccountRepository) GetAccountWithTx(ctx context.Context, tx *sql.Tx, accountID int64) (*models.Account, error) {
	logger.Info("Retrieving account within transaction: account_id=%d", accountID)

	query := `
		SELECT account_id, balance
		FROM accounts
		WHERE account_id = $1
	`
	var account models.Account
	err := tx.QueryRowContext(ctx, query, accountID).Scan(&account.AccountID, &account.Balance)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warn("Account not found in database (transaction): %d", accountID)
			return nil, errors.ErrAccountNotFound
		}
		logger.Error("Database error retrieving account %d (transaction): %v", accountID, err)
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	logger.Info("Successfully retrieved account within transaction: account_id=%d, balance=%s", accountID, account.Balance.String())
	return &account, nil
}

// UpdateBalanceWithTx updates an account's balance within a transaction
func (r *PostgresAccountRepository) UpdateBalanceWithTx(ctx context.Context, tx *sql.Tx, accountID int64, newBalance decimal.Decimal) error {
	logger.Info("Updating account balance within transaction: account_id=%d, new_balance=%s", accountID, newBalance.String())

	// Validate new balance
	if newBalance.IsNegative() {
		logger.Warn("Invalid new balance for account %d: %s (negative amount)", accountID, newBalance.String())
		return errors.ErrInvalidAmount
	}

	query := `
		UPDATE accounts
		SET balance = $1
		WHERE account_id = $2
	`
	result, err := tx.ExecContext(ctx, query, newBalance, accountID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "check_constraint_violation" {
			logger.Warn("Check constraint violation updating account %d balance: %s", accountID, newBalance.String())
			return errors.ErrInvalidAmount
		}
		logger.Error("Database error updating account %d balance: %v", accountID, err)
		return fmt.Errorf("failed to update balance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected for account %d: %v", accountID, err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Warn("No rows affected when updating account %d balance", accountID)
		return errors.ErrAccountNotFound
	}

	logger.Info("Successfully updated account balance within transaction: account_id=%d, new_balance=%s", accountID, newBalance.String())
	return nil
}

package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/khamiruf/internal_transfers_system_go/internal/errors"
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
	// Validate initial balance
	if initialBalance.IsNegative() {
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
				return errors.ErrAccountAlreadyExists
			case "check_constraint_violation":
				return errors.ErrInvalidAmount
			}
		}
		return fmt.Errorf("failed to create account: %w", err)
	}
	return nil
}

// GetAccount retrieves an account by its ID
func (r *PostgresAccountRepository) GetAccount(ctx context.Context, accountID int64) (*models.Account, error) {
	query := `
		SELECT account_id, balance
		FROM accounts
		WHERE account_id = $1
	`
	var account models.Account
	err := r.db.QueryRowContext(ctx, query, accountID).Scan(&account.AccountID, &account.Balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &account, nil
}

// GetAccountWithTx retrieves an account by its ID within a transaction
func (r *PostgresAccountRepository) GetAccountWithTx(ctx context.Context, tx *sql.Tx, accountID int64) (*models.Account, error) {
	query := `
		SELECT account_id, balance
		FROM accounts
		WHERE account_id = $1
	`
	var account models.Account
	err := tx.QueryRowContext(ctx, query, accountID).Scan(&account.AccountID, &account.Balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &account, nil
}

// UpdateBalanceWithTx updates an account's balance within a transaction
func (r *PostgresAccountRepository) UpdateBalanceWithTx(ctx context.Context, tx *sql.Tx, accountID int64, newBalance decimal.Decimal) error {
	// Validate new balance
	if newBalance.IsNegative() {
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
			return errors.ErrInvalidAmount
		}
		return fmt.Errorf("failed to update balance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrAccountNotFound
	}

	return nil
}

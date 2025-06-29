package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/khamiruf/internal_transfers_system_go/internal/errors"
	"github.com/khamiruf/internal_transfers_system_go/internal/logger"
	"github.com/khamiruf/internal_transfers_system_go/internal/models"
	"github.com/lib/pq"
)

type PostgresTransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{db: db}
}

func (r *PostgresTransactionRepository) GetTransactionsByAccount(ctx context.Context, accountID int64) ([]*models.Transaction, error) {
	logger.Info("Retrieving transactions for account: %d", accountID)

	query := `
		SELECT id, source_account_id, destination_account_id, amount, status, created_at
		FROM transactions
		WHERE source_account_id = $1 OR destination_account_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
		logger.Error("Database error retrieving transactions for account %d: %v", accountID, err)
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		var tx models.Transaction
		var createdAt time.Time
		err := rows.Scan(
			&tx.ID,
			&tx.SourceAccountID,
			&tx.DestinationAccountID,
			&tx.Amount,
			&tx.Status,
			&createdAt,
		)
		if err != nil {
			logger.Error("Failed to scan transaction for account %d: %v", accountID, err)
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		tx.CreatedAt = createdAt.Format(time.RFC3339)
		transactions = append(transactions, &tx)
	}

	if err = rows.Err(); err != nil {
		logger.Error("Error iterating transactions for account %d: %v", accountID, err)
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	logger.Info("Successfully retrieved %d transactions for account %d", len(transactions), accountID)
	return transactions, nil
}

// CreateTransactionWithTx creates a transaction record within a database transaction
func (r *PostgresTransactionRepository) CreateTransactionWithTx(ctx context.Context, tx *sql.Tx, transaction *models.Transaction) (*models.Transaction, error) {
	logger.Info("Creating transaction record in database: source=%d, destination=%d, amount=%s, status=%s",
		transaction.SourceAccountID, transaction.DestinationAccountID, transaction.Amount.String(), transaction.Status)

	// Validate transaction
	if err := transaction.Validate(); err != nil {
		logger.Warn("Transaction validation failed: %v", err)
		return nil, err
	}

	query := `
		INSERT INTO transactions (source_account_id, destination_account_id, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, source_account_id, destination_account_id, amount, status, created_at
	`

	var createdTx models.Transaction
	var createdAt time.Time
	err := tx.QueryRowContext(ctx, query,
		transaction.SourceAccountID,
		transaction.DestinationAccountID,
		transaction.Amount,
		transaction.Status,
		time.Now(),
	).Scan(
		&createdTx.ID,
		&createdTx.SourceAccountID,
		&createdTx.DestinationAccountID,
		&createdTx.Amount,
		&createdTx.Status,
		&createdAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation":
				logger.Warn("Foreign key violation creating transaction: source=%d, destination=%d",
					transaction.SourceAccountID, transaction.DestinationAccountID)
				return nil, errors.ErrAccountNotFound
			case "check_constraint_violation":
				logger.Warn("Check constraint violation creating transaction: amount=%s", transaction.Amount.String())
				return nil, errors.ErrInvalidAmount
			}
		}
		logger.Error("Database error creating transaction: %v", err)
		return nil, fmt.Errorf("failed to record transaction: %w", err)
	}

	createdTx.CreatedAt = createdAt.Format(time.RFC3339)

	logger.Info("Successfully created transaction record in database: id=%d, source=%d, destination=%d, amount=%s",
		createdTx.ID, createdTx.SourceAccountID, createdTx.DestinationAccountID, createdTx.Amount.String())
	return &createdTx, nil
}

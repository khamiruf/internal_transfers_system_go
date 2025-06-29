package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/khamiruf/internal_transfers_system_go/internal/errors"
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
	query := `
		SELECT id, source_account_id, destination_account_id, amount, status, created_at
		FROM transactions
		WHERE source_account_id = $1 OR destination_account_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
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
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		tx.CreatedAt = createdAt.Format(time.RFC3339)
		transactions = append(transactions, &tx)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	return transactions, nil
}

// CreateTransactionWithTx creates a transaction record within a database transaction
func (r *PostgresTransactionRepository) CreateTransactionWithTx(ctx context.Context, tx *sql.Tx, transaction *models.Transaction) error {
	// Validate transaction
	if err := transaction.Validate(); err != nil {
		return err
	}

	query := `
		INSERT INTO transactions (source_account_id, destination_account_id, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := tx.ExecContext(ctx, query,
		transaction.SourceAccountID,
		transaction.DestinationAccountID,
		transaction.Amount,
		transaction.Status,
		time.Now(),
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation":
				return errors.ErrAccountNotFound
			case "check_constraint_violation":
				return errors.ErrInvalidAmount
			}
		}
		return fmt.Errorf("failed to record transaction: %w", err)
	}

	return nil
}

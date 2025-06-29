package models

import (
	"github.com/khamiruf/internal_transfers_system_go/internal/errors"
	"github.com/shopspring/decimal"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending  TransactionStatus = "pending"
	TransactionStatusComplete TransactionStatus = "complete"
	TransactionStatusFailed   TransactionStatus = "failed"
)

// Transaction represents a financial transaction in the system
type Transaction struct {
	ID                   int64             `json:"id"`
	SourceAccountID      int64             `json:"source_account_id"`
	DestinationAccountID int64             `json:"destination_account_id"`
	Amount               decimal.Decimal   `json:"amount"`
	Status               TransactionStatus `json:"status"`
	CreatedAt            string            `json:"created_at"`
}

// Validate checks if the transaction is valid
func (t *Transaction) Validate() error {
	if t.Amount.LessThanOrEqual(decimal.Zero) {
		return errors.ErrInvalidAmount
	}
	if t.SourceAccountID == t.DestinationAccountID {
		return errors.ErrSameAccount
	}
	return nil
}

// IsComplete checks if the transaction is complete
func (t *Transaction) IsComplete() bool {
	return t.Status == TransactionStatusComplete
}

// IsFailed checks if the transaction failed
func (t *Transaction) IsFailed() bool {
	return t.Status == TransactionStatusFailed
}

// IsPending checks if the transaction is pending
func (t *Transaction) IsPending() bool {
	return t.Status == TransactionStatusPending
}

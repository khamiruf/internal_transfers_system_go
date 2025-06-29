package models

import (
	"github.com/khamiruf/internal_transfers_system_go/internal/errors"
	"github.com/shopspring/decimal"
)

// Account represents an account in the system
type Account struct {
	AccountID int64
	Balance   decimal.Decimal
	CreatedAt string
	UpdatedAt string
}

// HasSufficientBalance checks if the account has sufficient balance for a withdrawal
func (a *Account) HasSufficientBalance(amount decimal.Decimal) bool {
	return a.Balance.GreaterThanOrEqual(amount)
}

// Credit adds the specified amount to the account balance
func (a *Account) Credit(amount decimal.Decimal) {
	a.Balance = a.Balance.Add(amount)
}

// Debit subtracts the specified amount from the account balance
// Returns an error if insufficient balance
func (a *Account) Debit(amount decimal.Decimal) error {
	if !a.HasSufficientBalance(amount) {
		return errors.ErrInsufficientBalance
	}
	a.Balance = a.Balance.Sub(amount)
	return nil
}

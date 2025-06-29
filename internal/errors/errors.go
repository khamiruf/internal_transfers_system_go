// Package errors provides domain-specific errors for the internal transfers system
package errors

import "errors"

var (
	// ErrInsufficientBalance is returned when an account has insufficient balance for a debit operation
	ErrInsufficientBalance = errors.New("insufficient balance")

	// ErrAccountNotFound is returned when an account cannot be found
	ErrAccountNotFound = errors.New("account not found")

	// ErrSourceAccountNotFound is returned when the source account cannot be found
	ErrSourceAccountNotFound = errors.New("source account not found")

	// ErrDestinationAccountNotFound is returned when the destination account cannot be found
	ErrDestinationAccountNotFound = errors.New("destination account not found")

	// ErrAccountAlreadyExists is returned when trying to create an account that already exists
	ErrAccountAlreadyExists = errors.New("account already exists")

	// ErrInvalidAmount is returned when a transaction amount is invalid (zero or negative)
	ErrInvalidAmount = errors.New("invalid amount: must be greater than zero")

	// ErrSameAccount is returned when trying to transfer between the same account
	ErrSameAccount = errors.New("source and destination accounts must be different")

	// ErrDatabaseError is returned when a database operation fails
	ErrDatabaseError = errors.New("database operation failed")

	// ErrValidationFailed is returned when input validation fails
	ErrValidationFailed = errors.New("validation failed")
)

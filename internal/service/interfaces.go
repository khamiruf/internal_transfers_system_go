package service

import (
	"context"

	"github.com/khamiruf/internal_transfers_system_go/internal/api/dto"
	"github.com/khamiruf/internal_transfers_system_go/internal/models"
)

// AccountService defines the interface for account-related operations
type AccountService interface {
	CreateAccount(ctx context.Context, req *dto.CreateAccountRequest) error
	GetAccount(ctx context.Context, accountID int64) (*models.Account, error)
}

// TransactionService defines the interface for transaction-related operations
type TransactionService interface {
	CreateTransaction(ctx context.Context, req *dto.CreateTransactionRequest) error
}

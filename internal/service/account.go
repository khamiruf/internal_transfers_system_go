package service

import (
	"context"

	"github.com/khamiruf/internal_transfers_system_go/internal/api/dto"
	"github.com/khamiruf/internal_transfers_system_go/internal/errors"
	"github.com/khamiruf/internal_transfers_system_go/internal/models"
	"github.com/khamiruf/internal_transfers_system_go/internal/repository"
)

// accountService implements the AccountService interface
type accountService struct {
	repo repository.AccountRepository
}

// NewAccountService creates a new account service instance
func NewAccountService(repo repository.AccountRepository) AccountService {
	return &accountService{
		repo: repo,
	}
}

// CreateAccount creates a new account with validation
func (s *accountService) CreateAccount(ctx context.Context, req *dto.CreateAccountRequest) error {
	if req.InitialBalance.IsNegative() {
		return errors.ErrInvalidAmount
	}

	return s.repo.CreateAccount(ctx, req.AccountID, req.InitialBalance)
}

// GetAccount retrieves an account by its ID
func (s *accountService) GetAccount(ctx context.Context, accountID int64) (*models.Account, error) {
	return s.repo.GetAccount(ctx, accountID)
}

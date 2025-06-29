package service

import (
	"context"

	"github.com/khamiruf/internal_transfers_system_go/internal/api/dto"
	"github.com/khamiruf/internal_transfers_system_go/internal/errors"
	"github.com/khamiruf/internal_transfers_system_go/internal/logger"
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
	logger.Info("Creating account with ID: %d, initial balance: %s", req.AccountID, req.InitialBalance.String())

	if req.InitialBalance.IsNegative() {
		logger.Warn("Invalid initial balance for account %d: %s (negative amount)", req.AccountID, req.InitialBalance.String())
		return errors.ErrInvalidAmount
	}

	err := s.repo.CreateAccount(ctx, req.AccountID, req.InitialBalance)
	if err != nil {
		logger.Error("Failed to create account %d: %v", req.AccountID, err)
		return err
	}

	logger.Info("Successfully created account %d with initial balance %s", req.AccountID, req.InitialBalance.String())
	return nil
}

// GetAccount retrieves an account by its ID
func (s *accountService) GetAccount(ctx context.Context, accountID int64) (*models.Account, error) {
	logger.Info("Retrieving account: %d", accountID)

	account, err := s.repo.GetAccount(ctx, accountID)
	if err != nil {
		logger.Error("Failed to retrieve account %d: %v", accountID, err)
		return nil, err
	}

	logger.Info("Successfully retrieved account %d with balance %s", accountID, account.Balance.String())
	return account, nil
}

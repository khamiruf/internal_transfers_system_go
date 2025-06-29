package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/khamiruf/internal_transfers_system_go/internal/api/dto"
	domainErrors "github.com/khamiruf/internal_transfers_system_go/internal/errors"
	"github.com/khamiruf/internal_transfers_system_go/internal/logger"
	"github.com/khamiruf/internal_transfers_system_go/internal/models"
	"github.com/khamiruf/internal_transfers_system_go/internal/repository"
)

// transactionService implements the TransactionService interface
type transactionService struct {
	transactionRepo repository.TransactionRepository
	accountRepo     repository.AccountRepository
	db              *sql.DB
}

// NewTransactionService creates a new transaction service instance
func NewTransactionService(transactionRepo repository.TransactionRepository, accountRepo repository.AccountRepository, db *sql.DB) TransactionService {
	return &transactionService{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		db:              db,
	}
}

// withTransaction executes a function within a database transaction
func (s *transactionService) withTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	logger.Info("Starting database transaction")

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		logger.Error("Failed to start transaction: %v", err)
		return fmt.Errorf("error starting transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			logger.Error("Transaction panic occurred: %v", p)
			tx.Rollback()
			panic(p) // re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		logger.Error("Transaction failed, rolling back: %v", err)
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Error("Failed to rollback transaction: %v", rbErr)
			return fmt.Errorf("error rolling back transaction: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		logger.Error("Failed to commit transaction: %v", err)
		return fmt.Errorf("error committing transaction: %w", err)
	}

	logger.Info("Transaction committed successfully")
	return nil
}

// CreateTransaction processes a transaction between two accounts
func (s *transactionService) CreateTransaction(ctx context.Context, req *dto.CreateTransactionRequest) (*dto.TransactionResponse, error) {
	logger.Info("Processing transaction: source=%d, destination=%d, amount=%s",
		req.SourceAccountID, req.DestinationAccountID, req.Amount.String())

	transaction := &models.Transaction{
		SourceAccountID:      req.SourceAccountID,
		DestinationAccountID: req.DestinationAccountID,
		Amount:               req.Amount,
		Status:               models.TransactionStatusPending,
	}

	if err := transaction.Validate(); err != nil {
		logger.Warn("Transaction validation failed: %v", err)
		return nil, err
	}

	var createdTransaction *dto.TransactionResponse
	err := s.withTransaction(ctx, func(tx *sql.Tx) error {
		// Get source account
		logger.Info("Retrieving source account: %d", req.SourceAccountID)
		sourceAccount, err := s.accountRepo.GetAccountWithTx(ctx, tx, req.SourceAccountID)
		if err != nil {
			if errors.Is(err, domainErrors.ErrAccountNotFound) {
				logger.Warn("Source account not found: %d", req.SourceAccountID)
				return domainErrors.ErrSourceAccountNotFound
			}
			logger.Error("Failed to retrieve source account %d: %v", req.SourceAccountID, err)
			return err
		}

		logger.Info("Source account %d current balance: %s", req.SourceAccountID, sourceAccount.Balance.String())

		// Check sufficient balance
		if !sourceAccount.HasSufficientBalance(req.Amount) {
			logger.Warn("Insufficient balance: account=%d, current_balance=%s, required_amount=%s",
				req.SourceAccountID, sourceAccount.Balance.String(), req.Amount.String())
			return domainErrors.ErrInsufficientBalance
		}

		// Get destination account
		logger.Info("Retrieving destination account: %d", req.DestinationAccountID)
		destAccount, err := s.accountRepo.GetAccountWithTx(ctx, tx, req.DestinationAccountID)
		if err != nil {
			if errors.Is(err, domainErrors.ErrAccountNotFound) {
				logger.Warn("Destination account not found: %d", req.DestinationAccountID)
				return domainErrors.ErrDestinationAccountNotFound
			}
			logger.Error("Failed to retrieve destination account %d: %v", req.DestinationAccountID, err)
			return err
		}

		logger.Info("Destination account %d current balance: %s", req.DestinationAccountID, destAccount.Balance.String())

		// Calculate new balances
		sourceNewBalance := sourceAccount.Balance.Sub(req.Amount)
		destNewBalance := destAccount.Balance.Add(req.Amount)

		logger.Info("Updating source account %d balance: %s -> %s",
			req.SourceAccountID, sourceAccount.Balance.String(), sourceNewBalance.String())

		// Update source account balance
		err = s.accountRepo.UpdateBalanceWithTx(ctx, tx, req.SourceAccountID, sourceNewBalance)
		if err != nil {
			logger.Error("Failed to update source account %d balance: %v", req.SourceAccountID, err)
			return err
		}

		logger.Info("Updating destination account %d balance: %s -> %s",
			req.DestinationAccountID, destAccount.Balance.String(), destNewBalance.String())

		// Update destination account balance
		err = s.accountRepo.UpdateBalanceWithTx(ctx, tx, req.DestinationAccountID, destNewBalance)
		if err != nil {
			logger.Error("Failed to update destination account %d balance: %v", req.DestinationAccountID, err)
			return err
		}

		// Mark transaction as complete
		transaction.Status = models.TransactionStatusComplete

		logger.Info("Recording transaction: source=%d, destination=%d, amount=%s, status=%s",
			transaction.SourceAccountID, transaction.DestinationAccountID, transaction.Amount.String(), transaction.Status)

		// Record the transaction and get the created transaction with ID
		createdTx, err := s.transactionRepo.CreateTransactionWithTx(ctx, tx, transaction)
		if err != nil {
			logger.Error("Failed to record transaction: %v", err)
			return err
		}

		// Convert to response DTO
		createdTransaction = &dto.TransactionResponse{
			ID:                   createdTx.ID,
			SourceAccountID:      createdTx.SourceAccountID,
			DestinationAccountID: createdTx.DestinationAccountID,
			Amount:               createdTx.Amount,
			CreatedAt:            createdTx.CreatedAt,
		}

		logger.Info("Transaction completed successfully: id=%d, source=%d, destination=%d, amount=%s",
			createdTx.ID, req.SourceAccountID, req.DestinationAccountID, req.Amount.String())

		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdTransaction, nil
}

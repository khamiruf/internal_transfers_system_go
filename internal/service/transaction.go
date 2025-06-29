package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/khamiruf/internal_transfers_system_go/internal/api/dto"
	domainErrors "github.com/khamiruf/internal_transfers_system_go/internal/errors"
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
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("error rolling back transaction: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// CreateTransaction processes a transaction between two accounts
func (s *transactionService) CreateTransaction(ctx context.Context, req *dto.CreateTransactionRequest) error {
	transaction := &models.Transaction{
		SourceAccountID:      req.SourceAccountID,
		DestinationAccountID: req.DestinationAccountID,
		Amount:               req.Amount,
		Status:               models.TransactionStatusPending,
	}

	if err := transaction.Validate(); err != nil {
		return err
	}

	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		sourceAccount, err := s.accountRepo.GetAccountWithTx(ctx, tx, req.SourceAccountID)
		if err != nil {
			if errors.Is(err, domainErrors.ErrAccountNotFound) {
				return domainErrors.ErrSourceAccountNotFound
			}
			return err
		}

		if !sourceAccount.HasSufficientBalance(req.Amount) {
			return domainErrors.ErrInsufficientBalance
		}

		destAccount, err := s.accountRepo.GetAccountWithTx(ctx, tx, req.DestinationAccountID)
		if err != nil {
			if errors.Is(err, domainErrors.ErrAccountNotFound) {
				return domainErrors.ErrDestinationAccountNotFound
			}
			return err
		}

		sourceNewBalance := sourceAccount.Balance.Sub(req.Amount)
		destNewBalance := destAccount.Balance.Add(req.Amount)

		err = s.accountRepo.UpdateBalanceWithTx(ctx, tx, req.SourceAccountID, sourceNewBalance)
		if err != nil {
			return err
		}

		err = s.accountRepo.UpdateBalanceWithTx(ctx, tx, req.DestinationAccountID, destNewBalance)
		if err != nil {
			return err
		}

		transaction.Status = models.TransactionStatusComplete

		err = s.transactionRepo.CreateTransactionWithTx(ctx, tx, transaction)
		if err != nil {
			return err
		}

		return nil
	})
}

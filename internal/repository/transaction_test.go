package repository

import (
	"context"
	"testing"

	"github.com/khamiruf/internal_transfers_system_go/internal/errors"
	"github.com/khamiruf/internal_transfers_system_go/internal/models"
	"github.com/khamiruf/internal_transfers_system_go/internal/testutil"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestTransactionRepository_CreateTransactionWithTx(t *testing.T) {
	db := testutil.NewTestDB(t)
	testutil.SetupTestDB(t, db)
	defer testutil.CleanupTestDB(t, db)

	repo := NewTransactionRepository(db)
	accountRepo := NewAccountRepository(db)
	ctx := context.Background()

	// Create test accounts
	sourceID := int64(888888)
	destID := int64(888889)
	initialBalance := decimal.NewFromFloat(1000.00)

	err := accountRepo.CreateAccount(ctx, sourceID, initialBalance)
	assert.NoError(t, err)
	err = accountRepo.CreateAccount(ctx, destID, initialBalance)
	assert.NoError(t, err)

	tests := []struct {
		name          string
		transaction   *models.Transaction
		expectedError error
	}{
		{
			name: "valid transaction",
			transaction: &models.Transaction{
				SourceAccountID:      sourceID,
				DestinationAccountID: destID,
				Amount:               decimal.NewFromFloat(100.50),
			},
			expectedError: nil,
		},
		{
			name: "same source and destination",
			transaction: &models.Transaction{
				SourceAccountID:      sourceID,
				DestinationAccountID: sourceID,
				Amount:               decimal.NewFromFloat(100.50),
			},
			expectedError: errors.ErrSameAccount,
		},
		{
			name: "negative amount",
			transaction: &models.Transaction{
				SourceAccountID:      sourceID,
				DestinationAccountID: destID,
				Amount:               decimal.NewFromFloat(-100.50),
			},
			expectedError: errors.ErrInvalidAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start a database transaction
			tx, err := db.BeginTx(ctx, nil)
			assert.NoError(t, err)
			defer tx.Rollback()

			// Test the WithTx method
			createdTx, err := repo.CreateTransactionWithTx(ctx, tx, tt.transaction)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, createdTx)
				// Commit the transaction to persist changes
				err = tx.Commit()
				assert.NoError(t, err)

				// Verify transaction was created
				transactions, err := repo.GetTransactionsByAccount(ctx, tt.transaction.SourceAccountID)
				assert.NoError(t, err)
				assert.NotEmpty(t, transactions)
				assert.Equal(t, tt.transaction.SourceAccountID, transactions[0].SourceAccountID)
				assert.Equal(t, tt.transaction.DestinationAccountID, transactions[0].DestinationAccountID)
				assert.True(t, tt.transaction.Amount.Equal(transactions[0].Amount))
			}
		})
	}
}

func TestTransactionRepository_GetTransactionsByAccount(t *testing.T) {
	db := testutil.NewTestDB(t)
	testutil.SetupTestDB(t, db)
	defer testutil.CleanupTestDB(t, db)

	repo := NewTransactionRepository(db)
	accountRepo := NewAccountRepository(db)
	ctx := context.Background()

	// Create test accounts
	sourceID := int64(777777)
	destID := int64(777778)
	initialBalance := decimal.NewFromFloat(1000.00)

	err := accountRepo.CreateAccount(ctx, sourceID, initialBalance)
	assert.NoError(t, err)
	err = accountRepo.CreateAccount(ctx, destID, initialBalance)
	assert.NoError(t, err)

	// Create a transaction using WithTx method
	tx, err := db.BeginTx(ctx, nil)
	assert.NoError(t, err)

	transaction := &models.Transaction{
		SourceAccountID:      sourceID,
		DestinationAccountID: destID,
		Amount:               decimal.NewFromFloat(100.50),
	}
	_, err = repo.CreateTransactionWithTx(ctx, tx, transaction)
	assert.NoError(t, err)

	err = tx.Commit()
	assert.NoError(t, err)

	tests := []struct {
		name          string
		accountID     int64
		expectedCount int
	}{
		{
			name:          "account with transactions",
			accountID:     sourceID,
			expectedCount: 1,
		},
		{
			name:          "account without transactions",
			accountID:     int64(999999),
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transactions, err := repo.GetTransactionsByAccount(ctx, tt.accountID)
			assert.NoError(t, err)
			assert.Len(t, transactions, tt.expectedCount)
		})
	}
}

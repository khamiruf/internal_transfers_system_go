package repository

import (
	"context"
	"testing"

	"github.com/khamiruf/internal_transfers_system_go/internal/errors"
	"github.com/khamiruf/internal_transfers_system_go/internal/testutil"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestAccountRepository_CreateAccount(t *testing.T) {
	db := testutil.NewTestDB(t)
	testutil.SetupTestDB(t, db)
	defer testutil.CleanupTestDB(t, db)

	repo := NewAccountRepository(db)
	ctx := context.Background()

	// First create an account that we'll use for duplicate testing
	err := repo.CreateAccount(ctx, 1, decimal.NewFromFloat(100.00))
	assert.NoError(t, err)

	tests := []struct {
		name          string
		accountID     int64
		balance       decimal.Decimal
		expectedError error
	}{
		{
			name:          "valid account creation",
			accountID:     999999, // Using a very high number to avoid conflicts
			balance:       decimal.NewFromFloat(100.50),
			expectedError: nil,
		},
		{
			name:          "duplicate account",
			accountID:     1, // This account was created above
			balance:       decimal.NewFromFloat(100.50),
			expectedError: errors.ErrAccountAlreadyExists,
		},
		{
			name:          "negative balance",
			accountID:     3,
			balance:       decimal.NewFromFloat(-100.50),
			expectedError: errors.ErrInvalidAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreateAccount(ctx, tt.accountID, tt.balance)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				// Verify account was created
				account, err := repo.GetAccount(ctx, tt.accountID)
				assert.NoError(t, err)
				assert.Equal(t, tt.accountID, account.AccountID)
				assert.True(t, tt.balance.Equal(account.Balance))
			}
		})
	}
}

func TestAccountRepository_GetAccount(t *testing.T) {
	db := testutil.NewTestDB(t)
	testutil.SetupTestDB(t, db)
	defer testutil.CleanupTestDB(t, db)

	repo := NewAccountRepository(db)
	ctx := context.Background()

	// Create a test account
	testAccountID := int64(888888)
	initialBalance := decimal.NewFromFloat(100.00)
	err := repo.CreateAccount(ctx, testAccountID, initialBalance)
	assert.NoError(t, err)

	tests := []struct {
		name          string
		accountID     int64
		expectedError error
	}{
		{
			name:          "existing account",
			accountID:     testAccountID,
			expectedError: nil,
		},
		{
			name:          "non-existent account",
			accountID:     999999,
			expectedError: errors.ErrAccountNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, err := repo.GetAccount(ctx, tt.accountID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, account)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, account)
				assert.Equal(t, tt.accountID, account.AccountID)
				assert.True(t, initialBalance.Equal(account.Balance))
			}
		})
	}
}

func TestAccountRepository_UpdateBalanceWithTx(t *testing.T) {
	db := testutil.NewTestDB(t)
	testutil.SetupTestDB(t, db)
	defer testutil.CleanupTestDB(t, db)

	repo := NewAccountRepository(db)
	ctx := context.Background()

	// Create a test account
	testAccountID := int64(888888)
	initialBalance := decimal.NewFromFloat(100.00)
	err := repo.CreateAccount(ctx, testAccountID, initialBalance)
	assert.NoError(t, err)

	tests := []struct {
		name          string
		accountID     int64
		newBalance    decimal.Decimal
		expectedError error
	}{
		{
			name:          "valid balance update",
			accountID:     testAccountID,
			newBalance:    decimal.NewFromFloat(200.50),
			expectedError: nil,
		},
		{
			name:          "negative balance",
			accountID:     testAccountID,
			newBalance:    decimal.NewFromFloat(-100.50),
			expectedError: errors.ErrInvalidAmount,
		},
		{
			name:          "non-existent account",
			accountID:     999999,
			newBalance:    decimal.NewFromFloat(100.50),
			expectedError: errors.ErrAccountNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start a database transaction
			tx, err := db.BeginTx(ctx, nil)
			assert.NoError(t, err)
			defer tx.Rollback()

			// Test the WithTx method
			err = repo.UpdateBalanceWithTx(ctx, tx, tt.accountID, tt.newBalance)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				// Commit the transaction to persist changes
				err = tx.Commit()
				assert.NoError(t, err)

				// Verify balance was updated
				account, err := repo.GetAccount(ctx, tt.accountID)
				assert.NoError(t, err)
				assert.True(t, tt.newBalance.Equal(account.Balance))
			}
		})
	}
}

func TestAccountRepository_DecimalPrecision(t *testing.T) {
	db := testutil.NewTestDB(t)
	testutil.SetupTestDB(t, db)
	defer testutil.CleanupTestDB(t, db)

	repo := NewAccountRepository(db)
	ctx := context.Background()

	// Test with precise decimal values that might cause issues
	testCases := []struct {
		name      string
		accountID int64
		balance   decimal.Decimal
	}{
		{
			name:      "simple decimal",
			accountID: 1001,
			balance:   decimal.NewFromFloat(1000.50),
		},
		{
			name:      "large decimal",
			accountID: 1002,
			balance:   decimal.NewFromFloat(999999.99),
		},
		{
			name:      "small decimal",
			accountID: 1003,
			balance:   decimal.NewFromFloat(0.01),
		},
		{
			name:      "zero decimal",
			accountID: 1004,
			balance:   decimal.Zero,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create account
			err := repo.CreateAccount(ctx, tc.accountID, tc.balance)
			assert.NoError(t, err)

			// Verify via repository
			account, err := repo.GetAccount(ctx, tc.accountID)
			assert.NoError(t, err)
			assert.True(t, tc.balance.Equal(account.Balance),
				"Expected balance %s, got %s", tc.balance.String(), account.Balance.String())

			// Verify via direct database query - PostgreSQL DECIMAL(20,5) will show 5 decimal places
			var balanceStr string
			err = db.QueryRow("SELECT balance FROM accounts WHERE account_id = $1", tc.accountID).Scan(&balanceStr)
			assert.NoError(t, err)

			// Parse the stored string back to decimal and verify it equals the original
			storedBalance, err := decimal.NewFromString(balanceStr)
			assert.NoError(t, err)
			assert.True(t, tc.balance.Equal(storedBalance),
				"Parsed balance %s doesn't match original %s", storedBalance.String(), tc.balance.String())
		})
	}
}

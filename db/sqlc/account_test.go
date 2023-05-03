package db

import (
	"context"
	"database/sql"
	"github.com/gaggudeep/bank_go/util"
	"github.com/stretchr/testify/require"
	"testing"
)

func createRandomAccount(t *testing.T) *Account {
	arg := CreateAccountParams{
		OwnerName: util.RandomOwnerName(),
		Balance:   util.RandomMoney(),
		Currency:  util.RandomCurrency(),
	}

	acc, err := testQueries.CreateAccount(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, acc)
	require.Equal(t, arg.OwnerName, acc.OwnerName)
	require.Equal(t, arg.Balance, acc.Balance)
	require.Equal(t, arg.Currency, acc.Currency)
	require.NotZero(t, acc.ID)
	require.NotZero(t, acc.CreatedAt)

	return &acc
}

func TestGetAccount(t *testing.T) {
	acc := *createRandomAccount(t)
	acc2, err := testQueries.GetAccount(context.Background(), acc.ID)

	require.NoError(t, err)
	require.NotEmpty(t, acc2)
	require.Equal(t, acc, acc2)
}

func TestAddToAccountBalance(t *testing.T) {
	acc := *createRandomAccount(t)

	arg := AddToAccountBalanceParams{
		ID:     acc.ID,
		Amount: util.RandomMoney(),
	}

	updatedAcc, err := testQueries.AddToAccountBalance(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, updatedAcc)
	require.Equal(t, acc.ID, updatedAcc.ID)
	require.Equal(t, acc.OwnerName, updatedAcc.OwnerName)
	require.Equal(t, acc.CreatedAt, updatedAcc.CreatedAt)
	require.Equal(t, acc.Currency, updatedAcc.Currency)
}

func TestDeleteAccountIfUserExists(t *testing.T) {
	acc := *createRandomAccount(t)

	err := testQueries.DeleteAccount(context.Background(), acc.ID)

	require.NoError(t, err)
}

func TestGetAccountIfUserDoesNotExists(t *testing.T) {
	acc, err := testQueries.GetAccount(context.Background(), -1)

	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, acc)
}

func TestGetAccounts(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomAccount(t)
	}

	arg := GetAccountsParams{
		Limit:  5,
		Offset: 5,
	}

	accounts, err := testQueries.GetAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, accounts, 5)

	for _, acc := range accounts {
		require.NotEmpty(t, acc)
	}
}

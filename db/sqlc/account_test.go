package db

import (
	"context"
	"database/sql"
	"github.com/gaggudeep/bank_go/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomAccount(t *testing.T) *Account {
	user := createRandomUser(t)
	arg := CreateAccountParams{
		OwnerName: user.Username,
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
	require.Equal(t, acc.ID, acc2.ID)
	require.Equal(t, acc.OwnerName, acc2.OwnerName)
	require.Equal(t, acc.Balance, acc2.Balance)
	require.Equal(t, acc.Currency, acc2.Currency)
	require.WithinDuration(t, acc.CreatedAt, acc2.CreatedAt, time.Second)
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
	var lastAcc *Account
	for i := 0; i < 10; i++ {
		lastAcc = createRandomAccount(t)
	}

	arg := GetAccountsParams{
		OwnerName: lastAcc.OwnerName,
		Limit:     5,
		Offset:    0,
	}

	accounts, err := testQueries.GetAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, accounts)

	for _, acc := range accounts {
		require.NotEmpty(t, acc)
		require.Equal(t, lastAcc.OwnerName, acc.OwnerName)
	}
}

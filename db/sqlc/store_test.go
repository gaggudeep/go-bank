package db

import (
	"context"
	"github.com/stretchr/testify/require"
	"log"
	"math/big"
	"testing"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	fromAcc := *createRandomAccount(t)
	toAcc := *createRandomAccount(t)

	n := 5
	amt := "10.36"
	negativeAmt := "-10.36"

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTxPreventingCircularWait(context.Background(), TransferTxParams{
				FromAccountID: fromAcc.ID,
				ToAccountID:   toAcc.ID,
				Amount:        amt,
			})
			errs <- err
			results <- result
		}()
	}

	txs := make(map[int]bool)
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, toAcc.ID, transfer.ToAccountID)
		require.Equal(t, fromAcc.ID, transfer.FromAccountID)
		require.Equal(t, amt, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		fromTx := result.FromTransaction
		require.NotEmpty(t, fromTx)
		require.Equal(t, fromAcc.ID, fromTx.AccountID)
		require.Equal(t, negativeAmt, fromTx.Amount)
		require.NotZero(t, fromTx.ID)
		require.NotZero(t, fromTx.CreatedAt)

		_, err = store.GetTransaction(context.Background(), fromTx.ID)
		require.NoError(t, err)

		toTx := result.ToTransaction
		require.NotEmpty(t, toTx)
		require.Equal(t, toAcc.ID, toTx.AccountID)
		require.Equal(t, amt, toTx.Amount)
		require.NotZero(t, toTx.ID)
		require.NotZero(t, toTx.CreatedAt)

		_, err = store.GetTransaction(context.Background(), toTx.ID)
		require.NoError(t, err)

		resFromAcc := result.FromAccount
		require.NotEmpty(t, resFromAcc)
		require.Equal(t, fromAcc.ID, resFromAcc.ID)

		resToAcc := result.ToAccount
		require.NotEmpty(t, resToAcc)
		require.Equal(t, toAcc.ID, resToAcc.ID)

		ratFromAccBal := toRat(t, fromAcc.Balance)
		ratToAccBal := toRat(t, toAcc.Balance)

		fromAccTransferredAmt := ratFromAccBal.Sub(ratFromAccBal, toRat(t, resFromAcc.Balance))
		toAccTransferredAmt := ratToAccBal.Sub(toRat(t, resToAcc.Balance), ratToAccBal)
		require.Equal(t, fromAccTransferredAmt.String(), toAccTransferredAmt.String())
		require.True(t, fromAccTransferredAmt.Sign() == 1)

		quo := fromAccTransferredAmt.Quo(fromAccTransferredAmt, toRat(t, amt))
		require.True(t, quo.IsInt())
		require.True(t, quo.Cmp(big.NewRat(1, 1)) >= 0)
		require.True(t, quo.Cmp(big.NewRat(int64(n), 1)) <= 0)

		txNum, exact := quo.Float64()
		if exact == false {
			log.Printf("not exact conversioj from %v to %f", quo, txNum)
		}
		require.NotContains(t, txs, txNum)
		txs[int(txNum)] = true
	}
}

func TestTransferTxDeadlock(t *testing.T) {
	store := NewStore(testDB)

	acc1 := *createRandomAccount(t)
	acc2 := *createRandomAccount(t)

	n := 10
	amt := "10.36"

	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccId := acc1.ID
		toAccId := acc2.ID

		if i%2 == 0 {
			fromAccId = acc2.ID
			toAccId = acc1.ID
		}

		go func() {
			_, err := store.TransferTxPreventingCircularWait(context.Background(), TransferTxParams{
				FromAccountID: fromAccId,
				ToAccountID:   toAccId,
				Amount:        amt,
			})
			errs <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	updatedAcc1, err := store.GetAccount(context.Background(), acc1.ID)
	require.NoError(t, err)
	require.Equal(t, acc1.Balance, updatedAcc1.Balance)

	updatedAcc2, err := store.GetAccount(context.Background(), acc2.ID)
	require.NoError(t, err)
	require.Equal(t, acc2.Balance, updatedAcc2.Balance)
}

func toRat(t *testing.T, val string) *big.Rat {
	ratVal, success := big.NewRat(1, 1).SetString(val)
	require.True(t, success)

	return ratVal
}

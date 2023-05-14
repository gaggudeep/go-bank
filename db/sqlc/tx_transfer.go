package db

import (
	"context"
	"strconv"
)

type TransferTxParams struct {
	FromAccountID int64  `json:"from_account_id"`
	ToAccountID   int64  `json:"to_account_id"`
	Amount        string `json:"amount"`
}

type TransferTxResult struct {
	Transfer        Transfer    `json:"transfer"`
	FromAccount     Account     `json:"from_account"`
	ToAccount       Account     `json:"to_account"`
	FromTransaction Transaction `json:"from_transaction"`
	ToTransaction   Transaction `json:"to_transaction"`
}

func transferMoney(ctx context.Context, q *Queries, accID1 *int64, accID2 *int64,
	amt1 string, amt2 string) (acc1 Account, acc2 Account, err error) {
	acc1, err = q.AddToAccountBalance(ctx, AddToAccountBalanceParams{
		ID:     *accID1,
		Amount: amt1,
	})
	if err != nil {
		return
	}

	acc2, err = q.AddToAccountBalance(ctx, AddToAccountBalanceParams{
		ID:     *accID2,
		Amount: amt2,
	})
	if err != nil {
		return
	}

	return
}

func (store *SQLStore) TransferTxPreventingCircularWait(ctx context.Context,
	arg TransferTxParams) (TransferTxResult, error) {
	var res TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		res.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		negatedAmtFloat, err := strconv.ParseFloat(arg.Amount, 64)
		if err != nil {
			return err
		}
		negatedAmt := strconv.FormatFloat(-negatedAmtFloat, 'f', -1, 64)

		res.FromTransaction, err = q.CreateTransaction(ctx, CreateTransactionParams{
			AccountID: arg.FromAccountID,
			Amount:    negatedAmt,
		})
		if err != nil {
			return err
		}

		res.ToTransaction, err = q.CreateTransaction(ctx, CreateTransactionParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		if arg.FromAccountID < arg.ToAccountID {
			res.FromAccount, res.ToAccount, err = transferMoney(
				ctx, q, &arg.FromAccountID, &arg.ToAccountID, negatedAmt, arg.Amount)
			if err != nil {
				return err
			}
		} else {
			res.ToAccount, res.FromAccount, err = transferMoney(
				ctx, q, &arg.ToAccountID, &arg.FromAccountID, arg.Amount, negatedAmt)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return res, err
}

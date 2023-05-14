package db

import (
	"context"
)

type CreateUserTxParams struct {
	CreateUserParams
	AfterCreate func(User) error
}

type CreateUserTxResult struct {
	User User
}

func (store *SQLStore) CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error) {
	var res CreateUserTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		res.User, err = q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			return err
		}

		return arg.AfterCreate(res.User)
	})

	return res, err
}

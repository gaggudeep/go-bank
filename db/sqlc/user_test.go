package db

import (
	"context"
	"database/sql"
	"github.com/gaggudeep/bank_go/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomUser(t *testing.T) *User {
	hashedPwd, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	arg := CreateUserParams{
		Username:       util.RandomOwnerName(),
		HashedPassword: hashedPwd,
		Name:           util.RandomOwnerName(),
		Email:          util.RandomEmail(),
	}

	user, err := testQueries.CreateUser(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, arg.Name, user.Name)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.Username, user.Username)
	require.NotZero(t, user.CreatedAt)
	require.True(t, user.PasswordChangedAt.IsZero())

	return &user
}

func TestGetUser(t *testing.T) {
	user := *createRandomUser(t)
	user2, err := testQueries.GetUser(context.Background(), user.Username)

	require.NoError(t, err)
	require.NotEmpty(t, user2)
	require.Equal(t, user.Username, user2.Username)
	require.Equal(t, user.Name, user2.Name)
	require.Equal(t, user.HashedPassword, user2.HashedPassword)
	require.Equal(t, user.Email, user2.Email)
	require.WithinDuration(t, user.CreatedAt, user2.CreatedAt, time.Second)
	require.WithinDuration(t, user.PasswordChangedAt, user2.PasswordChangedAt, time.Second)
}

func TestUpdateUserOnlyName(t *testing.T) {
	oldUser := createRandomUser(t)
	newName := util.RandomOwnerName()

	updatedUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		Name: sql.NullString{
			String: newName,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEqual(t, oldUser.Name, updatedUser.Name)
	require.Equal(t, newName, updatedUser.Name)
	require.Equal(t, oldUser.Email, updatedUser.Email)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, oldUser.CreatedAt, updatedUser.CreatedAt)
	require.Equal(t, oldUser.PasswordChangedAt, updatedUser.PasswordChangedAt)
}

func TestUpdateUserNameAllFields(t *testing.T) {
	oldUser := createRandomUser(t)
	newName := util.RandomOwnerName()
	newEmail := util.RandomEmail()
	newPassword := util.RandomString(6)
	newHashedPassword, err := util.HashPassword(newPassword)
	require.NoError(t, err)

	updatedUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		Name: sql.NullString{
			String: newName,
			Valid:  true,
		},
		Email: sql.NullString{
			String: newEmail,
			Valid:  true,
		},
		HashedPassword: sql.NullString{
			String: newHashedPassword,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.Equal(t, newName, updatedUser.Name)
	require.Equal(t, newEmail, updatedUser.Email)
	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)
	require.Equal(t, oldUser.CreatedAt, updatedUser.CreatedAt)
	require.WithinDuration(t, updatedUser.PasswordChangedAt, time.Now(), time.Second)
	require.NotEqual(t, oldUser.PasswordChangedAt, updatedUser.PasswordChangedAt)
	require.NotEqual(t, oldUser.Name, updatedUser.Name)
	require.NotEqual(t, oldUser.Email, updatedUser.Email)
	require.NotEqual(t, oldUser.HashedPassword, updatedUser.HashedPassword)
}

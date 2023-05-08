package util

import (
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestPassword(t *testing.T) {
	pwd := RandomString(6)

	hashedPwd1, err := HashPassword(pwd)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPwd1)

	err = ValidatePassword(pwd, hashedPwd1)
	require.NoError(t, err)

	wrongPwd := RandomString(6)
	err = ValidatePassword(wrongPwd, hashedPwd1)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

	hashedPwd2, err := HashPassword(pwd)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPwd2)

	err = ValidatePassword(pwd, hashedPwd2)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPwd2)
	require.NotEqual(t, hashedPwd1, hashedPwd2)

}

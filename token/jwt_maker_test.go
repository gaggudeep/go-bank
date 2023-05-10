package token

import (
	"fmt"
	"github.com/gaggudeep/bank_go/util"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var maker Maker

func init() {
	var err error
	maker, err = NewJWTMaker(util.RandomString(32))
	if err != nil {
		panic(fmt.Sprintf("error creating maker: %v", err))
	}
}

func TestJWTMaker(t *testing.T) {
	username := util.RandomOwnerName()
	duration := time.Minute
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	token, payload, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

func TestExpiredJWTToken(t *testing.T) {
	token, payload, err := maker.CreateToken(util.RandomOwnerName(), -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, jwt.ErrTokenInvalidClaims.Error()+
		": "+jwt.ErrTokenExpired.Error())
	require.Nil(t, payload)
}

func TestInvalidHWTTokenAlgoNone(t *testing.T) {
	payload, err := NewPayload(util.RandomOwnerName(), time.Minute)
	require.NoError(t, err)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
	token, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	payload, err = maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, jwt.ErrTokenUnverifiable.Error()+
		": error while executing keyfunc: "+ErrInvalidToken.Error())
	require.Nil(t, payload)
}

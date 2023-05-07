package api

import (
	db "github.com/gaggudeep/bank_go/db/sqlc"
	"github.com/gaggudeep/bank_go/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func newTestServer(t *testing.T, store db.Store) *Server {
	config := &util.Config{
		SecurityConfig: util.SecurityConfig{
			TokenConfig: util.TokenConfig{
				SymmetricKey:   util.RandomString(32),
				AccessDuration: time.Minute,
			},
		},
		CustomValidators: util.CustomValidators,
	}

	server, err := NewServer(store, config)
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}

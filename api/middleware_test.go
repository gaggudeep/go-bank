package api

import (
	"fmt"
	"github.com/gaggudeep/bank_go/token"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthMiddleware(t *testing.T) {
	testCases := []struct {
		name      string
		setupAuth func(*http.Request, token.Maker)
		checkResp func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, "user", time.Minute)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "NoAuth",
			setupAuth: func(req *http.Request, maker token.Maker) {
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name: "UnSupportedAuthScheme",
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, "unsupported", "user", time.Minute)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name: "InvalidAuthFormat",
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, "", "user", time.Minute)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, "user", -time.Minute)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
	}

	server := newTestServer(t, nil)
	authPath := "/auth"
	server.router.GET(
		authPath,
		authMiddleware(server.tokenMaker),
		func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{})
		},
	)

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)

			tc.setupAuth(req, server.tokenMaker)
			server.router.ServeHTTP(rec, req)
			tc.checkResp(rec)
		})
	}
}

func addAuthorization(t *testing.T, req *http.Request, maker token.Maker,
	authScheme string, username string, duration time.Duration) {
	token, err := maker.CreateToken(username, duration)
	require.NoError(t, err)

	authHeader := fmt.Sprintf("%s %s", authScheme, token)
	req.Header.Set(authorizationHeaderKey, authHeader)
}

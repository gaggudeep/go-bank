package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	mockdb "github.com/gaggudeep/bank_go/db/mock"
	db "github.com/gaggudeep/bank_go/db/sqlc"
	"github.com/gaggudeep/bank_go/token"
	"github.com/gaggudeep/bank_go/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func randomAccount(ownerName string) db.Account {
	return db.Account{
		ID:        int64(util.RandomFloat(1, 1000)),
		OwnerName: ownerName,
		Balance:   util.RandomMoney(),
		Currency:  util.RandomCurrency(),
	}
}

func TestGetAccount(t *testing.T) {
	user, _ := randomUser(t)
	acc := randomAccount(user.Username)

	testCases := []struct {
		name       string
		accId      int64
		setupAuth  func(*http.Request, token.Maker)
		buildStubs func(*mockdb.MockStore)
		checkResp  func(*httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			accId: acc.ID,
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc.ID)).
					Times(1).
					Return(acc, nil)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rec.Code)
				requireBodyMatchAccount(t, rec.Body, &acc)
			},
		},
		{
			name:  "UnauthorizedUser",
			accId: acc.ID,
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, "unauthorized_user", time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(acc.ID)).Times(1).Return(acc, nil)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name:  "NoAuthorization",
			accId: acc.ID,
			setupAuth: func(req *http.Request, maker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name:  "NotFound",
			accId: acc.ID,
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, rec.Code)
			},
		},
		{
			name:  "InternalError",
			accId: acc.ID,
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, rec.Code)
			},
		},
		{
			name:  "InvalidID",
			accId: 0,
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, store)

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs(store)

			rec := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accId)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(req, server.tokenMaker)
			server.router.ServeHTTP(rec, req)
			tc.checkResp(rec)
		})
	}
}

func TestCreateAccount(t *testing.T) {
	user, _ := randomUser(t)
	acc := randomAccount(user.Username)

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(*http.Request, token.Maker)
		buildStubs    func(*mockdb.MockStore)
		checkResponse func(*httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"currency": acc.Currency,
			},
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateAccountParams{
					OwnerName: acc.OwnerName,
					Currency:  acc.Currency,
					Balance:   "0",
				}

				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(acc, nil)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rec.Code)
				requireBodyMatchAccount(t, rec.Body, &acc)
			},
		},
		{
			name: "NoAuthorization",
			body: gin.H{
				"currency": acc.Currency,
			},
			setupAuth: func(req *http.Request, maker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"currency": acc.Currency,
			},
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, rec.Code)
			},
		},
		{
			name: "InvalidCurrency",
			body: gin.H{
				"currency": "invalid",
			},
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, store)

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs(store)

			rec := httptest.NewRecorder()
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(req, server.tokenMaker)
			server.router.ServeHTTP(rec, req)
			tc.checkResponse(rec)
		})
	}
}

func TestGetAccounts(t *testing.T) {
	user, _ := randomUser(t)

	n := 5
	accounts := make([]db.Account, n)
	for i := 0; i < n; i++ {
		accounts[i] = randomAccount(user.Username)
	}

	type Query struct {
		page     int
		pageSize int
	}

	testCases := []struct {
		name          string
		query         Query
		setupAuth     func(*http.Request, token.Maker)
		buildStubs    func(*mockdb.MockStore)
		checkResponse func(*httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				page:     1,
				pageSize: n,
			},
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.GetAccountsParams{
					OwnerName: user.Username,
					Limit:     int32(n),
					Offset:    0,
				}

				store.EXPECT().GetAccounts(gomock.Any(), gomock.Eq(arg)).Times(1).Return(accounts, nil)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rec.Code)
				requireBodyMatchAccounts(t, rec.Body, accounts)
			},
		},
		{
			name: "NoAuthorization",
			query: Query{
				page:     1,
				pageSize: n,
			},
			setupAuth: func(req *http.Request, maker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccounts(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name: "InternalError",
			query: Query{
				page:     1,
				pageSize: n,
			},
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccounts(gomock.Any(), gomock.Any()).
					Times(1).Return([]db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, rec.Code)
			},
		},
		{
			name: "InvalidPage",
			query: Query{
				page:     -1,
				pageSize: n,
			},
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccounts(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name: "InvalidPageSize",
			query: Query{
				page:     1,
				pageSize: 100000,
			},
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccounts(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, store)

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs(store)

			rec := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, "/accounts", nil)
			require.NoError(t, err)

			q := req.URL.Query()
			q.Add("page", fmt.Sprintf("%d", tc.query.page))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			req.URL.RawQuery = q.Encode()

			tc.setupAuth(req, server.tokenMaker)
			server.router.ServeHTTP(rec, req)
			tc.checkResponse(rec)
		})
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, acc *db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var actualAcc db.Account
	err = json.Unmarshal(data, &actualAcc)
	require.NoError(t, err)
	require.Equal(t, *acc, actualAcc)
}

func requireBodyMatchAccounts(t *testing.T, body *bytes.Buffer, accounts []db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccounts []db.Account
	err = json.Unmarshal(data, &gotAccounts)
	require.NoError(t, err)
	require.Equal(t, accounts, gotAccounts)
}

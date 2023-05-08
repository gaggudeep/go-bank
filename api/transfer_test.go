package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	mockdb "github.com/gaggudeep/bank_go/db/mock"
	db "github.com/gaggudeep/bank_go/db/sqlc"
	"github.com/gaggudeep/bank_go/token"
	"github.com/gaggudeep/bank_go/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTransfer(t *testing.T) {
	amt := "10"
	negAmt := "-10"
	user1, _ := randomUser(t)
	user2, _ := randomUser(t)
	user3, _ := randomUser(t)
	acc1 := randomAccount(user1.Username)
	acc2 := randomAccount(user2.Username)
	acc3 := randomAccount(user3.Username)
	acc1.Currency = util.USD
	acc2.Currency = util.USD
	acc3.Currency = util.EUR

	testCases := []struct {
		name       string
		body       gin.H
		setupAuth  func(*http.Request, token.Maker)
		buildStubs func(*mockdb.MockStore)
		checkResp  func(*httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"from_account_id": acc1.ID,
				"to_account_id":   acc2.ID,
				"amount":          amt,
				"currency":        util.USD,
			},
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(acc1.ID)).Times(1).Return(acc1, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(acc2.ID)).Times(1).Return(acc2, nil)

				arg := db.TransferTxParams{
					FromAccountID: acc1.ID,
					ToAccountID:   acc2.ID,
					Amount:        amt,
				}

				store.EXPECT().
					TransferTxPreventingCircularWait(gomock.Any(), gomock.Eq(arg)).
					Times(1)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "NoAuthorization",
			body: gin.H{
				"from_account_id": acc1.ID,
				"to_account_id":   acc2.ID,
				"amount":          amt,
				"currency":        util.USD,
			},
			setupAuth: func(req *http.Request, maker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTxPreventingCircularWait(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name: "FromAccountNotFound",
			body: gin.H{
				"from_account_id": acc1.ID,
				"to_account_id":   acc2.ID,
				"amount":          amt,
				"currency":        util.USD,
			},
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(acc1.ID)).
					Times(1).Return(db.Account{}, sql.ErrNoRows)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(acc2.ID)).Times(0)
				store.EXPECT().TransferTxPreventingCircularWait(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, rec.Code)
			},
		},
		{
			name: "ToAccountNotFound",
			body: gin.H{
				"from_account_id": acc1.ID,
				"to_account_id":   acc2.ID,
				"amount":          amt,
				"currency":        util.USD,
			},
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(acc1.ID)).Times(1).Return(acc1, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(acc2.ID)).
					Times(1).Return(db.Account{}, sql.ErrNoRows)
				store.EXPECT().TransferTxPreventingCircularWait(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, rec.Code)
			},
		},
		{
			name: "FromAccountCurrencyMismatch",
			body: gin.H{
				"from_account_id": acc3.ID,
				"to_account_id":   acc2.ID,
				"amount":          amt,
				"currency":        util.USD,
			},
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(acc3.ID)).Times(1).Return(acc3, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(acc2.ID)).Times(0)
				store.EXPECT().TransferTxPreventingCircularWait(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name: "ToAccountCurrencyMismatch",
			body: gin.H{
				"from_account_id": acc1.ID,
				"to_account_id":   acc3.ID,
				"amount":          amt,
				"currency":        util.USD,
			},
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(acc1.ID)).Times(1).Return(acc1, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(acc3.ID)).Times(1).Return(acc3, nil)
				store.EXPECT().TransferTxPreventingCircularWait(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name: "InvalidCurrency",
			body: gin.H{
				"from_account_id": acc1.ID,
				"to_account_id":   acc2.ID,
				"amount":          amt,
				"currency":        "XYZ",
			},
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTxPreventingCircularWait(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name: "NegativeAmount",
			body: gin.H{
				"from_account_id": acc1.ID,
				"to_account_id":   acc2.ID,
				"amount":          negAmt,
				"currency":        util.USD,
			},
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTxPreventingCircularWait(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name: "GetAccountError",
			body: gin.H{
				"from_account_id": acc1.ID,
				"to_account_id":   acc2.ID,
				"amount":          amt,
				"currency":        util.USD,
			},
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(1).Return(db.Account{}, sql.ErrConnDone)
				store.EXPECT().TransferTxPreventingCircularWait(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, rec.Code)
			},
		},
		{
			name: "TransferTxError",
			body: gin.H{
				"from_account_id": acc1.ID,
				"to_account_id":   acc2.ID,
				"amount":          amt,
				"currency":        util.USD,
			},
			setupAuth: func(req *http.Request, maker token.Maker) {
				addAuthorization(t, req, maker, authorizationSchemeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(acc1.ID)).Times(1).Return(acc1, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(acc2.ID)).Times(1).Return(acc2, nil)
				store.EXPECT().TransferTxPreventingCircularWait(gomock.Any(), gomock.Any()).
					Times(1).Return(db.TransferTxResult{}, sql.ErrTxDone)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, rec.Code)
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

			req, err := http.NewRequest(http.MethodPost, "/transfers", bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(req, server.tokenMaker)
			server.router.ServeHTTP(rec, req)
			tc.checkResp(rec)
		})
	}
}

package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	mockdb "github.com/gaggudeep/bank_go/db/mock"
	db "github.com/gaggudeep/bank_go/db/sqlc"
	"github.com/gaggudeep/bank_go/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func randomAccount() db.Account {
	return db.Account{
		ID:        int64(util.RandomFloat(1, 1000)),
		OwnerName: util.RandomOwnerName(),
		Balance:   util.RandomMoney(),
		Currency:  util.RandomCurrency(),
	}
}

func TestGetAccount(t *testing.T) {
	acc := randomAccount()

	testCases := []struct {
		name       string
		accId      int64
		buildStubs func(store *mockdb.MockStore)
		checkResp  func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			accId: acc.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc.ID)).
					Times(1).
					Return(acc, nil)
			},
			checkResp: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, &acc)
			},
		},
		{
			name:  "NotFound",
			accId: acc.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResp: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:  "InternalError",
			accId: acc.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResp: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "InvalidID",
			accId: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResp: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	config := &util.Config{
		CustomValidators: util.CustomValidators,
	}
	server := NewServer(store, config)

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs(store)

			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accId)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.checkResp(recorder)
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

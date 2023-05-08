package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	mockdb "github.com/gaggudeep/bank_go/db/mock"
	db "github.com/gaggudeep/bank_go/db/sqlc"
	"github.com/gaggudeep/bank_go/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type EqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

func (e EqCreateUserParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	err := util.ValidatePassword(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}

	e.arg.HashedPassword = arg.HashedPassword
	return reflect.DeepEqual(e.arg, arg)
}

func EqCreateUserParams(arg *db.CreateUserParams, password string) gomock.Matcher {
	return EqCreateUserParamsMatcher{*arg, password}
}

func (e EqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func randomUser(t *testing.T) (user db.User, pwd string) {
	pwd = util.RandomString(6)
	hashedPwd, err := util.HashPassword(pwd)
	require.NoError(t, err)

	user = db.User{
		Username:       util.RandomOwnerName(),
		HashedPassword: hashedPwd,
		Name:           util.RandomOwnerName(),
		Email:          util.RandomEmail(),
	}
	return
}

func TestCreateUser(t *testing.T) {
	user, pwd := randomUser(t)

	testCases := []struct {
		name       string
		body       gin.H
		buildStubs func(*mockdb.MockStore)
		checkResp  func(*httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username": user.Username,
				"password": pwd,
				"name":     user.Name,
				"email":    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := &db.CreateUserParams{
					Username: user.Username,
					Name:     user.Name,
					Email:    user.Email,
				}

				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, pwd)).
					Times(1).
					Return(user, nil)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rec.Code)
				requireBodyMatchUser(t, rec.Body, &user)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"username": user.Username,
				"password": pwd,
				"name":     user.Name,
				"email":    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, rec.Code)
			},
		},
		{
			name: "DuplicateUsername",
			body: gin.H{
				"username": user.Username,
				"password": pwd,
				"name":     user.Name,
				"email":    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, &pq.Error{Code: "23505"})
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, rec.Code)
			},
		},
		{
			name: "InvalidUsername",
			body: gin.H{
				"username": "invalid-user#",
				"password": pwd,
				"name":     user.Name,
				"email":    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name: "InvalidUsername",
			body: gin.H{
				"username": user.Username,
				"password": pwd,
				"name":     user.Name,
				"email":    "invalid-email",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResp: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name: "InvalidUsername",
			body: gin.H{
				"username": user.Username,
				"password": "123",
				"name":     user.Name,
				"email":    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
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

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/users", bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(rec, req)
			tc.checkResp(rec)
		})
	}
}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user *db.User) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var actualUser db.User
	err = json.Unmarshal(data, &actualUser)
	require.NoError(t, err)
	require.Equal(t, user.Username, actualUser.Username)
	require.Equal(t, user.Name, actualUser.Name)
	require.Equal(t, user.Email, actualUser.Email)
	require.Empty(t, actualUser.HashedPassword)
}

package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mockdb "github.com/nhoc20170861/simple-bank/db/mock"
	db "github.com/nhoc20170861/simple-bank/db/sqlc"
	"github.com/nhoc20170861/simple-bank/token"
	util "github.com/nhoc20170861/simple-bank/util"
	"github.com/stretchr/testify/require"
)

// TestGetAccountAPI tests the GetAccount API endpoint
func TestGetAccountAPI(t *testing.T) {
	// Create a random account for testing
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	// Define test cases for the GetAccount API
	testCases := []struct {
		// name of the test case
		name string
		// accountID to be used in the test case
		accountID int64
		// add authorization header
		setupAuth func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		// buildStubs sets up mock expectations
		buildStubs func(store *mockdb.MockStore)
		// checkResponse validates the response
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// Expect GetAccount to be called with the correct account ID
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				// Check if the response body matches the expected account
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// Expect GetAccount to be called with the correct account ID
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalServerError",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// Expect GetAccount to be called with the correct account ID
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// Expect GetAccount not to be called
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "UnauthorizedUser",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, request, tokenMaker, authorizationTypeBearer, "unauthorized_user", time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// Expect GetAccount to be called with the correct account ID
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).Return(account, nil).Times(1)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	// Execute each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create a new mock store
			store := mockdb.NewMockStore(ctrl)

			tc.buildStubs(store)

			// Create a new test server
			testServer := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			// Create a new HTTP GET request
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/accounts/%d", tc.accountID), nil)
			require.NoError(t, err)

			tc.setupAuth(t, req, testServer.tokenMaker)

			// Serve the HTTP request
			testServer.router.ServeHTTP(recorder, req)

			// Check the response
			tc.checkResponse(recorder)
		})
	}
}

func TestCreateAccountAPI(t *testing.T) {
	// Create a random account for testing
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	// Define test cases for the CreateAccount API
	testCases := []struct {
		// name of the test case
		name string
		// body of request
		body gin.H
		// auth method
		setupAuth func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		// buildStubs sets up mock expectations
		buildStubs func(store *mockdb.MockStore)
		// checkResponse validates the response
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"owner":    account.Owner,
				"currency": account.Currency,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    account.Owner,
					Currency: account.Currency,
					Balance:  0,
				}

				// Expect CreateAccount to be called with the correct arguments
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				// Check if the response body matches the expected account
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{

			name: "InvalidCurrency",
			body: gin.H{
				"owner":    account.Owner,
				"currency": "invalid",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// Expect CreateAccount not to be called
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalServerError",
			body: gin.H{
				"owner":    account.Owner,
				"currency": account.Currency,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// Expect CreateAccount to be called with the correct arguments
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "StatusForbidden",
			body: gin.H{
				"owner":    account.Owner,
				"currency": account.Currency,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    account.Owner,
					Currency: account.Currency,
					Balance:  0,
				}
				// Expect CreateAccount to be called with the correct arguments
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Eq(arg)).Times(1).Return(db.Account{}, db.ErrForeignKeyViolation)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
	}

	// Execute each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create a new mock store
			store := mockdb.NewMockStore(ctrl)

			tc.buildStubs(store)

			// Create a new test server
			testServer := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			// Create a new HTTP POST request
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, req, testServer.tokenMaker)

			// Serve the HTTP request
			testServer.router.ServeHTTP(recorder, req)

			// Check the response
			tc.checkResponse(recorder)
		})
	}
}

func TestListAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	n := 5
	accounts := make([]db.Account, n)

	for i := 0; i < n; i++ {
		accounts[i] = randomAccount(user.Username)
	}

	type Query struct {
		pageID   int
		pageSize int
	}

	testCases := []struct {
		name          string
		query         Query
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListAccountsParams{
					Owner:  user.Username,
					Limit:  int32(n),
					Offset: 0,
				}
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccounts(t, recorder.Body, accounts)
			},
		},
		{
			name: "InvalidPageID",
			query: Query{
				pageID:   -1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidPageSize",
			query: Query{
				pageID:   1,
				pageSize: 100000,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},

			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalServer",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create a new mock store
			store := mockdb.NewMockStore(ctrl)

			tc.buildStubs(store)

			// Create a new test server
			testServer := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, "/accounts", nil)
			require.NoError(t, err)

			tc.setupAuth(t, req, testServer.tokenMaker)
			// Add query parameters to request URL
			q := req.URL.Query()
			q.Add("page_id", fmt.Sprintf("%d", tc.query.pageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			req.URL.RawQuery = q.Encode()
			// Serve the HTTP request
			testServer.router.ServeHTTP(recorder, req)

			// Check the response
			tc.checkResponse(recorder)
		})
	}
}
func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)

	if err != nil {
		t.Logf("unmarshaling response body to account failed: %v", err)
		t.Logf("response body: %s", data)
	}
	require.NoError(t, err)
	if !reflect.DeepEqual(account, gotAccount) {
		t.Logf("expected account: %v, got: %v", account, gotAccount)
	}
	require.Equal(t, account, gotAccount)
}

func requireBodyMatchAccounts(t *testing.T, body *bytes.Buffer, accounts []db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccounts []db.Account
	err = json.Unmarshal(data, &gotAccounts)
	require.NoError(t, err)
	require.Equal(t, accounts, gotAccounts)
}

package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "tutorial.sqlc.dev/app/db/mock"
	db "tutorial.sqlc.dev/app/db/sqlc"
	"tutorial.sqlc.dev/app/utils"
)

func TestGetAccount(t *testing.T) {
	account := randomAccount()

	testCase := []struct {
		name          string
		accountID     int64
		buildMock     func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{name: "OK",
			accountID: account.ID,
			buildMock: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccounts(gomock.Any(), gomock.Eq(account.ID)).
					Return(account, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{name: "NotFound",
			accountID: account.ID,
			buildMock: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccounts(gomock.Any(), gomock.Eq(account.ID)).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
				require.Contains(t, recorder.Body.String(), "account not found")
			},
		},
		{name: "InternalError",
			accountID: account.ID,
			buildMock: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccounts(gomock.Any(), gomock.Eq(account.ID)).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{name: "InvalidID",
			accountID: 0,
			buildMock: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}
	for i := range testCase {
		tc := testCase[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildMock(store)
			server := NewServer(store)
			recorder := httptest.NewRecorder()
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(recorder)
		})

	}
}

func randomAccount() db.Account {
	return db.Account{
		ID:        utils.RandomInt(1, 1000),
		Owner:     utils.RandomOwner(),
		Currency:  utils.RandomCurrency(),
		Balance:   utils.RandomMoney(),
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)

	require.Equal(t, account.ID, gotAccount.ID)
	require.Equal(t, account.Owner, gotAccount.Owner)
	require.Equal(t, account.Balance, gotAccount.Balance)
	require.Equal(t, account.Currency, gotAccount.Currency)
	require.Equal(t, account.CreatedAt.Valid, gotAccount.CreatedAt.Valid)
	require.WithinDuration(t, account.CreatedAt.Time, gotAccount.CreatedAt.Time, time.Second)
}

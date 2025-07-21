package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"tutorial.sqlc.dev/app/utils"
)

func createRandomTransfer(t *testing.T, from, to Account) Transfer {
	arg := CreateTransferParams{
		FromAccountID: sql.NullInt64{Int64: from.ID, Valid: true},
		ToAccountID:   sql.NullInt64{Int64: to.ID, Valid: true},
		Amount:        utils.RandomMoney(),
	}
	transfer, err := testQueries.CreateTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)
	require.Equal(t, arg.FromAccountID, transfer.FromAccountID)
	require.Equal(t, arg.ToAccountID, transfer.ToAccountID)
	require.Equal(t, arg.Amount, transfer.Amount)
	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)
	return transfer
}

func TestCreateTransfer(t *testing.T) {
	from := createRandomAccount(t)
	to := createRandomAccount(t)
	createRandomTransfer(t, from, to)
}

func TestGetTransfer(t *testing.T) {
	from := createRandomAccount(t)
	to := createRandomAccount(t)
	transfer1 := createRandomTransfer(t, from, to)
	transfer2, err := testQueries.GetTransfer(context.Background(), transfer1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, transfer2)
	require.Equal(t, transfer1.ID, transfer2.ID)
	require.Equal(t, transfer1.FromAccountID, transfer2.FromAccountID)
	require.Equal(t, transfer1.ToAccountID, transfer2.ToAccountID)
	require.Equal(t, transfer1.Amount, transfer2.Amount)
	require.WithinDuration(t, transfer1.CreatedAt.Time, transfer2.CreatedAt.Time, time.Second)
}

func TestListTransfers(t *testing.T) {
	from := createRandomAccount(t)
	to := createRandomAccount(t)
	for i := 0; i < 10; i++ {
		createRandomTransfer(t, from, to)
	}
	arg := ListTransfersParams{
		FromAccountID: sql.NullInt64{Int64: from.ID, Valid: true},
		ToAccountID:   sql.NullInt64{Int64: to.ID, Valid: true},
		Limit:         5,
		Offset:        5,
	}
	transfers, err := testQueries.ListTransfers(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, transfers, 5)
	for _, transfer := range transfers {
		require.NotEmpty(t, transfer)
	}
}

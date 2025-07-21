package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTranferTx(t *testing.T) {
	store := NewStore(testDB)
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Printf(">> before: account1 balance: %d, account2 balance: %d\n", account1.Balance, account2.Balance)
	n := 10
	amount := int64(10)
	errs := make(chan error)
	results := make(chan TransferTxResult)
	for i := 0; i < n; i++ {
		txName := fmt.Sprintf("tx-%d", i+1)
		go func() {
			ctx := context.WithValue(context.Background(), txKey, txName)
			arg := TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			}
			result, err := store.transferTx(ctx, arg)
			errs <- err
			results <- result
		}()
	}

	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
		result := <-results
		require.NotEmpty(t, result)
		require.Equal(t, account1.ID, result.FromAccount.ID)
		require.Equal(t, account2.ID, result.ToAccount.ID)
		require.Equal(t, amount, result.Transfer.Amount)
		require.NotZero(t, result.Transfer.ID)
		require.NotZero(t, result.FromEntry.ID)

		_, err = store.GetTransfer(context.Background(), result.Transfer.ID)
		require.NoError(t, err)

		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, -amount, fromEntry.Amount)
		require.Equal(t, account1.ID, fromEntry.AccountID.Int64)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		ToEntry := result.ToEntry
		require.NotEmpty(t, ToEntry)
		require.Equal(t, amount, ToEntry.Amount)
		require.Equal(t, account2.ID, ToEntry.AccountID.Int64)
		require.NotZero(t, ToEntry.ID)
		require.NotZero(t, ToEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), ToEntry.ID)
		require.NoError(t, err)
		//check account balances
		fmt.Println(">> tx: ", i+1, " | from account balance: ", result.FromAccount.Balance, " | to account balance: ", result.ToAccount.Balance)
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		ToAccount := result.ToAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		diff1 := account1.Balance - fromAccount.Balance
		diff2 := ToAccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1, k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	updatedAccount1, err := store.GetAccounts(context.Background(), account1.ID)
	require.NoError(t, err)
	require.Equal(t, account1.Balance-int64(n)*amount, updatedAccount1.Balance)
	updatedAccount2, err := store.GetAccounts(context.Background(), account2.ID)
	require.NoError(t, err)
	fmt.Printf(">> after: account1 balance: %d, account2 balance: %d\n", updatedAccount1.Balance, updatedAccount2.Balance)
	require.Equal(t, account2.Balance+int64(n)*amount, updatedAccount2.Balance)
	require.NotEqual(t, account1.Balance, updatedAccount1.Balance)
	require.NotEqual(t, account2.Balance, updatedAccount2.Balance)
	require.Equal(t, updatedAccount1.Balance+updatedAccount2.Balance, account1.Balance+account2.Balance)
}

func TestTranferTxDeadLock(t *testing.T) {
	store := NewStore(testDB)
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Printf(">> before: account1 balance: %d, account2 balance: %d\n", account1.Balance, account2.Balance)
	n := 10
	amount := int64(10)
	errs := make(chan error)
	results := make(chan TransferTxResult)
	for i := 0; i < n; i++ {
		FromAccountID := account1.ID
		ToAccountID := account2.ID
		if i%2 == 0 {
			FromAccountID = account2.ID
			ToAccountID = account1.ID
		}
		txName := fmt.Sprintf("tx-%d", i+1)
		go func(fromID, toID int64, name string, idx int) {
			ctx := context.WithValue(context.Background(), txKey, name)
			arg := TransferTxParams{
				FromAccountID: fromID,
				ToAccountID:   toID,
				Amount:        amount,
			}
			result, err := store.transferTx(ctx, arg)
			errs <- err
			results <- result
		}(FromAccountID, ToAccountID, txName, i)
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
		result := <-results
		fmt.Printf(">> tx:  %d  | from account balance:  %d  | to account balance:  %d\n", i+1, result.FromAccount.Balance, result.ToAccount.Balance)
	}

	updatedAccount1, err := store.GetAccounts(context.Background(), account1.ID)
	require.NoError(t, err)
	updatedAccount2, err := store.GetAccounts(context.Background(), account2.ID)
	require.NoError(t, err)
	fmt.Printf(">> after: account1 balance: %d, account2 balance: %d\n", updatedAccount1.Balance, updatedAccount2.Balance)
}

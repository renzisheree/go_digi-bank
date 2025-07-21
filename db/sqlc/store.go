package db

import (
	"context"
	"database/sql"
	"fmt"
)

type store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *store {
	return &store{
		Queries: New(db),
		db:      db,
	}
}
func (store *store) executeTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	if err := fn(q); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

var txKey = struct{}{}

func (store *store) transferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult
	err := store.executeTx(ctx, func(q *Queries) error {
		var err error
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: sql.NullInt64{Int64: arg.FromAccountID, Valid: true},
			ToAccountID:   sql.NullInt64{Int64: arg.ToAccountID, Valid: true},
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: sql.NullInt64{Int64: arg.FromAccountID, Valid: true},
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: sql.NullInt64{Int64: arg.ToAccountID, Valid: true},
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		// Update account balances with proper ordering to avoid deadlocks
		var fromAccount, toAccount Account
		if arg.FromAccountID < arg.ToAccountID {
			fromAccount, toAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
		} else {
			toAccount, fromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
		}
		if err != nil {
			return err
		}
		result.FromAccount = fromAccount
		result.ToAccount = toAccount
		return nil
	})

	return result, err
}

func addMoney(ctx context.Context, q *Queries, account1ID int64, amount1 int64, account2ID int64, amount2 int64) (account1 Account, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:      account1ID,
		Ammount: amount1,
	})
	if err != nil {
		return
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:      account2ID,
		Ammount: amount2,
	})
	return
}

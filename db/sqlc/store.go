package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provides all function to execute db queries and transaction
type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
}

// SQLStore provides all fucntion to excecute SQL queries and transaction
type SQLStore struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

// execTx executes a function within a database transaction.
// It takes a context and a function that takes Queries as input and returns an error.
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	// Begin a new transaction
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	fmt.Println("starting tx")

	// Create a new Queries object for the transaction
	q := New(tx)

	// Execute the function with the transaction's Queries object
	err = fn(q)
	if err != nil {
		fmt.Printf("tx err: %v, attempting rollback...\n", err)
		// Attempt to rollback the transaction if the function returns an error
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}
	fmt.Println("tx successful, committing...")
	// Commit the transaction if no error occurred
	err = tx.Commit()
	if err != nil {
		return err
	}
	fmt.Println("tx committed")
	return nil
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

func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult
	fmt.Println("starting transfer tx")
	if store == nil {
		return result, fmt.Errorf("store is nil")
	}
	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		fmt.Println("starting transfer execTx")
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		fmt.Printf("created transfer: %v\n", result.Transfer)

		fmt.Println("starting from entry")
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		fmt.Printf("created from entry: %v\n", result.FromEntry)

		fmt.Println("starting to entry")
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})

		if err != nil {
			return err
		}

		fmt.Printf("created to entry: %v\n", result.ToEntry)

		// TODO: update account's balance
		result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
			ID:     arg.FromAccountID,
			Amount: -arg.Amount,
		})

		if err != nil {
			return err
		}

		result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
			ID:     arg.ToAccountID,
			Amount: arg.Amount,
		})

		fmt.Printf("updated from account: %v\n", result.FromAccount)
		return nil
	})

	fmt.Println("tx complete")
	return result, err
}

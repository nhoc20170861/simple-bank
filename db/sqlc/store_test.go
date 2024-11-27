package db

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)
	if testDB == nil {
		log.Fatal("cannot connect to db:")
	}
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Printf(">> before: account1.Balance: %d, account2.Balance: %d\n", account1.Balance, account2.Balance)

	n := 1
	amount := int64(10)
	errs := make(chan error, n)
	results := make(chan TransferTxResult, n)
	existed := make(map[int]bool)

	for i := 0; i < n; i++ {
		go func(i int) {
			fmt.Printf(">> tx %d started\n", i)
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})
			fmt.Printf(">> tx %d completed\n", i)
			errs <- err
			results <- result
		}(i)
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err, fmt.Sprintf("tx %d failed: %v", i, err))

		result := <-results
		require.NotEmpty(t, result, fmt.Sprintf("tx %d returned empty result", i))

		// Check the transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer, fmt.Sprintf("tx %d: transfer is empty", i))
		require.Equal(t, account1.ID, transfer.FromAccountID, fmt.Sprintf("tx %d: from_account_id is not equal to account1.ID", i))
		require.Equal(t, account2.ID, transfer.ToAccountID, fmt.Sprintf("tx %d: to_account_id is not equal to account2.ID", i))
		require.Equal(t, amount, transfer.Amount, fmt.Sprintf("tx %d: amount is not equal to %d", i, amount))
		require.NotZero(t, transfer.ID, fmt.Sprintf("tx %d: transfer.ID is zero", i))
		require.NotZero(t, transfer.CreatedAt, fmt.Sprintf("tx %d: transfer.CreatedAt is zero", i))

		// Check the transfer in the database
		gotTransfer, err := store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err, fmt.Sprintf("tx %d: get transfer failed: %v", i, err))
		require.Equal(t, transfer, gotTransfer, fmt.Sprintf("tx %d: transfer is not equal to gotTransfer", i))

		// Check the from entry
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry, fmt.Sprintf("tx %d: fromEntry is empty", i))
		require.Equal(t, account1.ID, fromEntry.AccountID, fmt.Sprintf("tx %d: fromEntry.AccountID is not equal to account1.ID", i))
		require.Equal(t, -amount, fromEntry.Amount, fmt.Sprintf("tx %d: fromEntry.Amount is not equal to -%d", i, amount))
		require.NotZero(t, fromEntry.ID, fmt.Sprintf("tx %d: fromEntry.ID is zero", i))
		require.NotZero(t, fromEntry.CreatedAt, fmt.Sprintf("tx %d: fromEntry.CreatedAt is zero", i))

		// Check the from entry in the database
		gotFromEntry, err := store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err, fmt.Sprintf("tx %d: get fromEntry failed: %v", i, err))
		require.Equal(t, fromEntry, gotFromEntry, fmt.Sprintf("tx %d: fromEntry is not equal to gotFromEntry", i))

		// Check the to entry
		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry, fmt.Sprintf("tx %d: toEntry is empty", i))
		require.Equal(t, account2.ID, toEntry.AccountID, fmt.Sprintf("tx %d: toEntry.AccountID is not equal to account2.ID", i))
		require.Equal(t, amount, toEntry.Amount, fmt.Sprintf("tx %d: toEntry.Amount is not equal to %d", i, amount))
		require.NotZero(t, toEntry.ID, fmt.Sprintf("tx %d: toEntry.ID is zero", i))
		require.NotZero(t, toEntry.CreatedAt, fmt.Sprintf("tx %d: toEntry.CreatedAt is zero", i))

		// Check the to entry in the database
		gotToEntry, err := store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err, fmt.Sprintf("tx %d: get toEntry failed: %v", i, err))
		require.Equal(t, toEntry, gotToEntry, fmt.Sprintf("tx %d: toEntry is not equal to gotToEntry", i))

		// Check account
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount, fmt.Sprintf("tx %d: fromAccount is empty", i))
		require.Equal(t, account1.ID, fromAccount.ID, fmt.Sprintf("tx %d: fromAccount.ID is not equal to account1.ID", i))

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount, fmt.Sprintf("tx %d: toAccount is empty", i))
		require.Equal(t, account2.ID, toAccount.ID, fmt.Sprintf("tx %d: toAccount.ID is not equal to account2.ID", i))

		// check account's balance
		fmt.Println(">> tx", i, "before:", fromAccount.Balance, toAccount.Balance)
		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2, fmt.Sprintf("tx %d: diff1 is not equal to diff2", i))
		require.True(t, diff1 > 0, fmt.Sprintf("tx %d: diff1 is not greater than 0", i))
		require.True(t, diff1%amount == 0, fmt.Sprintf("tx %d: diff1 is not divisible by amount", i))

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n, fmt.Sprintf("tx %d: k is not between 1 and n", i))
		require.NotEmpty(t, result.Transfer, fmt.Sprintf("tx %d: transfer is empty", i))

		require.NotContains(t, existed, k)
		existed[k] = true
	}

	// Check the final updated balances
	updateAccount1, err1 := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err1)
	updateAccount2, err2 := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err2)

	fmt.Printf(">> after: %v, %v\n", updateAccount1.Balance, updateAccount2.Balance)
	require.Equal(t, account1.Balance-int64(n)*amount, updateAccount1.Balance)
	require.Equal(t, account2.Balance+int64(n)*amount, updateAccount2.Balance)
}

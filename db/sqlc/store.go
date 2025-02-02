package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provides all functions to execute db queries and transactions
type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

func (s *Store) execTx(ctx context.Context, fn func(*Queries error) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	
	q:= New(tx)
	err = fn(q)
	
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to rollback transaction: %w", rbErr)
		}
		return err
	}
	
	return tx.Commit()
}

// TransferTxParams represents the input parameters required for transferring money between two accounts.
type TransferTxParams struct {
	FromAccountID int64
	ToAccountID int64
	Amount int64
}

type TransferTxResult struct {
	Transfer Transfer `json:"transfer"`
	FromAccount Account `json:"from_account"`
	ToAccount Account `json:"to_account"`
	FromEntry Entry `json:"from_entry"`
	ToEntry Entry `json:"to_entry"`
}

// TransferTx performs a money transfer from one account to another.
// It creates a database transaction to ensure the consistency of the transfer operation.
// The function executes multiple queries in a single transaction, including creating transfer records
// and updating account balances, ensuring atomicity for the entire operation.
func (s *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult
	
	err:= s.execTx(ctx, func(q *Queries) error {
		result.Transfer, err := q.CreateTransfer(ctx, Create)
		
		return nil
	})
	
	return result, err
}

package txManager

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ctxTxKey struct{}

type Querier interface{
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults 
}

func GetQuerier(ctx context.Context, pool *pgxpool.Pool) Querier {
	if tx, ok := GetTx(ctx); ok {
		return tx
	}
	return pool
}

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

func (tm *TxManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	
	ctx = context.WithValue(ctx, ctxTxKey{}, tx)

	if err := fn(ctx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func GetTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(ctxTxKey{}).(pgx.Tx)
	return tx, ok
}
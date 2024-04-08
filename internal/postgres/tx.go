package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type TxBeginner interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (Tx, error)
}

type Tx interface {
	QueryExecer
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

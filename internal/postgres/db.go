package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/monobearotaku/online-chat-api/internal/config"
)

type Db interface {
	QueryExecer
	TxBeginner
}

type db struct {
	*pgx.Conn
}

func (db *db) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (Tx, error) {
	return db.Conn.BeginTx(ctx, txOptions)
}

func NewDbConnection(ctx context.Context, config config.Config) (Db, func()) {
	conn, err := pgx.Connect(context.Background(),
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
			config.Db.User,
			config.Db.Password,
			config.Db.Host,
			config.Db.Port,
			config.Db.Name,
		))

	if err != nil {
		panic(fmt.Errorf("unable to connect to Database: %w", err))
	}

	return &db{Conn: conn}, func() {
		conn.Close(ctx)
	}
}

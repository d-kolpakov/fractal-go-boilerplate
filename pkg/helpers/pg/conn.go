package pg

import (
	"context"
	"fmt"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/stats"
	"github.com/d-kolpakov/logger"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"math/rand"
	"time"
)

type ConnWrapper struct {
	pool    *pgxpool.Pool
	ls      *LSHelper
	connStr string
}

func NewConn(pool *pgxpool.Pool, stats *stats.Stats, l *logger.Logger, connStr string) *ConnWrapper {
	ls := &LSHelper{
		stats: stats,
		l:     l,
	}
	return &ConnWrapper{
		pool:    pool,
		ls:      ls,
		connStr: connStr,
	}
}

func (w *ConnWrapper) ConnString() string {
	return w.connStr
}

func (w *ConnWrapper) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	t := time.Now()
	rows, err := w.pool.Query(ctx, sql, args...)
	w.ls.logQuery(ctx, sql, args)
	w.ls.statDuration(ctx, sql, time.Since(t))

	if err != nil {
		w.ls.statError(ctx, sql)
	}

	return rows, err
}

func (w *ConnWrapper) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	t := time.Now()
	row := w.pool.QueryRow(ctx, sql, args...)
	w.ls.logQuery(ctx, sql, args)
	w.ls.statDuration(ctx, sql, time.Since(t))

	return row
}

func (w *ConnWrapper) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	t := time.Now()
	tag, err := w.pool.Exec(ctx, sql, args...)
	w.ls.logQuery(ctx, sql, args)
	w.ls.statDuration(ctx, sql, time.Since(t))

	if err != nil {
		w.ls.statError(ctx, sql)
	}

	return tag, err
}

func (w *ConnWrapper) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (TxWrapper, error) {
	rand.Seed(time.Now().UnixNano())
	hash := fmt.Sprintf("tx:%d", rand.Uint64())

	txw := TxWrapper{
		ls:   w.ls,
		hash: hash,
		t:    time.Now(),
	}
	tx, err := w.pool.BeginTx(ctx, txOptions)
	txw.tx = tx

	w.ls.l.NewLogEvent().
		WithTag("kind", "tx_begin").
		Log(w.ls.logHash(ctx, hash), txOptions)

	return txw, err
}

package pg

import (
	"context"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"time"
)

type TxWrapper struct {
	ls   *LSHelper
	hash string
	t    time.Time
	tx   pgx.Tx
}

func (tx *TxWrapper) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	t := time.Now()
	rows, err := tx.tx.Query(ctx, sql, args...)
	tx.ls.logQuery(tx.ls.logHash(ctx, tx.hash), sql, args)
	tx.ls.statDurationTx(ctx, sql, time.Since(t), tx.hash)

	if err != nil {
		tx.ls.statErrorTx(ctx, sql, tx.hash)
	}

	return rows, err
}

func (tx *TxWrapper) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	t := time.Now()
	row := tx.tx.QueryRow(ctx, sql, args...)
	tx.ls.logQuery(tx.ls.logHash(ctx, tx.hash), sql, args)
	tx.ls.statDurationTx(ctx, sql, time.Since(t), tx.hash)

	return row
}

func (tx *TxWrapper) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	t := time.Now()
	tag, err := tx.tx.Exec(ctx, sql, args...)
	tx.ls.logQuery(tx.ls.logHash(ctx, tx.hash), sql, args)
	tx.ls.statDurationTx(ctx, sql, time.Since(t), tx.hash)

	if err != nil {
		tx.ls.statErrorTx(ctx, sql, tx.hash)
	}

	return tag, err
}

func (tx *TxWrapper) Commit(ctx context.Context) error {
	err := tx.tx.Commit(ctx)

	if err != nil {
		tx.ls.statErrorTx(ctx, "tx.error.commit", tx.hash)
		tx.ls.l.NewLogEvent().
			WithTag("kind", "tx_commited").
			WithTag("status", "fail").
			Alert(tx.ls.logHash(ctx, tx.hash), err)
	} else {
		tx.ls.l.NewLogEvent().
			WithTag("kind", "tx_commited").
			WithTag("status", "success").
			Log(tx.ls.logHash(ctx, tx.hash), "")
	}

	tx.ls.statDurationTx(ctx, "tx.commited", time.Since(tx.t), tx.hash)

	return err
}

func (tx *TxWrapper) Rollback(ctx context.Context) error {
	err := tx.tx.Rollback(ctx)

	if err != nil {
		tx.ls.statErrorTx(ctx, "tx.error.rollback", tx.hash)
		tx.ls.l.NewLogEvent().
			WithTag("kind", "tx_rollbacked").
			WithTag("status", "fail").
			Alert(tx.ls.logHash(ctx, tx.hash), err)
	} else {
		tx.ls.l.NewLogEvent().
			WithTag("kind", "tx_rollbacked").
			WithTag("status", "success").
			Log(tx.ls.logHash(ctx, tx.hash), "")
	}

	tx.ls.statDurationTx(ctx, "tx.rollbacked", time.Since(tx.t), tx.hash)

	return err
}

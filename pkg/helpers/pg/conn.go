package pg

import (
	"context"
	"fmt"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/stats"
	"github.com/d-kolpakov/logger"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"strings"
	"time"
)

type ConnWrapper struct {
	pool  *pgxpool.Pool
	stats *stats.Stats
	l     *logger.Logger
}

func NewConn(pool *pgxpool.Pool, stats *stats.Stats, l *logger.Logger) *ConnWrapper {
	return &ConnWrapper{
		pool:  pool,
		stats: stats,
		l:     l,
	}
}

func (w *ConnWrapper) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	t := time.Now()
	rows, err := w.pool.Query(ctx, sql, args...)
	w.logQuery(ctx, sql, args)
	w.statDuration(ctx, sql, time.Since(t))

	if err != nil {
		w.statError(ctx, sql)
	}

	return rows, err
}

func (w *ConnWrapper) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	t := time.Now()
	row := w.pool.QueryRow(ctx, sql, args...)
	w.logQuery(ctx, sql, args)
	w.statDuration(ctx, sql, time.Since(t))

	return row
}

func (w *ConnWrapper) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	t := time.Now()
	tag, err := w.pool.Exec(ctx, sql, args...)
	w.logQuery(ctx, sql, args)
	w.statDuration(ctx, sql, time.Since(t))

	if err != nil {
		w.statError(ctx, sql)
	}

	return tag, err
}

func (w *ConnWrapper) convertQueryString(q string) string {
	q = strings.ReplaceAll(q, ".", "_")
	return strings.ReplaceAll(q, " ", "_")
}

type QueryLog struct {
	Query string
	Args  map[string]interface{}
}

func (w *ConnWrapper) logQuery(ctx context.Context, q string, args []interface{}) {
	ql := QueryLog{
		Query: q,
	}

	argsConverted := make(map[string]interface{}, len(args))

	for i, v := range args {
		argsConverted[fmt.Sprintf("$%d", i+1)] = v
	}
	ql.Args = argsConverted
	w.l.NewLogEvent().
		WithTag("kind", "pgxpool_query").
		Log(ctx, ql)
}

func (w *ConnWrapper) statDuration(ctx context.Context, q string, t time.Duration) {
	if w.stats == nil {
		return
	}
	intVal := int64(t)

	eType := fmt.Sprintf("server.query.duration.%s", w.convertQueryString(q))

	w.stats.InsertStat(eType, nil, nil, &intVal, nil, w.getXFrSourceFromCtx(ctx), w.getRequestIDFromCtx(ctx))
}

func (w *ConnWrapper) statError(ctx context.Context, q string) {
	if w.stats == nil {
		return
	}

	eType := fmt.Sprintf("server.query.error.%s", w.convertQueryString(q))

	w.stats.InsertStat(eType, nil, nil, nil, nil, w.getXFrSourceFromCtx(ctx), w.getRequestIDFromCtx(ctx))
}

func (w *ConnWrapper) getXFrSourceFromCtx(ctx context.Context) *string {
	var key logger.ContextUIDKey = "source"
	var res *string
	source := ctx.Value(key)
	if source != nil {
		sourceString, ok := source.(string)
		if ok {
			res = &sourceString
		}
	}

	return res
}

func (w *ConnWrapper) getRequestIDFromCtx(ctx context.Context) *string {
	var key logger.ContextUIDKey = "requestID"
	var res *string
	id := ctx.Value(key)
	if id != nil {
		idString, ok := id.(string)
		if ok {
			res = &idString
		}
	}

	return res
}

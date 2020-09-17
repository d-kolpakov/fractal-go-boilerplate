package pg

import (
	"context"
	"fmt"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/stats"
	"github.com/d-kolpakov/logger"
	"strings"
	"time"
)

const hKey logger.ContextUIDKey = "tx_hash"

type LSHelper struct {
	stats *stats.Stats
	l     *logger.Logger
}

func getRequestIDFromCtx(ctx context.Context) *string {
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

func convertQueryString(q string) string {
	q = strings.ReplaceAll(q, ".", "_")
	return strings.ReplaceAll(q, " ", "_")
}

func (ls *LSHelper) logQuery(ctx context.Context, q string, args []interface{}) {
	ql := QueryLog{
		Query: q,
	}

	argsConverted := make(map[string]interface{}, len(args))

	for i, v := range args {
		argsConverted[fmt.Sprintf("$%d", i+1)] = v
	}
	ql.Args = argsConverted
	ls.l.NewLogEvent().
		WithTag("kind", "pgxpool_query").
		Log(ctx, ql)
}

type QueryLog struct {
	Query string
	Args  map[string]interface{}
}

func (ls *LSHelper) statDuration(ctx context.Context, q string, t time.Duration) {
	if ls.stats == nil {
		return
	}
	intVal := int64(t)

	eType := fmt.Sprintf("server.query.duration.%s", convertQueryString(q))

	ls.stats.InsertStat(eType, nil, nil, &intVal, nil, nil, getRequestIDFromCtx(ctx))
}

func (ls *LSHelper) statError(ctx context.Context, q string) {
	if ls.stats == nil {
		return
	}

	eType := fmt.Sprintf("server.query.error.%s", convertQueryString(q))

	ls.stats.InsertStat(eType, nil, nil, nil, nil, nil, getRequestIDFromCtx(ctx))
}

func (ls *LSHelper) statDurationTx(ctx context.Context, q string, t time.Duration, hash string) {
	if ls.stats == nil {
		return
	}
	intVal := int64(t)

	eType := fmt.Sprintf("server.query.duration.%s", convertQueryString(q))

	ls.stats.InsertStat(eType, nil, nil, &intVal, nil, &hash, getRequestIDFromCtx(ctx))
}

func (ls *LSHelper) statErrorTx(ctx context.Context, q string, hash string) {
	if ls.stats == nil {
		return
	}

	eType := fmt.Sprintf("server.query.error.%s", convertQueryString(q))

	ls.stats.InsertStat(eType, nil, nil, nil, nil, &hash, getRequestIDFromCtx(ctx))
}

func (ls *LSHelper) logHash(ctx context.Context, hash string) context.Context {
	return context.WithValue(ctx, hKey, hash)
}

package stats

import (
	"context"
	"fmt"
	"github.com/d-kolpakov/logger"
	"github.com/jackc/pgx/v4/pgxpool"
	"regexp"
	"strings"
	"time"
)

const DefaultStatsTable = "default_stats_table"

type Stats struct {
	o  *Options
	db *pgxpool.Pool
	l  *logger.Logger
}

type Options struct {
	Sn         string
	table      string
	Expiration time.Duration
}

func GetStatsHelper(o *Options, db *pgxpool.Pool, l *logger.Logger) *Stats {
	o.table = getStatsTable(o.Sn)
	stats := &Stats{
		o:  o,
		db: db,
		l:  l,
	}

	stats.migrate()
	stats.collectGarbage()

	return stats
}

func (s *Stats) InsertStat(eventType string, url, stringVal *string, intVal *int64, data, xFrSource, requestID *string) {
	t := time.Now().In(time.UTC)
	go s.insertStat(eventType, url, stringVal, intVal, data, xFrSource, requestID, t)
}

func (s *Stats) insertStat(eventType string, url, stringVal *string, intVal *int64, data, xFrSource, requestID *string, t time.Time) {
	q := fmt.Sprintf(`INSERT INTO %s (event_type, created_at, url, string_val, int_val, data, x_fr_source, request_id)
    VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, s.o.table)

	args := []interface{}{
		eventType,
		t,
		url,
		stringVal,
		intVal,
		data,
		xFrSource,
		requestID,
	}

	s.l.NewLogEvent().
		WithTag("kind", "sql_query").
		WithTag("process", "stats_insert").
		Debug(context.Background(), fmt.Sprintf("q: %s, args: %v", q, args))

	_, err := s.db.Exec(context.Background(), q, args...)
	if err != nil {
		s.l.NewLogEvent().
			WithTag("kind", "sql_error").
			WithTag("process", "stats_insert").
			Error(context.Background(), err)
	}
}

func getStatsTable(sn string) string {
	res := DefaultStatsTable
	if len(sn) == 0 {
		return res
	}
	reg, err := regexp.Compile(`[^a-zA-Z]+`)

	if err != nil {
		return res
	}
	sn = strings.ToLower(sn)

	res = "stats_"

	res += reg.ReplaceAllString(sn, "_")

	return res
}

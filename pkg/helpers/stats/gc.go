package stats

import (
	"context"
	"fmt"
	"time"
)

func (s *Stats) collectGarbage() {
	s.l.NewLogEvent().
		WithTag("process", "stats_gc").
		Debug(context.Background(), "init stats gc")

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		s.removeOldRows()
		for t := range ticker.C {
			s.l.NewLogEvent().
				WithTag("kind", "tick_event").
				WithTag("process", "stats_gc").
				Debug(context.Background(), fmt.Sprintf("gc run %v", t))

			s.removeOldRows()
		}
	}()
}

func (s *Stats) removeOldRows() {
	q := fmt.Sprintf(`DELETE FROM %s WHERE created_at < $1;`, s.o.table)
	args := []interface{}{
		time.Now().In(time.UTC).Add(-1 * s.o.Expiration),
	}

	s.l.NewLogEvent().
		WithTag("kind", "sql_query").
		WithTag("process", "stats_gc").
		Debug(context.Background(), fmt.Sprintf("q: %s, args: %v", q, args))

	_, err := s.db.Exec(q, args...)
	if err != nil {
		s.l.NewLogEvent().
			WithTag("kind", "sql_error").
			WithTag("process", "stats_gc").
			Error(context.Background(), err)
	}
}

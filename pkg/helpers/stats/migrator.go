package stats

import (
	"context"
	"fmt"
)

func (s *Stats) migrate() {
	queryPattern := `create table IF NOT EXISTS %s
(
	id serial not null
		constraint %s_pk
			primary key,
	event_type varchar not null,
	created_at timestamp not null,
	x_fr_source varchar,
	request_id varchar,
	url varchar,
	string_val text,
	int_val NUMERIC,
	data jsonb
);`
	q := fmt.Sprintf(queryPattern, s.o.table, s.o.table)

	s.l.NewLogEvent().
		WithTag("kind", "sql_query").
		WithTag("process", "stats_migration").
		Debug(context.Background(), fmt.Sprintf("q: %s, args: %v", q, nil))

	_, err := s.db.Exec(q)

	if err != nil {
		s.l.NewLogEvent().
			WithTag("kind", "sql_error").
			WithTag("process", "stats_migration").
			Error(context.Background(), err)
	}
}

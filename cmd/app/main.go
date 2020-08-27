package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/d-kolpakov/fractal-go-boilerplate/databases"
	"github.com/d-kolpakov/fractal-go-boilerplate/internal/routes"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/logger/drivers"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/stats"
	"github.com/d-kolpakov/logger"
	"github.com/d-kolpakov/logger/drivers/stdout"
	"github.com/dhnikolas/configo"
	"log"
	"net/http"
	"time"
)

const ServiceName = "fractal-go-boilerplate"

func main() {
	lDrivers := make([]logger.LogDriver, 0, 5)

	stdoutLD := &stdout.STDOUTDriver{}
	stdoutLDWrapped := &drivers.STDOUTDriver{Base: stdoutLD}

	lDrivers = append(lDrivers, stdoutLDWrapped)

	lc := logger.LoggerConfig{
		ServiceName: ServiceName,
		Level:       configo.EnvInt("logging-level", logger.TRACE),
		Buffer:      configo.EnvInt("logging-buffer-size", 1000),
		Output:      lDrivers,
		TagsFromCtx: map[logger.ContextUIDKey]string{
			logger.ContextUIDKey("requestID"): "n",
			logger.ContextUIDKey("token"):     "n",
			logger.ContextUIDKey("source"):    "n",
			logger.ContextUIDKey("from"):      "n",
		},
		NeedToLog: func(ctx context.Context, configuredLevel, level int) bool {
			//todo log function
			return true
		},
	}
	l, err := logger.GetLogger(lc)

	if err != nil {
		panic(err)
	}

	l.NewLogEvent().Debug(context.Background(), fmt.Sprintf(`start %s service`, ServiceName))

	statsOption := &stats.Options{
		Sn:         ServiceName,
		Expiration: time.Duration(configo.EnvInt("stats_expiration", 240)) * time.Hour,
	}
	statsClient := stats.GetStatsHelper(statsOption, getStatsDb(), l)
	stdoutLDWrapped.SetStats(statsClient)

	route := routes.Routing{
		ServiceName: ServiceName,
		L:           l,
		Db:          nil,
		AppVersion:  configo.EnvString("app-version", "1.0.0"),
		Stats:       statsClient,
	}

	err = route.InitRouter()

	if err != nil {
		panic(err)
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", configo.EnvInt("app-server-port", 8081)), route.R))
}

func getStatsDb() *sql.DB {
	o := &databases.Options{
		Username:    configo.EnvString("db-stats-username", "db_user"),
		Password:    configo.EnvString("db-stats-password", "pwd0123456789"),
		Host:        configo.EnvString("db-stats-host", "localhost:5432"),
		DB:          configo.EnvString("db-stats-name", "stats"),
		Timeout:     configo.EnvInt("db-stats-timeout", 20),
		MaxOpenConn: configo.EnvInt("db-stats-max-conns", 5),
		MaxIdleConn: configo.EnvInt("db-stats-max-iddle-conns", 3),
	}

	return databases.GetConnection(o)
}

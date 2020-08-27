package main

import (
	"context"
	"fmt"
	"github.com/d-kolpakov/fractal-go-boilerplate/internal/routes"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/logger/drivers"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/stats"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/pg"
	"github.com/d-kolpakov/logger"
	"github.com/d-kolpakov/logger/drivers/stdout"
	"github.com/dhnikolas/configo"
	"github.com/jackc/pgx/v4/pgxpool"
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

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", configo.EnvInt("app-server-port", 8080)), route.R))
}

func getStatsDb() (*pgxpool.Pool, error) {

	//Задаем параметры для подключения к БД
	cfg := &pg.Config{}
	cfg.Host = configo.EnvString("db-stats-host", "localhost")
	cfg.Username = configo.EnvString("db-stats-username", "db_user")
	cfg.Password = configo.EnvString("db-stats-password", "pwd0123456789")
	cfg.Port = configo.EnvString("db-stats-port", "5432")
	cfg.DbName = configo.EnvString("db-stats-name", "stats")
	cfg.Timeout = configo.EnvInt("db-stats-timeout", 20)

	//Создаем конфиг для пула
	poolConfig, err := pg.NewPoolConfig(cfg)
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = int32(configo.EnvInt("db-stats-max-conns", 5))

	c, err := pg.NewConnection(poolConfig)
	if err != nil {
		return nil, err
	}

	return c, nil
}

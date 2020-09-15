package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/d-kolpakov/fractal-go-boilerplate/internal/migrations/common"
	"github.com/d-kolpakov/fractal-go-boilerplate/internal/routes"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/logger/drivers"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/natsclient"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/pg"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/stats"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/probs"
	"github.com/d-kolpakov/logger"
	"github.com/d-kolpakov/logger/drivers/stdout"
	"github.com/dhnikolas/configo"
	"github.com/go-chi/chi"
	"github.com/golang-migrate/migrate"
	bindata "github.com/golang-migrate/migrate/source/go_bindata"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	stan "github.com/nats-io/stan.go"
)

const ServiceName = "fractal-go-boilerplate"

func main() {
	sig := make(chan os.Signal, 1)
	notify := []os.Signal{
		syscall.SIGABRT,
		syscall.SIGALRM,
		syscall.SIGBUS,
		syscall.SIGFPE,
		syscall.SIGHUP,
		syscall.SIGILL,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGPIPE,
		syscall.SIGQUIT,
		syscall.SIGSEGV,
		syscall.SIGTERM,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
		syscall.SIGPROF,
		syscall.SIGSYS,
		syscall.SIGTRAP,
		syscall.SIGVTALRM,
		syscall.SIGXCPU,
		syscall.SIGXFSZ,
	}

	signal.Notify(sig, notify...)

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
			logger.ContextUIDKey("shard"):     "-1",
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
	statsPgPool, err := getStatsDb()

	if err != nil {
		panic(err)
	}
	statsClient := stats.GetStatsHelper(statsOption, statsPgPool, l)
	stdoutLDWrapped.SetStats(statsClient)

	conn, err := getDb(statsClient, l)

	if err != nil {
		panic(err)
	}

	resource := bindata.Resource(common.AssetNames(), common.Asset)
	source, err := bindata.WithInstance(resource)

	if err != nil {
		panic(err)
	}
	m, err := migrate.NewWithSourceInstance("common", source, conn.Common().ConnString())

	if err != nil {
		panic(err)
	}

	err = m.Up()

	if err != nil && err.Error() != "no change" {
		panic(err)
	}

	shardResource := bindata.Resource(common.AssetNames(), common.Asset)
	shardSource, err := bindata.WithInstance(shardResource)

	if err != nil {
		panic(err)
	}
	for i, sConn := range conn.AllShards() {
		m, err := migrate.NewWithSourceInstance(fmt.Sprintf("shard_%d", i), shardSource, sConn.ConnString())

		if err != nil {
			panic(err)
		}

		err = m.Up()

		if err != nil && err.Error() != "no change" {
			panic(err)
		}
	}

	route := routes.Routing{
		ServiceName: ServiceName,
		Stan:        getNatsConsumer(),
		L:           l,
		Db:          nil,
		AppVersion:  configo.EnvString("app-version", "1.0.0"),
		Stats:       statsClient,
		Port:        configo.EnvInt("app-server-port", 8080),
	}

	err = route.InitRouter()

	if err != nil {
		panic(err)
	}

	go func(r *chi.Mux) {
		probs.Ready()
		log.Println(http.ListenAndServe(fmt.Sprintf(":%d", route.Port), route.R))
	}(route.R)

	select {
	case <-sig:
		route.Deregister()
	}
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

func getNatsConsumer() *natsclient.NatsConnection {
	rand.Seed(time.Now().UnixNano())

	no := &natsclient.NatsOptions{
		Cluster:  configo.EnvString("app-nats-cluster", "test-cluster"),
		ClientId: configo.EnvString("app-nats-client-id", fmt.Sprintf("%s-%s", ServiceName, strconv.Itoa(rand.Intn(10000)))),
		Host:     configo.EnvString("app-nats-host", "localhost:4222"),
		ConnectionLostCallback: func(c stan.Conn, e error) {
			probs.SetLivenessError(e)
			panic(e)
		},
		QueueGroup: configo.EnvString("app-queue-group", ServiceName),
	}

	nc, err := natsclient.New(no)

	if err != nil {
		probs.SetReadinessError(err)
		panic(err)
	}

	return nc
}

func getDb(stats *stats.Stats, l *logger.Logger) (*pg.Wrapper, error) {
	return nil, errors.New("not implemented")
	//Задаем параметры для подключения к БД
	cfg := &pg.Config{}
	cfg.Host = configo.EnvString("db-host", "localhost")
	cfg.Username = configo.EnvString("db-username", "db_user")
	cfg.Password = configo.EnvString("db-password", "pwd0123456789")
	cfg.Port = configo.EnvString("db-port", "5432")
	cfg.DbName = configo.EnvString("db-name", "service_registry")
	cfg.Timeout = configo.EnvInt("db-timeout", 20)

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

	wClient := pg.NewConn(c, stats, l, poolConfig.ConnString())
	wc := pg.NewWrapper(wClient, nil)

	return wc, nil
}

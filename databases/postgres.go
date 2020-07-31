package databases

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"net/url"
)

type Options struct {
	Username    string
	Password    string
	Host        string
	DB          string
	Timeout     int
	MaxOpenConn int
	MaxIdleConn int
}

func GetConnection(o *Options) *sql.DB {

	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable&connect_timeout=%d",
		url.QueryEscape(o.Username),
		url.QueryEscape(o.Password),
		o.Host,
		o.DB,
		o.Timeout)

	db, _ := sql.Open("postgres", connStr)
	db.SetMaxOpenConns(o.MaxOpenConn)
	db.SetMaxIdleConns(o.MaxIdleConn)

	// Проверять подключение к базе данных нужно через Ping.
	// При вызове sql.Open не происходит реальной инициализации подключения
	err := db.Ping()
	if err != nil {
		panic(fmt.Sprintf("Error %s database connection, %v", o.DB, err))
	}

	return db
}

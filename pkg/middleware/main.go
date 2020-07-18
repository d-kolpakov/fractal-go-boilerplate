package middleware

import (
	"database/sql"
	"github.com/d-kolpakov/logger"
)

type Controller struct {
	Db    *sql.DB
	L     *logger.Logger
	SName string
}

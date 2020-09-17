package middleware

import (
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/pg"
	"github.com/d-kolpakov/logger"
)

type Controller struct {
	Db    *pg.Wrapper
	L     *logger.Logger
	SName string
}

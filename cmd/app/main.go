package main

import (
	"fmt"
	"github.com/d-kolpakov/fractal-go-boilerplate/internal/routes"
	"github.com/d-kolpakov/logger"
	"github.com/dhnikolas/configo"
	"log"
	"net/http"
)

const ServiceName = "fractal-go-boilerplate"

func main() {
	lc := logger.LoggerConfig{
		ServiceName: ServiceName,
		Level:       configo.EnvInt("logging-level", logger.LOG),
		Buffer:      configo.EnvInt("logging-buffer-size", 1000),
		Output:      nil,
	}
	l, err := logger.GetLogger(lc)

	if err != nil {
		panic(err)
	}

	l.Debug(fmt.Sprintf(`start %s service`, ServiceName))

	route := routes.Routing{
		ServiceName: ServiceName,
		L:           l,
		Db:          nil,
		AppVersion:  configo.EnvString("app-version", "1.0.0"),
	}

	err = route.InitRouter()

	if err != nil {
		panic(err)
	}

	log.Fatal(http.ListenAndServe(":8080", route.R))
}

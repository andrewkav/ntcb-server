//go:build wireinject
// +build wireinject

package server

import (
	"ntcb-server/service"
	"os"
	"time"

	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/google/wire"
	"github.com/jinzhu/gorm"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func NewLogger() zerolog.Logger {
	level, _ := zerolog.ParseLevel(viper.GetString("log-level"))
	return zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
		Level(level).
		With().
		Timestamp().
		Logger()
}

func GetClickhouseGormDB() (*gorm.DB, error) {
	return gorm.Open("clickhouse", viper.GetString("dsn"))

}

func GetTelemetryService() (*service.TelemetryService, error) {
	wire.Build(
		NewLogger,
		GetClickhouseGormDB,
		service.NewTelemetryService,
	)

	return &service.TelemetryService{}, nil
}

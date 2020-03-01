// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package server

import (
	"github.com/jinzhu/gorm"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"ntcb-server/service"
	"os"
	"time"
)

import (
	_ "github.com/ClickHouse/clickhouse-go"
)

// Injectors from wire.go:

func GetTelemetryService() (*service.TelemetryService, error) {
	db, err := GetClickhouseGormDB()
	if err != nil {
		return nil, err
	}
	logger := NewLogger()
	telemetryService := service.NewTelemetryService(db, logger)
	return telemetryService, nil
}

// wire.go:

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

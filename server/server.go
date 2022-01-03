package server

import (
	"ntcb-server/migration"
	"ntcb-server/ntcb"

	"github.com/spf13/viper"
)

func ListenAndServe() {
	logger := NewLogger()

	dsn := viper.GetString("dsn")
	addr := viper.GetString("host") + ":" + viper.GetString("port")

	if err := migration.Migrate(dsn); err != nil {
		logger.Fatal().Caller().Err(err).Msgf("unable to perform migration")
	}

	ts, err := GetTelemetryService()
	if err != nil {
		logger.Fatal().Caller().Err(err).Msg("unable to create telemetry service")
	}

	srvOptions := ntcb.ServerOptions{
		Address: addr,
		Debug:   viper.GetBool("debug"),
		OnConnectionClosed: func(c *ntcb.Conn, err error) {
			logger.Error().
				Caller().
				Err(err).
				Str("deviceID", c.DeviceID()).
				Str("IP", c.RemoteAddr()).
				Msg("connection error has occurred")
		},
		OnTelemetryMessage: func(c *ntcb.Conn, tm ntcb.TelemetryMessage) {
			logger.Debug().
				Str("deviceID", c.DeviceID()).
				Str("IP", c.RemoteAddr()).
				Msgf("telemetry data received, data=%v", tm)

			if err := ts.Save(c.DeviceID(), &tm); err != nil {
				logger.Error().
					Caller().
					Err(err).
					Str("deviceID", c.DeviceID()).
					Str("IP", c.RemoteAddr()).
					Msg("unable to save telemetry message")
			}
		},
		OnNewConnection: func(c *ntcb.Conn) {
			logger.Info().
				Str("deviceID", c.DeviceID()).
				Str("IP", c.RemoteAddr()).
				Msg("new connection established")
		},
	}
	srv := ntcb.NewServer(srvOptions)

	logger.Info().Msgf("starting NTCB server at %s ", addr)

	if err := srv.ListenAndServe(); err != nil {
		logger.Fatal().Err(err).Msg("unable to start server")
	}
}

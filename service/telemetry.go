package service

import (
	"encoding/json"
	"ntcb-server/dao"
	"ntcb-server/ntcb"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type TelemetryService struct {
	db     *gorm.DB
	logger zerolog.Logger
}

func NewTelemetryService(db *gorm.DB, logger zerolog.Logger) *TelemetryService {
	return &TelemetryService{db: db, logger: logger}
}

func newTelemetryMessage(deviceID string, tm *ntcb.TelemetryMessage) (*dao.TelemetryMessage, error) {
	tmJson, err := json.Marshal(tm)
	if err != nil {
		return nil, err
	}
	return &dao.TelemetryMessage{
		DeviceID:          deviceID,
		SeqNo:             tm.SeqNo,
		Timestamp:         time.Unix(int64(tm.Timestamp), 0),
		EventCode:         tm.EventCode,
		Status:            tm.Status,
		Alarming:          tm.Type == ntcb.MessageTypeAlarming,
		NavValid:          tm.IsNavStatusValid(),
		NavSatelliteCount: tm.GetNavStatusSatelliteCount(),
		NavTimestamp:      time.Unix(int64(tm.LatValidNavTimestamp), 0),
		Lon:               float64(tm.LastValidLon) / 600000.0,
		Lat:               float64(tm.LastValidLat) / 600000.0,
		Alt:               float64(tm.LastValidAlt) / 10.0,
		Speed:             float32(tm.GetSpeed()),
		Direction:         float32(tm.Direction),
		Odometer:          tm.CANOdometer,
		EngineRPM:         tm.GetEngineRPM(),
		IgnitionOn:        tm.IgnitionOn(),
		FuelLevelLiters:   tm.GetFuelLevelLiters(),
		EngineTemp:        tm.CANEngineCoolerTemp,
		AccelPosition:     tm.CANAccelerometerPosition,
		BrakePosition:     tm.CANBrakePosition,
		DistUntilService:  float32(tm.CANDistanceUntilService) * 5,
		Details:           string(tmJson),
	}, nil
}

func (t *TelemetryService) Save(deviceID string, message *ntcb.TelemetryMessage) error {
	daoMsg, err := newTelemetryMessage(deviceID, message)
	if err != nil {
		return errors.Wrap(err, "unable to create telemetry message")
	}
	if err := t.db.Save(daoMsg).Error; err != nil {
		return errors.Wrap(err, "unable to save message to DB")
	}

	return nil
}

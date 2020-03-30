package dao

import (
	"time"
)

type TelemetryMessage struct {
	DeviceID          string
	SeqNo             uint32
	Timestamp         time.Time
	EventCode         uint16
	Status            uint8
	Alarming          bool
	NavValid          bool
	NavSatelliteCount byte
	NavTimestamp      time.Time
	Lon               float64
	Lat               float64
	Alt               float64
	Speed             float32
	Direction         float32
	Odometer          float32
	EngineRPM         uint16
	IgnitionOn        bool
	FuelLevelLiters   float32
	EngineTemp        int8
	AccelPosition     uint8
	BrakePosition     uint8
	DistUntilService  float32
	Details           string
}

func (TelemetryMessage) TableName() string {
	return "telemetry"
}

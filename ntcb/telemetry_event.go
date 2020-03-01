package ntcb

import (
	"encoding/binary"
	"reflect"
)

type MessageType string

const (
	MessageTypeAlarming MessageType = "alarming"
	MessageTypeArray    MessageType = "array"
	MessageTypeCurrent  MessageType = "current"
)

type TelemetryMessage struct {
	Type MessageType
	RawTelemetryMessage
}

type RawTelemetryMessage struct {
	SeqNo                        uint32
	EventCode                    uint16
	Timestamp                    uint32
	Status                       uint8
	FuncModuleStatus1            uint8
	FuncModuleStatus2            uint8
	GSMLevel                     uint8
	NavStatus                    uint8
	LatValidNavTimestamp         uint32
	LastValidLat                 uint32
	LastValidLon                 uint32
	LastValidAlt                 uint32
	Speed                        float32
	Direction                    uint16
	Odometer                     float32
	LastLegDistance              float32
	LastLegDurationSec           uint16
	LastLegDurationSec2          uint16
	MainBatteryVoltage           uint16
	SecondaryBatteryVoltage      uint16
	AnalogueInVoltage1           uint16
	AnalogueInVoltage2           uint16
	AnalogueInVoltage3           uint16
	AnalogueInVoltage4           uint16
	AnalogueInVoltage5           uint16
	AnalogueInVoltage6           uint16
	AnalogueInVoltage7           uint16
	AnalogueInVoltage8           uint16
	DiscreteSensor1              uint8
	DiscreteSensor2              uint8
	OutputState1                 uint8
	OutputState2                 uint8
	ImpulseCounter1              uint32
	ImpulseCounter2              uint32
	AnalogueSensorFreq1          uint16
	AnalogueSensorFreq2          uint16
	MotoHoursSec                 uint32
	RS485FuelSensor1             uint16
	RS485FuelSensor2             uint16
	RS485FuelSensor3             uint16
	RS485FuelSensor4             uint16
	RS485FuelSensor5             uint16
	RS485FuelSensor6             uint16
	RS232FuelSensor              uint16
	TempDiscreteSensor1          int8
	TempDiscreteSensor2          int8
	TempDiscreteSensor3          int8
	TempDiscreteSensor4          int8
	TempDiscreteSensor5          int8
	TempDiscreteSensor6          int8
	TempDiscreteSensor7          int8
	TempDiscreteSensor8          int8
	CANFuelLevel                 uint16
	CANFuelConsumption           float32
	CanEngineRPM                 uint16
	CANEngineCoolerTemp          int8
	CANOdometer                  float32
	CANAxisLoad1                 uint16
	CANAxisLoad2                 uint16
	CANAxisLoad3                 uint16
	CANAxisLoad4                 uint16
	CANAxisLoad5                 uint16
	CANAccelerometerPosition     uint8
	CANBrakePosition             uint8
	CANEngineLoad                uint8
	CANDieselGasFilterFluidLevel uint8
	CANEngineFullWorkTimeSec     uint32
	CANDistanceUntilService      int16
	CANSpeed                     uint8
}

func (tm *RawTelemetryMessage) GetNavStatusSatelliteCount() uint8 {
	return tm.NavStatus >> 2
}

func (tm *RawTelemetryMessage) IsNavStatusValid() bool {
	return tm.NavStatus&0b00000010 > 0
}

func (tm *RawTelemetryMessage) GetEngineRPM() uint16 {
	if tm.CanEngineRPM == 0xffff {
		return 0
	}

	return tm.CanEngineRPM
}

func (tm *RawTelemetryMessage) GetSpeed() uint8 {
	if tm.CANSpeed == 0xff {
		return 0
	}

	return tm.CANSpeed
}

func (tm *RawTelemetryMessage) IgnitionOn() bool {
	return tm.DiscreteSensor1&1 > 0
}

func (tm *RawTelemetryMessage) GetFuelLevelLiters() float32 {
	if tm.CANFuelLevel == 0x7fff {
		return -1
	}

	if tm.CANFuelLevel&0x8000 == 0 {
		return float32(tm.CANFuelLevel&0x7fff) * 0.1
	}

	return 0
}

func FlexTelemetryMessageSize(ba BitArray) uint16 {
	var te RawTelemetryMessage
	teValue := reflect.ValueOf(te)

	var dataSize uint16
	for i := 0; i < 69; i++ {
		if ba.IsSet(i) {
			dataSize += uint16(binary.Size(teValue.Field(i).Interface()))
		}
	}

	return dataSize
}

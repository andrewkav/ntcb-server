package ntcb

import (
	"bytes"
	"encoding/hex"
	"io"
	"net"
	"reflect"
	"testing"
	"time"
)

func TestTelemetryConversion(t *testing.T) {
	tm := RawTelemetryMessage{CANFuelLevel: 202}

	if tm.GetFuelLevelLiters() != 20.2 {
		t.Errorf("unexpected fuel leve in liters")
	}
}

func TestBitArray_IsSet(t *testing.T) {
	zeroes := []int{7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	ones := []int{0, 1, 2, 3, 4, 5, 6, 20, 21, 22, 23}
	// 11111110 00000000 00001111
	ba := BitArray([]byte{0xfe, 0x00, 0x0f})

	for _, p := range zeroes {
		if ba.IsSet(p) {
			t.Fatalf("expected %d to be set", p)
		}
	}

	for _, p := range ones {
		if !ba.IsSet(p) {
			t.Fatalf("expected %d to be  not set", p)
		}
	}

	if ba.IsSet(24) {
		t.Fail()
	}
}

type faker struct {
	io.ReadWriter
}

func (f faker) Close() error                     { return nil }
func (f faker) LocalAddr() net.Addr              { return nil }
func (f faker) RemoteAddr() net.Addr             { return nil }
func (f faker) SetDeadline(time.Time) error      { return nil }
func (f faker) SetReadDeadline(time.Time) error  { return nil }
func (f faker) SetWriteDeadline(time.Time) error { return nil }

var handshake = "404e5443010000000000000013004c472a3e533a313030303030303030303030303030"

func TestConnHandshake(t *testing.T) {
	var rw = &bytes.Buffer{}

	c := Conn{conn: faker{ReadWriter: rw}}

	handshakeBytes, _ := hex.DecodeString(handshake)
	rw.Write(handshakeBytes)

	err := c.handshake()
	if err != nil {
		t.Errorf("unexpected error during handshake")
	}

	if hex.EncodeToString(rw.Bytes()) != "404e544300000000010000000300455e2a3c53" {
		t.Errorf("unexpected handshake reply, %s", hex.EncodeToString(rw.Bytes()))
	}
}

var flex10TelemetryArray = "7e41010900000000107df74f5e002b7df74f5e1f78ed019440fd00000000001700d104871001d182c409413333f6423205db383619"

func TestHandleMultipleFlexTelemetryMessage(t *testing.T) {
	var rw = &bytes.Buffer{}
	ba, _ := NewBitArrayFromString("111100011110110000110000000010000000000000000000000010111000001100011000")
	c := Conn{conn: faker{ReadWriter: rw}, flexBitField: ba, telemetryMessageChan: make(chan TelemetryMessage, 1)}

	flex10TelemetryArrayBytes, _ := hex.DecodeString(flex10TelemetryArray)

	err := c.handleMultipleFlexTelemetryMessage(MessageTypeArray, flex10TelemetryArrayBytes)
	if err != nil {
		t.Errorf("unexpected error processing flex 1.0 telemetry array")
	}

	if hex.EncodeToString(rw.Bytes()) != "7e4101df" {
		t.Errorf("unexpected handshake reply, %x", rw.Bytes())
	}

	tm := <-c.telemetryMessageChan

	if !reflect.DeepEqual(tm, TelemetryMessage{
		Type: MessageTypeArray,
		RawTelemetryMessage: RawTelemetryMessage{
			SeqNo:                    0x9,
			EventCode:                0x1000,
			Timestamp:                0x5e4ff77d,
			NavStatus:                0x2b,
			LatValidNavTimestamp:     0x5e4ff77d,
			LastValidLat:             0x1ed781f,
			LastValidLon:             0xfd4094,
			Direction:                0x17,
			MainBatteryVoltage:       0x4d1,
			SecondaryBatteryVoltage:  0x1087,
			DiscreteSensor1:          0x1,
			CANFuelLevel:             0x82d1,
			CanEngineRPM:             0x9c4,
			CANEngineCoolerTemp:      65,
			CANOdometer:              123.1,
			CANAccelerometerPosition: 0x32,
			CANBrakePosition:         0x5,
			CANDistanceUntilService:  14555,
			CANSpeed:                 0x36,
		},
	}) {
		t.Errorf("unexpected telemetry message, %#v", tm)
	}

}

var flex10TelemetryMessage = "7e540d0000000d000000001004fd4f5e002b04fd4f5e1f78ed019440fd00000000001700d104871001d182c409413333f6423205db383667"

func TestHandleAlarmingFlexTelemetryMessage(t *testing.T) {
	var rw = &bytes.Buffer{}
	ba, _ := NewBitArrayFromString("111100011110110000110000000010000000000000000000000010111000001100011000")
	c := Conn{conn: faker{ReadWriter: rw}, flexBitField: ba, telemetryMessageChan: make(chan TelemetryMessage, 1)}

	flex10TelemetryMessageBytes, _ := hex.DecodeString(flex10TelemetryMessage)

	err := c.handleSingleFlexTelemetryMessage(MessageTypeAlarming, flex10TelemetryMessageBytes)
	if err != nil {
		t.Errorf("unexpected error processing flex 1.0 telemetry message")
	}

	if hex.EncodeToString(rw.Bytes()) != "7e540d00000090" {
		t.Errorf("unexpected handshake reply, %x", rw.Bytes())
	}

	tm := <-c.telemetryMessageChan

	if !reflect.DeepEqual(tm,
		TelemetryMessage{
			Type: MessageTypeAlarming,
			RawTelemetryMessage: RawTelemetryMessage{
				SeqNo:                    0xd,
				EventCode:                0x1000,
				Timestamp:                0x5e4ffd04,
				NavStatus:                0x2b,
				LatValidNavTimestamp:     0x5e4ffd04,
				LastValidLat:             0x1ed781f,
				LastValidLon:             0xfd4094,
				Direction:                0x17,
				MainBatteryVoltage:       0x4d1,
				SecondaryBatteryVoltage:  0x1087,
				DiscreteSensor1:          0x1,
				CANFuelLevel:             0x82d1,
				CanEngineRPM:             0x9c4,
				CANEngineCoolerTemp:      65,
				CANOdometer:              123.1,
				CANAccelerometerPosition: 0x32,
				CANBrakePosition:         0x5,
				CANDistanceUntilService:  14555,
				CANSpeed:                 0x36,
			},
		}) {
		t.Errorf("unexpected telemetry message, %#v", tm)
	}
}

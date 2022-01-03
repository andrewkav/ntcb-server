package ntcb

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"strings"
	"time"
)

var (
	ErrCheckSumMismatch = DataExchangeError("checksum mismatch")
)

type DataExchangeError string

func (e DataExchangeError) Error() string {
	return string(e)
}

func IsNTCBDataExchangeError(err error) bool {
	_, ok := err.(DataExchangeError)
	return err != nil && ok
}

type ProtocolError string

func (e ProtocolError) Error() string {
	return string(e)
}

type protoMessageType int

const (
	messageTypeHandshake protoMessageType = iota + 1
	messageTypeProtoNegotiation
)

const (
	flexProtocolVersion10 = 10
	flexProtocolVersion20 = 20
)

type Header struct {
	Pre [4]byte
	IDr uint32
	IDs uint32
	N   uint16
	CSd uint8
	CSp uint8
}

func (h Header) xor() uint8 {
	buff := bytes.Buffer{}
	h.CSp = 0
	_ = binary.Write(&buff, binary.LittleEndian, h)
	var s uint8
	for _, b := range buff.Bytes() {
		s ^= b
	}

	return s
}

func xor(d []byte) uint8 {
	var s uint8
	for i := range d {
		s ^= d[i]
	}

	return s
}

type Conn struct {
	debug bool

	conn net.Conn

	proto           uint8
	protoVersion    uint8
	structVersion   uint8
	dataSize        uint8
	flexBitField    BitArray
	flexMessageSize uint16
	id              string
	lastPingAt      time.Time

	telemetryMessageChan chan TelemetryMessage
}

func (c *Conn) DeviceID() string {
	return c.id
}

func (c *Conn) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func parseNTCBHeader(r []byte) (h Header, err error) {
	err = binary.Read(bytes.NewBuffer(r), binary.LittleEndian, &h)
	return
}

func parseNTCBMessageType(r []byte) (protoMessageType, error) {
	switch bs := string(r); {
	case strings.HasPrefix(bs, "*>S:"):
		return messageTypeHandshake, nil
	case strings.HasPrefix(bs, "*>FLEX"):
		return messageTypeProtoNegotiation, nil
	}

	return 0, nil
}

func (c *Conn) writeNTCBReply(h Header, body interface{}) error {
	var bodyBuff bytes.Buffer
	if err := binary.Write(&bodyBuff, binary.LittleEndian, body); err != nil {
		return err
	}

	replHeader := Header{
		Pre: h.Pre,
		IDr: h.IDs,
		IDs: h.IDr,
		N:   uint16(bodyBuff.Len()),
		CSd: xor(bodyBuff.Bytes()),
	}

	replHeader.CSp = replHeader.xor()

	if err := binary.Write(c.conn, binary.LittleEndian, replHeader); err != nil {
		return err
	}
	if _, err := c.conn.Write(bodyBuff.Bytes()); err != nil {
		return err
	}

	if c.debug {
		log.Printf("ntcb: message sent, remoteAddr=%s, deviceID=%s, msg=%x\n", c.RemoteAddr(), c.id, bodyBuff.Bytes())
	}

	return nil
}

func (c *Conn) handleHandshake(h Header, r *bytes.Buffer) (err error) {
	c.id = r.String()[20:]

	return c.writeNTCBReply(h, []byte("*<S"))
}

type protoNegotiationMsg struct {
	Pre             [6]byte
	Protocol        uint8
	ProtocolVersion uint8
	StructVersion   uint8
}

func (c *Conn) handleProtocolNegotiation(h Header, d []byte) (err error) {
	// skip *<FLEX
	bodyBytes := d[6:]

	if bodyBytes[0] != 0xb0 {
		err = fmt.Errorf("unsupported protocol %s", hex.EncodeToString(bodyBytes[:1]))
		return
	}

	c.proto, c.protoVersion, c.structVersion, c.dataSize = bodyBytes[0], bodyBytes[1], bodyBytes[2], bodyBytes[3]
	c.flexBitField = bodyBytes[4:]
	c.flexMessageSize = FlexTelemetryMessageSize(c.flexBitField)

	if c.protoVersion != c.structVersion {
		c.protoVersion, c.structVersion = flexProtocolVersion10, flexProtocolVersion10
	}

	return c.writeNTCBReply(h, protoNegotiationMsg{
		Pre:             [6]byte{'*', '<', 'F', 'L', 'E', 'X'},
		Protocol:        c.proto,
		ProtocolVersion: flexProtocolVersion10,
		StructVersion:   flexProtocolVersion10,
	})
}

func (c *Conn) readNTCBMessage(buf *bufio.Reader) (header Header, typ protoMessageType, messageBytes *bytes.Buffer, err error) {
	messageBytes = &bytes.Buffer{}
	if _, err = io.CopyN(messageBytes, buf, 16); err != nil {
		return
	}

	header, err = parseNTCBHeader(messageBytes.Bytes())
	if err != nil {
		return
	}

	if _, err = io.CopyN(messageBytes, buf, int64(header.N)); err != nil {
		return
	}

	// validate crc
	if header.xor() != header.CSp {
		err = ErrCheckSumMismatch
		return
	}
	if xor(messageBytes.Bytes()[16:]) != header.CSd {
		err = ErrCheckSumMismatch
		return
	}

	typ, err = parseNTCBMessageType(messageBytes.Bytes()[16:])

	return
}

func (c *Conn) handleNTCBMessage(buffReader *bufio.Reader) error {
	header, msgType, msgBuff, err := c.readNTCBMessage(buffReader)
	if err != nil {
		return err
	}

	if c.debug {
		log.Printf("ntcb: message recieved, remoteAddr=%s, deviceID=%s, msg=%x\n", c.RemoteAddr(), c.id, msgBuff.Bytes())
	}

	switch msgType {
	case messageTypeProtoNegotiation:
		// skip NTCB header
		if err := c.handleProtocolNegotiation(header, msgBuff.Bytes()[16:]); err != nil {
			return err
		}
	default:
		if c.debug {
			log.Printf("unrecognized NTCB message, remoteAddr=%s, deviceID=%s, type=%d\n", c.RemoteAddr(), c.id, msgType)
		}

		return nil
	}

	return nil
}

func (c *Conn) writeFlexReply(flexHeader []byte, body interface{}) error {
	var replBuffer = &bytes.Buffer{}
	if _, err := replBuffer.Write(flexHeader); err != nil {
		return err
	}

	if !reflect.ValueOf(body).IsNil() {
		if err := binary.Write(replBuffer, binary.LittleEndian, body); err != nil {
			return err
		}
	}

	crc := CRC8(replBuffer.Bytes())
	if err := replBuffer.WriteByte(crc); err != nil {
		return err
	}

	_, err := c.conn.Write(replBuffer.Bytes())

	if c.debug {
		log.Printf("ntcb: message sent, remoteAddr=%s, deviceID=%s, msg=%x\n", c.RemoteAddr(), c.id, replBuffer.Bytes())
	}
	return err
}

func readTelemetryMessage(r io.Reader, ba BitArray) (*RawTelemetryMessage, error) {
	var te RawTelemetryMessage
	teValue := reflect.ValueOf(&te).Elem()

	for i := 0; i < 122; i++ {
		if ba.IsSet(i) {
			fieldValue := teValue.Field(i)
			if err := binary.Read(r, binary.LittleEndian, fieldValue.Addr().Interface()); err != nil {
				return nil, DataExchangeError(err.Error())
			}
		}
	}

	return &te, nil
}

// handle as single flex message
func (c *Conn) handleSingleFlexTelemetryMessage(typ MessageType, msgBytes []byte) error {
	msgCRC8 := CRC8(msgBytes[:len(msgBytes)-1])
	if msgCRC8 != msgBytes[len(msgBytes)-1] {
		return ErrCheckSumMismatch
	}

	flexHeader := msgBytes[:2]

	msgBuff := bytes.NewBuffer(msgBytes[2:])
	var eventIndex *uint32

	if typ != MessageTypeCurrent {
		var ei uint32
		eventIndex = &ei
		if err := binary.Read(msgBuff, binary.LittleEndian, eventIndex); err != nil {
			return err
		}
	}

	te, err := readTelemetryMessage(msgBuff, c.flexBitField)
	if err != nil {
		return err
	}

	if c.telemetryMessageChan != nil {
		c.telemetryMessageChan <- TelemetryMessage{Type: typ, RawTelemetryMessage: *te}
	}

	return c.writeFlexReply(flexHeader, eventIndex)
}

func (c *Conn) handleMultipleFlexTelemetryMessage(typ MessageType, msgBytes []byte) error {
	msgCRC8 := CRC8(msgBytes[:len(msgBytes)-1])
	if msgCRC8 != msgBytes[len(msgBytes)-1] {
		return ErrCheckSumMismatch
	}

	flexHeader := msgBytes[0:2]

	msgReader := bytes.NewBuffer(msgBytes[2:])
	var msgCount byte
	if err := binary.Read(msgReader, binary.LittleEndian, &msgCount); err != nil {
		return err
	}

	for i := 0; i < int(msgCount); i++ {
		te, err := readTelemetryMessage(msgReader, c.flexBitField)
		if err != nil {
			return err
		}

		if c.telemetryMessageChan != nil {
			c.telemetryMessageChan <- TelemetryMessage{Type: typ, RawTelemetryMessage: *te}
		}
	}

	return c.writeFlexReply(flexHeader, &msgCount)
}

func (c *Conn) handleFlexMessage(buf *bufio.Reader) error {
	fp, err := buf.Peek(2)
	if err != nil {
		return err
	}

	switch string(fp) {
	case "~T":
		var msgBuff = &bytes.Buffer{}
		// header (2) + event index (4) + flex message (N) + crc (1)
		_, err := io.CopyN(msgBuff, buf, 2+4+int64(c.flexMessageSize)+1)
		if err != nil {
			return err
		}

		if c.debug {
			log.Printf("ntcb: message recieved, remoteAddr=%s, deviceID=%s, msg=%x\n", c.RemoteAddr(), c.id, msgBuff.Bytes())
		}

		if err := c.handleSingleFlexTelemetryMessage(MessageTypeAlarming, msgBuff.Bytes()); err != nil {
			return err
		}
	case "~C":
		var msgBuff = &bytes.Buffer{}
		// header (2) + flex message (N) + crc (1)
		_, err := io.CopyN(msgBuff, buf, 2+int64(c.flexMessageSize)+1)
		if err != nil {
			return err
		}

		if c.debug {
			log.Printf("ntcb: message recieved, remoteAddr=%s, deviceID=%s, msg=%x\n", c.RemoteAddr(), c.id, msgBuff.Bytes())
		}

		if err := c.handleSingleFlexTelemetryMessage(MessageTypeCurrent, msgBuff.Bytes()); err != nil {
			return err
		}
	case "~A":
		var msgBuff = &bytes.Buffer{}
		prefixBytes, err := buf.Peek(3)
		if err != nil {
			return err
		}
		// header (2) + size (1) + flex message (N) * size + crc (1)
		_, err = io.CopyN(msgBuff, buf, 2+1+int64(c.flexMessageSize)*int64(prefixBytes[2])+1)
		if err != nil {
			return err
		}

		if c.debug {
			log.Printf("ntcb: message recieved, remoteAddr=%s, deviceID=%s, msg=%x\n", c.RemoteAddr(), c.id, msgBuff.Bytes())
		}

		if err := c.handleMultipleFlexTelemetryMessage(MessageTypeArray, msgBuff.Bytes()); err != nil {
			return err
		}

	default:
		if _, err := buf.Discard(buf.Buffered()); err != nil {
			return err
		}
	}
	return nil
}

func (c *Conn) handshake() error {
	// wait 30 sec for the first message
	if err := c.conn.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
		return err
	}

	buffReader := bufio.NewReader(c.conn)
	header, msgType, msgBuff, err := c.readNTCBMessage(buffReader)
	if err != nil {
		if err == io.EOF {
			return ProtocolError("handshake: unexpected end of file")
		}
		return err
	}
	if msgType != messageTypeHandshake {
		return ProtocolError("handshake: invalid message type")
	}

	if c.debug {
		log.Printf("ntcb: handshake, remoteAddr=%s, body=%x\n", c.RemoteAddr(), msgBuff.Bytes())
	}

	return c.handleHandshake(header, msgBuff)
}

func (c *Conn) readLoop() error {
	buffReader := bufio.NewReader(c.conn)

	if err := c.conn.SetReadDeadline(time.Time{}); err != nil {
		return err
	}

	for {
		b, err := buffReader.Peek(1)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		switch b[0] {
		// NTCB message header
		case '@':
			if err := c.handleNTCBMessage(buffReader); err != nil {
				if IsNTCBDataExchangeError(err) {
					log.Printf("ntcb: data exchange error has occorred,  remoteAddr=%s, deviceID=%s, err=%v\n", c.RemoteAddr(), c.id, err)
					continue
				}
				return err
			}
		// FLEX ping
		case 0x7F:
			if c.debug {
				log.Printf("ntcb: ping, remoteAddr=%s, deviceID=%s\n", c.conn.RemoteAddr().String(), c.id)
			}

			_, err = buffReader.Discard(1)
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
			c.lastPingAt = time.Now()
			continue
		// FLEX message
		case '~':
			if err := c.handleFlexMessage(buffReader); err != nil {
				if IsNTCBDataExchangeError(err) {
					log.Printf("ntcb: data exchange error has occorred,  remoteAddr=%s, deviceID=%s, err=%v\n", c.RemoteAddr(), c.id, err)
					continue
				}
				return err
			}

		default:
			return nil
		}
	}
}

func (c *Conn) Close() error {
	if c.telemetryMessageChan != nil {
		close(c.telemetryMessageChan)
	}
	return c.conn.Close()
}

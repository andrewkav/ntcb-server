package ntcb

import (
	"net"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"
)

type ServerOptions struct {
	Debug              bool
	Address            string
	OnTelemetryMessage func(c *Conn, tm TelemetryMessage)
	OnNewConnection    func(c *Conn)
	OnConnectionClosed func(c *Conn, err error)
}

type Server struct {
	opts   ServerOptions
	conns  map[string]*Conn
	connMu sync.Mutex
	close  chan struct{}
}

func (s *Server) ActiveDeviceIDs() []string {
	IDs := make([]string, 0, len(s.conns))
	for ID := range s.conns {
		IDs = append(IDs, ID)
	}

	sort.Strings(IDs)

	return IDs
}

func (s *Server) handleNewConnection(conn net.Conn) *Conn {
	c := &Conn{
		debug:                s.opts.Debug,
		conn:                 conn,
		telemetryMessageChan: make(chan TelemetryMessage, 128),
	}

	go func() {
		for m := range c.telemetryMessageChan {
			s.opts.OnTelemetryMessage(c, m)
		}
	}()

	go func() {

		var connErr error

		defer func() {
			s.connMu.Lock()
			delete(s.conns, c.DeviceID())
			s.connMu.Unlock()

			_ = c.Close()

			if s.opts.OnConnectionClosed != nil {
				s.opts.OnConnectionClosed(c, connErr)
			}
		}()

		if connErr = c.handshake(); connErr != nil {
			return
		}

		s.connMu.Lock()
		s.conns[c.DeviceID()] = c
		s.connMu.Unlock()
		if s.opts.OnNewConnection != nil {
			s.opts.OnNewConnection(c)
		}

		if connErr = c.readLoop(); connErr != nil {
			return
		}

	}()

	return c
}

func (s *Server) ListenAndServe() error {
	l, err := net.Listen("tcp", s.opts.Address)
	if err != nil {
		return err
	}

	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	newConn := make(chan net.Conn)

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				time.Sleep(time.Second)
				continue
			}

			newConn <- conn
		}
	}()

	for {
		select {
		case <-interrupt:
			return nil
		case <-s.close:
			return nil
		case conn := <-newConn:
			s.handleNewConnection(conn)
		}
	}
}

func (s *Server) Stop() {
	s.close <- struct{}{}
}

func NewServer(options ServerOptions) *Server {
	return &Server{
		opts:  options,
		close: make(chan struct{}),
		conns: make(map[string]*Conn, 16),
	}
}

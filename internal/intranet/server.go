package intranet

import (
	"fmt"
	"net"
	"strings"
)

type server struct {
	port     int
	listener *net.TCPListener
	callback func([]byte, chan error)
	conns    []net.Conn
	errorCh  chan error
}

func NewServer() (Server, error) {
	serv := &server{}
	err := serv.buildTCP()
	if err != nil {
		return nil, err
	}

	return serv, nil
}

func (s *server) GetStruct() *server {
	return s
}

func (s *server) buildTCP() error {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	s.port = l.Addr().(*net.TCPAddr).Port
	s.listener = l
	s.errorCh = make(chan error, 1)
	return nil
}

func (s *server) SetRecvCallBack(cb func([]byte, chan error)) {
	s.callback = cb
}

func (s *server) GetPort() int {
	return s.port
}

func (s *server) Listen() {
	go func(s *server) {
		defer s.listener.Close()
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				s.errorCh <- fmt.Errorf("tcp accept error: %s", err)
				return
			} else {
				s.conns = append(s.conns, conn)
				go s.handleRecv(conn)
			}
		}
	}(s)
}

func (s *server) Done() {
	for _, conn := range s.conns {
		conn.Close()
	}
}

func (s *server) handleRecv(conn net.Conn) {
	defer conn.Close()
	for {
		// Make a buffer to hold incoming data.
		buf := make([]byte, 1024)
		// Read the incoming connection into the buffer.
		_, err := conn.Read(buf)
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				s.errorCh <- fmt.Errorf("recv error: %s", err)
			}
		} else {
			// Set a callback to process the data
			s.callback(buf, s.errorCh)
		}
	}
}

func (s *server) Send(request []byte) {
	for _, conn := range s.conns {
		conn.Write(request)
	}
}

func (s *server) GetErrorChannel() chan error {
	return s.errorCh
}

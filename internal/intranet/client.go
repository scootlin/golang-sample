package intranet

import (
	"fmt"
	"io"
	"net"
)

type client struct {
	conn     net.Conn
	callback func([]byte, chan error)
	errorCh  chan error
}

const hostName = "localhost"

func NewClient(port int) (Client, error) {
	addr := fmt.Sprintf("%s:%d", hostName, port)
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	cli := &client{}
	cli.conn = c
	cli.errorCh = make(chan error, 1)
	return cli, err
}

func (c *client) Run() {
	go func(c *client) {
		defer c.conn.Close()
		for {
			buf := make([]byte, 1024)
			_, err := c.conn.Read(buf)
			if err != nil {
				if err == io.EOF {
					c.conn.Close()
				} else {
					c.errorCh <- err
				}
				return
			} else {
				c.callback(buf, c.errorCh)
			}
		}
	}(c)

}

func (c *client) SetRecvCallBack(cb func([]byte, chan error)) {
	c.callback = cb
}

func (c *client) Send(data []byte) error {
	_, err := c.conn.Write(data)
	return err
}

func (c *client) GetErrorChannel() chan error {
	return c.errorCh
}

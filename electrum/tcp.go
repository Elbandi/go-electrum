package electrum

import (
	"bufio"
	"crypto/tls"
	"log"
	"net"
)

var (
	NetDial = net.Dial
)

type TCPTransport struct {
	conn      net.Conn
	responses chan []byte
	errors    chan error
}

func NewTCPTransport(addr string) (*TCPTransport, error) {
	conn, err := NetDial("tcp", addr)
	if err != nil {
		return nil, err
	}
	t := &TCPTransport{
		conn:      conn,
		responses: make(chan []byte),
		errors:    make(chan error),
	}
	go t.listen()
	return t, nil
}

func NewSSLTransport(addr string, config *tls.Config) (*TCPTransport, error) {
	conn, err := NetDial("tcp", addr)
	if err != nil {
		return nil, err
	}
	tlsConn := tls.Client(conn, config)
	if err = tlsConn.Handshake(); err != nil {
		return nil, err
	}
	t := &TCPTransport{
		conn:      tlsConn,
		responses: make(chan []byte),
		errors:    make(chan error),
	}
	go t.listen()
	return t, nil
}

func (t *TCPTransport) SendMessage(body []byte) error {
	log.Printf("%s <- %s", t.conn.RemoteAddr(), body)
	_, err := t.conn.Write(body)
	return err
}

const delim = byte('\n')

func (t *TCPTransport) listen() {
	defer t.conn.Close()
	reader := bufio.NewReader(t.conn)
	for {
		line, err := reader.ReadBytes(delim)
		if err != nil {
			t.errors <- err
			log.Printf("error %s", err)
			break
		}
		log.Printf("%s -> %s", t.conn.RemoteAddr(), line)
		t.responses <- line
	}
}

func (t *TCPTransport) Responses() <-chan []byte {
	return t.responses
}
func (t *TCPTransport) Errors() <-chan error {
	return t.errors
}

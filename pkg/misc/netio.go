package misc

import (
	"net"

	"github.com/ksrzmv/xch/pkg/message"
)

const (
	BUFFER_SIZE = 1024
)

func ReadMessageFrom(conn net.Conn) (*message.Message, error) {
	m := message.Init()

	// read client's input
	readBuf := make([]byte, BUFFER_SIZE)
	n, err := conn.Read(readBuf)
	if err != nil {
		return nil, err
	}

	// trims b to size n
	readBuf, err = Trim(readBuf, n)
	if err != nil {
		return nil, err
	}

	// unmarshal input message
	m, err = message.FromJson([]byte(readBuf))
	if err != nil {
		return nil, err
	}
	return m, nil
}

func SendMessageTo(conn net.Conn, m *message.Message) error {
	sendBuf, err := m.ToJson()
	if err != nil {
		return err
	}
	conn.Write(sendBuf)
	if err != nil {
		return err
	}
	return nil
}

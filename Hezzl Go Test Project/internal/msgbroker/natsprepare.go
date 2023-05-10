package msgbroker

import (
	"fmt"
	"github.com/nats-io/nats.go"
)

type NatsConn struct {
	conn *nats.Conn
}

func (n *NatsConn) Close() {
	n.conn.Close()
}

func (n *NatsConn) PublishMessage(ip, apiType, apiAddr string, errIn error) error {
	var err error

	err = n.conn.Publish("list", []byte(fmt.Sprintf("IP: %v | Request type: %v | Request addr: %v | Error: %v", ip, apiType, apiAddr, errIn)))

	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func PrepareNats() (*NatsConn, error) {
	var err error
	c, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	n := &NatsConn{conn: c}
	return n, nil
}

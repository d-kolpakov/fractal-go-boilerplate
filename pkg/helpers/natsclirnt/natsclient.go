package natsclient

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/nats-io/stan.go"
)

type NatsOptions struct {
	Cluster                string
	ClientId               string
	Host                   string
	ConnectionLostCallback func(c stan.Conn, e error)
	QueueGroup             string
}

type NatsConnection struct {
	c  stan.Conn
	no *NatsOptions
}

func New(no *NatsOptions) (*NatsConnection, error) {
	clusterName := no.Cluster
	clientId := no.ClientId + "-" + no.randomString(5)
	hostUrl := no.Host
	conn, err := stan.Connect(
		clusterName,
		clientId,
		stan.NatsURL(hostUrl),
		stan.Pings(10, 3),
		stan.SetConnectionLostHandler(no.ConnectionLostCallback),
	)
	if err != nil {
		return nil, err
	}
	nc := &NatsConnection{c: conn, no: no}

	return nc, nil
}

func (nc *NatsConnection) SendMessage(subj string, msg []byte) error {
	err := nc.c.Publish(subj, msg)
	return err
}

func (nc *NatsConnection) QueueSubscribe(subj string, f func(m *stan.Msg)) error {
	_, err := nc.c.QueueSubscribe(subj, nc.no.QueueGroup, f, stan.DurableName(nc.no.QueueGroup))
	return err
}

func (no *NatsOptions) randomString(size int) string {
	rb := make([]byte, size)
	rand.Read(rb)
	rs := base64.URLEncoding.EncodeToString(rb)
	rs = rs[0 : len(rs)-1]

	return rs
}

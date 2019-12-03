package tunnel

import (
	"fmt"
	"net"
	"strconv"

	"github.com/jedevc/AppArea/helpers"
	"golang.org/x/crypto/ssh"
)

type Forwarder struct {
	Request ForwardRequest
	conn    *ssh.ServerConn
}

func NewForwarder(conn *ssh.ServerConn, req ForwardRequest) Forwarder {
	return Forwarder{
		Request: req,
		conn:    conn,
	}
}

func (f *Forwarder) Connect() (Tunnel, error) {
	remoteAddress, remotePortStr, _ := net.SplitHostPort(f.conn.RemoteAddr().String())
	remotePort, _ := strconv.Atoi(remotePortStr)

	data := make([]byte, 0)
	helpers.PackString(&data, f.Request.Address)
	helpers.PackInt(&data, f.Request.Port)
	helpers.PackString(&data, remoteAddress)
	helpers.PackInt(&data, uint32(remotePort))

	ch, reqs, err := f.conn.OpenChannel("forwarded-tcpip", data)
	if err != nil {
		return nil, fmt.Errorf("could not open channel (is the port open?)")
	}
	go ssh.DiscardRequests(reqs)

	return ch, nil
}

type ForwardRequest struct {
	Address string
	Port    uint32
}

func ParseForwardRequest(payload []byte) (ForwardRequest, error) {
	req := ForwardRequest{}

	address, err := helpers.UnpackString(&payload)
	if err != nil {
		return req, err
	}
	req.Address = address

	port, err := helpers.UnpackInt(&payload)
	if err != nil {
		return req, err
	}
	req.Port = port

	if len(payload) != 0 {
		return req, fmt.Errorf("forward request parse error: unknown excess data")
	}

	return req, nil
}

package forward

import "io"

type Forwarder interface {
	Serve() error
	Close()
	AttachClientLog(io.Writer)

	ListenerAddress() string
	ListenerPort() uint32
}

package forward

type Forwarder interface {
	Serve() error
	Close()

	ListenerAddress() string
	ListenerPort() uint32
}

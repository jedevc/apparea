package forward

type Forwarder interface {
	ListenAndServe() error
	Close()

	ListenerAddress() string
}

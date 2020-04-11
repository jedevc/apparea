package forward

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync"

	"github.com/jedevc/AppArea/helpers"
	"golang.org/x/crypto/ssh"
)

type RawForwarder struct {
	Request  ForwardRequest
	baseConn *ssh.ServerConn

	lock     sync.Mutex
	closed   bool
	listener net.Listener
}

func NewRawForwarder(conn *ssh.ServerConn, req ForwardRequest) *RawForwarder {
	return &RawForwarder{
		Request:  req,
		baseConn: conn,
	}
}

// FIXME: type here is not right
func (f RawForwarder) connect() (io.ReadWriteCloser, error) {
	remoteAddress, remotePortStr, _ := net.SplitHostPort(f.baseConn.RemoteAddr().String())
	remotePort, _ := strconv.Atoi(remotePortStr)

	data := make([]byte, 0)
	helpers.PackString(&data, f.Request.Host)
	helpers.PackInt(&data, f.Request.Port)
	helpers.PackString(&data, remoteAddress)
	helpers.PackInt(&data, uint32(remotePort))

	ch, reqs, err := f.baseConn.OpenChannel("forwarded-tcpip", data)
	if err != nil {
		return nil, fmt.Errorf("could not open channel (is the port open?)")
	}
	go ssh.DiscardRequests(reqs)

	return ch, nil
}

func (f *RawForwarder) ListenAndServe() error {
	ln, err := net.Listen("tcp", f.Request.Address())
	if err != nil {
		return fmt.Errorf("Could not listen on %s", f.Request.Address())
	}
	f.listener = ln

	go func() {
		for {
			f.lock.Lock()
			if f.closed {
				f.lock.Unlock()
				break
			}
			f.lock.Unlock()

			incoming, err := f.listener.Accept()
			if err != nil {
				log.Printf("Could not accept connection")
				continue
			}

			outgoing, err := f.connect()
			if err != nil {
				log.Print("Could not open remote connection")
				incoming.Close()
				continue
			}
			closer := func() {
				incoming.Close()
				outgoing.Close()
			}

			var once sync.Once
			go func() {
				io.Copy(incoming, outgoing)
				once.Do(closer)
			}()
			go func() {
				io.Copy(outgoing, incoming)
				once.Do(closer)
			}()
		}
	}()

	return nil
}

func (f *RawForwarder) Close() {
	f.lock.Lock()
	f.closed = true
	if f.listener != nil {
		f.listener.Close()
	}
	f.lock.Unlock()
}

func (f *RawForwarder) ListenerAddress() string {
	if f.listener == nil {
		return ""
	}
	return f.listener.Addr().String()
}
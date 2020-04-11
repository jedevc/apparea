package forward

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jedevc/AppArea/helpers"
	"golang.org/x/crypto/ssh"
)

type HTTPForwarder struct {
	Request  ForwardRequest
	Hostname string

	connector *ssh.ServerConn
}

var httpServer *http.Server = nil
var httpMap map[string]*HTTPForwarder
var httpLock sync.Mutex

func NewHTTPForwarder(hostname string, conn *ssh.ServerConn, req ForwardRequest) *HTTPForwarder {
	return &HTTPForwarder{
		Request:   req,
		Hostname:  hostname,
		connector: conn,
	}
}

// FIXME: type here
func (f HTTPForwarder) connect() (io.ReadWriteCloser, error) {
	remoteAddress, remotePortStr, _ := net.SplitHostPort(f.connector.RemoteAddr().String())
	remotePort, _ := strconv.Atoi(remotePortStr)

	data := make([]byte, 0)
	helpers.PackString(&data, f.Request.Host)
	helpers.PackInt(&data, f.Request.Port)
	helpers.PackString(&data, remoteAddress)
	helpers.PackInt(&data, uint32(remotePort))

	ch, reqs, err := f.connector.OpenChannel("forwarded-tcpip", data)
	if err != nil {
		return nil, fmt.Errorf("could not open channel (is the port open?)")
	}
	go ssh.DiscardRequests(reqs)

	return ch, nil
}

func (f *HTTPForwarder) ListenAndServe() error {
	handler := func(w http.ResponseWriter, r *http.Request) {
		httpLock.Lock()
		fr, ok := httpMap[r.Host]
		httpLock.Unlock()

		if !ok {
			w.WriteHeader(404)
			fmt.Fprintf(w, "site not found")
			return
		}

		fr.handle(w, r)
	}

	httpLock.Lock()

	if httpServer == nil {
		httpServer = &http.Server{
			Addr:           ":8080",
			Handler:        http.HandlerFunc(handler),
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		// TODO: check for errors on listening
		go httpServer.ListenAndServe()

		httpMap = make(map[string]*HTTPForwarder)
	}

	if _, ok := httpMap[f.Hostname]; ok {
		httpLock.Unlock()
		return fmt.Errorf("site name already in use")
	}

	httpMap[f.Hostname] = f

	httpLock.Unlock()

	return nil
}

func (f *HTTPForwarder) Close() {
	httpLock.Lock()
	delete(httpMap, f.Hostname)

	if len(httpMap) == 0 {
		httpServer.Close()
		httpServer = nil
	}
	httpLock.Unlock()
}

func (f *HTTPForwarder) ListenerAddress() string {
	httpLock.Lock()
	_, ok := httpMap[f.Hostname]
	httpLock.Unlock()
	if !ok {
		return ""
	}

	parts := strings.Split(httpServer.Addr, ":")
	if len(parts) == 2 {
		return f.Hostname + ":" + parts[1]
	} else {
		return f.Hostname
	}
}

func (f *HTTPForwarder) ListenerPort() uint32 {
	parts := strings.Split(httpServer.Addr, ":")
	if len(parts) == 2 {
		port, _ := strconv.Atoi(parts[1])
		return uint32(port)
	} else {
		return 80
	}
}

func (f HTTPForwarder) handle(w http.ResponseWriter, r *http.Request) error {
	// TODO: reuse connections if possible?
	tunn, err := f.connect()
	if err != nil {
		return nil
	}

	r.Write(tunn)
	io.Copy(w, tunn)

	return nil
}

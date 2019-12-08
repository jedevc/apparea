package tunnel

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/jedevc/AppArea/helpers"
	"golang.org/x/crypto/ssh"
)

type Forwarder interface {
	ListenAndServe() error
}

type RawForwarder struct {
	Request   ForwardRequest
	connector *ssh.ServerConn
}

func NewRawForwarder(conn *ssh.ServerConn, req ForwardRequest) RawForwarder {
	return RawForwarder{
		Request:   req,
		connector: conn,
	}
}

func (f RawForwarder) connect() (Tunnel, error) {
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

func (f RawForwarder) ListenAndServe() error {
	ln, err := net.Listen("tcp", f.Request.Address())
	if err != nil {
		return fmt.Errorf("Could not listen on %s", f.Request.Address())
	}

	for {
		out, err := ln.Accept()
		if err != nil {
			log.Printf("Could not accept connection")
			continue
		}

		in, err := f.connect()
		if err != nil {
			log.Print("Could not open remote connection")
			out.Close()
			continue
		}
		closer := func() {
			in.Close()
			out.Close()
		}

		var once sync.Once
		go func() {
			io.Copy(in, out)
			once.Do(closer)
		}()
		go func() {
			io.Copy(out, in)
			once.Do(closer)
		}()

	}

	// FIXME: listener is never closed
	return nil
}

type HTTPForwarder struct {
	Request   ForwardRequest
	connector *ssh.ServerConn
}

var httpServer *http.Server = nil
var httpMap map[string]HTTPForwarder

func NewHTTPForwarder(conn *ssh.ServerConn, req ForwardRequest) HTTPForwarder {
	return HTTPForwarder{
		Request:   req,
		connector: conn,
	}
}

func (f HTTPForwarder) connect() (Tunnel, error) {
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

func (f HTTPForwarder) ListenAndServe() error {
	handler := func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		fr, ok := httpMap[host]
		if !ok {
			w.WriteHeader(404)
			fmt.Fprintf(w, "site not found")
			return
		}

		fr.handle(w, r)
	}

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

		httpMap = make(map[string]HTTPForwarder)
	}

	hostname := f.connector.User() + ".apparea.dev"
	if _, ok := httpMap[hostname]; ok {
		return fmt.Errorf("site name already in use")
	}

	httpMap[hostname] = f
	return nil
}

func (f HTTPForwarder) handle(w http.ResponseWriter, r *http.Request) error {
	// reuse connections if possible?
	tunn, err := f.connect()
	if err != nil {
		return nil
	}

	r.Write(tunn)
	io.Copy(w, tunn)

	return nil
}

type ForwardRequest struct {
	Host string
	Port uint32
}

func ParseForwardRequest(payload []byte) (ForwardRequest, error) {
	req := ForwardRequest{}

	host, err := helpers.UnpackString(&payload)
	if err != nil {
		return req, err
	}
	req.Host = host

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

func (fr ForwardRequest) Address() string {
	return fr.Host + ":" + strconv.FormatUint(uint64(fr.Port), 10)
}

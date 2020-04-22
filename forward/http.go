package forward

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jedevc/apparea/helpers"
	"golang.org/x/crypto/ssh"
)

type HTTPForwarder struct {
	Request  ForwardRequest
	Hostname string

	clientLog io.Writer

	connector *ssh.ServerConn
	useTLS    bool
}

var httpMap = make(map[string]*HTTPForwarder)
var httpLock sync.Mutex
var httpServer *http.Server

func httpHandler(w http.ResponseWriter, r *http.Request) {
	httpLock.Lock()
	fr, ok := httpMap[r.Host]
	httpLock.Unlock()

	if !ok {
		w.WriteHeader(404)
		fmt.Fprintf(w, "site not found")
		return
	}

	err := fr.handle(w, r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}
}

func ServeHTTP(address string) error {
	httpServer = &http.Server{
		Addr:           address,
		Handler:        http.HandlerFunc(httpHandler),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Listening for HTTP connections on %s...", address)
	return httpServer.ListenAndServe()
}

func NewHTTPForwarder(hostname string, conn *ssh.ServerConn, req ForwardRequest) *HTTPForwarder {
	return &HTTPForwarder{
		Request:   req,
		Hostname:  hostname,
		clientLog: ioutil.Discard,
		connector: conn,
	}
}

func (f *HTTPForwarder) UseTLS(use bool) *HTTPForwarder {
	f.useTLS = use
	return f
}

func (f *HTTPForwarder) AttachClientLog(w io.Writer) {
	f.clientLog = w
}

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
		return nil, fmt.Errorf("could not open channel: %w", err)
	}
	go ssh.DiscardRequests(reqs)

	if f.useTLS {
		res := NewTLSWrapper(ch)
		return res, nil
	}

	return ch, nil
}

func (f *HTTPForwarder) Serve() error {
	httpLock.Lock()
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
	httpLock.Unlock()
}

func (f *HTTPForwarder) ListenerAddress() string {
	return "http://" + f.Hostname
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
	now := time.Now()

	// connect back
	tunn, err := f.connect()
	if err != nil {
		return err
	}
	defer tunn.Close()

	// forward request
	err = r.Write(tunn)
	if err != nil {
		return fmt.Errorf("could not forward request: %w", err)
	}

	// read response
	resp, err := http.ReadResponse(bufio.NewReader(tunn), r)
	if err != nil {
		return fmt.Errorf("could not read back response: %w", err)
	}
	defer resp.Body.Close()

	fmt.Fprintf(f.clientLog, "%s [%d] %s %s\n", now.Format("2006/01/02 15:04:05"), resp.StatusCode, r.Method, r.URL.Path)

	// copy to response
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Set(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return fmt.Errorf("could not write copy response: %w", err)
	}

	return nil
}

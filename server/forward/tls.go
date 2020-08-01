package forward

import (
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
)

type TLSWrapper struct {
	channel io.ReadWriteCloser
	inConn  net.Conn
	outConn net.Conn
	tlsConn net.Conn
}

func NewTLSWrapper(ch io.ReadWriteCloser) TLSWrapper {
	inConn, outConn := net.Pipe()

	// we log errors cause there's not really anything else we can do with them
	go func() {
		// writing to client
		_, err := io.Copy(ch, inConn)
		if err != nil && !errors.Is(err, io.EOF) {
			log.Printf("could not copy to client: %s", err)
		}
	}()
	go func() {
		// reading from client
		_, err := io.Copy(inConn, ch)
		if err != nil && !errors.Is(err, io.EOF) {
			log.Printf("could not copy to client: %s", err)
		}
	}()

	fin := tls.Client(outConn, &tls.Config{
		InsecureSkipVerify: true,
	})

	return TLSWrapper{
		channel: ch,
		inConn:  inConn,
		outConn: outConn,
		tlsConn: fin,
	}
}

func (wrap TLSWrapper) Read(p []byte) (int, error) {
	return wrap.tlsConn.Read(p)
}

func (wrap TLSWrapper) Write(p []byte) (int, error) {
	return wrap.tlsConn.Write(p)
}

func (wrap TLSWrapper) Close() error {
	// we don't really care about errors here, let's just say they all succeed
	wrap.channel.Close()
	wrap.tlsConn.Close()
	wrap.inConn.Close()
	wrap.outConn.Close()

	return nil
}

package tunnel

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/jedevc/AppArea/helpers"
	"golang.org/x/crypto/ssh"
)

func Run(address string, config *ssh.ServerConfig) <-chan *Session {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen on %s (%s)", address, err)
	}

	sessions := make(chan *Session)

	log.Printf("Listening on %s...", address)
	go func() {
		for {
			tcpConn, err := listener.Accept()
			if err != nil {
				log.Printf("Failed to accept incoming connection (%s)", err)
				continue
			}

			sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, config)
			if err != nil {
				log.Printf("Failed to handshake (%s)", err)
				continue
			}

			session := launchSession(sshConn, chans, reqs)

			sessions <- session
		}
	}()

	return sessions
}

func launchSession(conn *ssh.ServerConn, chans <-chan ssh.NewChannel, reqs <-chan *ssh.Request) *Session {
	log.Printf("New SSH connection from %s (%s)", conn.RemoteAddr(), conn.ClientVersion())

	views := make(chan View)
	forwards := make(chan Forwarder)
	session := NewSession(views, forwards)

	var closer sync.Once

	go func() {
		for req := range reqs {
			if req.Type == "tcpip-forward" {
				forward, err := handleTCPIP(conn, req)
				if err != nil {
					log.Printf("internal error: %s", err)
					continue
				}
				forwards <- forward
			} else {
				// discard request
				if req.WantReply {
					req.Reply(false, nil)
				}
			}
		}

		closer.Do(func() {
			close(forwards)
			close(views)
		})
	}()
	go func() {
		for newChannel := range chans {
			if t := newChannel.ChannelType(); t == "session" {
				view, err := handleSessionChannel(conn, newChannel)
				if err != nil {
					log.Printf("internal error: %s", err)
					continue
				}
				views <- view
			} else {
				newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
			}
		}

		closer.Do(func() {
			close(forwards)
			close(views)
		})
	}()

	return session
}

func handleSessionChannel(conn *ssh.ServerConn, newChannel ssh.NewChannel) (View, error) {
	session, requests, err := newChannel.Accept()
	if err != nil {
		return nil, err
	}

	go func() {
		for req := range requests {
			switch req.Type {
			case "shell":
				if len(req.Payload) == 0 {
					req.Reply(true, nil)
				}
			}
		}
	}()

	return NewStatusView(session), nil
}

func handleTCPIP(conn *ssh.ServerConn, req *ssh.Request) (Forwarder, error) {
	fr, err := ParseForwardRequest(req.Payload)
	if err != nil {
		if req.WantReply {
			req.Reply(false, nil)
		}
		return nil, err
	}

	bs := make([]byte, 0)
	helpers.PackInt(&bs, fr.Port)
	req.Reply(true, bs)

	var fwd Forwarder
	if fr.Port == 80 {
		fwd = NewHTTPForwarder(conn, fr)
	} else {
		fwd = NewRawForwarder(conn, fr)
	}
	return fwd, nil
}

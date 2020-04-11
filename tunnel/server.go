package tunnel

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/jedevc/AppArea/config"
	"github.com/jedevc/AppArea/forward"
	"github.com/jedevc/AppArea/helpers"
	"golang.org/x/crypto/ssh"
)

type Server struct {
	Config   *config.Config
	Hostname string
}

func (server *Server) Run(address string) <-chan *Session {
	if server.Config == nil {
		log.Fatalf("Internal error: no config provided")
	}

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

			sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, server.Config.SSHConfig)
			if err != nil {
				log.Printf("Failed to handshake (%s)", err)
				continue
			}

			session := server.launchSession(sshConn, chans, reqs)

			sessions <- session
		}
	}()

	return sessions
}

func (server *Server) launchSession(conn *ssh.ServerConn, chans <-chan ssh.NewChannel, reqs <-chan *ssh.Request) *Session {
	log.Printf("New SSH connection from %s (%s)", conn.RemoteAddr(), conn.ClientVersion())

	views := make(chan View)
	forwards := make(chan forward.Forwarder)
	session := NewSession(views, forwards)

	var closer sync.Once

	go func() {
		for req := range reqs {
			if req.Type == "tcpip-forward" {
				forward, err := server.handleTCPForward(conn, req)
				if err != nil {
					session.Broadcast([]byte(fmt.Sprintf("Could not establish forwarding: %s", err)))
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
				view, err := server.handleSessionChannel(conn, newChannel)
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

func (server *Server) handleSessionChannel(conn *ssh.ServerConn, newChannel ssh.NewChannel) (View, error) {
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

func (server *Server) handleTCPForward(conn *ssh.ServerConn, req *ssh.Request) (forward.Forwarder, error) {
	fr, err := forward.ParseForwardRequest(req.Payload)
	if err != nil {
		if req.WantReply {
			req.Reply(false, nil)
		}
		return nil, err
	}

	var fwd forward.Forwarder
	switch fr.Port {
	case 80:
		user, parts, ok := server.Config.Users.LookupUser(conn.User())
		if !ok {
			panic("Internal error: user should exist")
		}

		var hostname string
		if len(parts) == 0 {
			hostname = fmt.Sprintf("%s.%s", user.Username, server.Hostname)
		} else {
			hostname = fmt.Sprintf("%s.%s.%s", strings.Join(parts, "."), user.Username, server.Hostname)
		}
		fwd = forward.NewHTTPForwarder(hostname, conn, fr)
	case 0:
		fwd = forward.NewRawForwarder(conn, fr)
	default:
		bs := make([]byte, 0)
		helpers.PackInt(&bs, fr.Port)
		req.Reply(false, bs)

		return nil, fmt.Errorf("Forward request invalid port")
	}

	bs := make([]byte, 0)
	helpers.PackInt(&bs, fr.Port)
	req.Reply(true, bs)

	return fwd, nil
}

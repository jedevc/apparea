package tunnel

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/jedevc/apparea/config"
	"github.com/jedevc/apparea/forward"
	"github.com/jedevc/apparea/helpers"
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

	log.Printf("Listening for SSH connections on %s...", address)
	go func() {
		for {
			tcpConn, err := listener.Accept()
			if err != nil {
				continue
			}

			sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, server.Config.SSHConfig)
			if err != nil {
				continue
			}

			session := server.launchSession(sshConn, chans, reqs)

			sessions <- session
		}
	}()

	return sessions
}

func (server *Server) launchSession(conn *ssh.ServerConn, chans <-chan ssh.NewChannel, reqs <-chan *ssh.Request) *Session {
	log.Printf("Incoming session from %s (%s)", conn.User(), conn.RemoteAddr())

	views := make(chan View)
	forwards := make(chan forward.Forwarder)
	session := NewSession(views, forwards)

	var closer sync.Once
	closeChans := func() {
		close(forwards)
		close(views)
		log.Printf("Closing session from %s (%s)", conn.User(), conn.RemoteAddr())
	}

	go func() {
		for req := range reqs {
			if req.Type == "tcpip-forward" {
				forward, err := server.handleTCPForward(conn, req)
				if err != nil {
					fmt.Fprintf(session, "Could not establish forwarding: %s\n", err)
					continue
				}
				forward.AttachClientLog(session)
				forwards <- forward
			} else {
				// discard request
				if req.WantReply {
					req.Reply(false, nil)
				}
			}
		}

		closer.Do(closeChans)
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

		closer.Do(closeChans)
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

	user, parts, ok := server.Config.Users.LookupUser(conn.User())
	if !ok {
		panic("Internal error: user should exist")
	}

	hostname := server.generateHost(user, parts)

	var fwd forward.Forwarder
	switch fr.Port {
	case 80:
		fwd = forward.NewHTTPForwarder(hostname, conn, fr)

		err := fwd.Serve()
		if err != nil {
			req.Reply(false, nil)
			return nil, err
		}

		log.Printf("Forwarding http from %s (%s)", conn.User(), conn.RemoteAddr())
		req.Reply(true, nil)
	case 443:
		fwd = forward.NewHTTPForwarder(hostname, conn, fr).UseTLS(true)

		err := fwd.Serve()
		if err != nil {
			req.Reply(false, nil)
			return nil, err
		}

		log.Printf("Forwarding https from %s (%s)", conn.User(), conn.RemoteAddr())
		req.Reply(true, nil)
	case 0:
		fwd = forward.NewRawForwarder(hostname, conn, fr)

		err := fwd.Serve()
		if err != nil {
			req.Reply(false, nil)
			return nil, err
		}

		log.Printf("Forwarding tcp from %s (%s) to :%d", conn.User(), conn.RemoteAddr(), fwd.ListenerPort())

		bs := make([]byte, 0)
		helpers.PackInt(&bs, fwd.ListenerPort())
		req.Reply(true, bs)
	default:
		req.Reply(false, nil)
		return nil, fmt.Errorf("Forward request invalid port")
	}

	return fwd, nil
}

func (server Server) generateHost(user config.User, parts []string) string {
	if len(parts) == 0 {
		return fmt.Sprintf("%s.%s", user.Username, server.Hostname)
	} else {
		return fmt.Sprintf("%s-%s.%s", strings.Join(parts, "-"), user.Username, server.Hostname)
	}
}

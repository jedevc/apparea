package tunnel

import (
	"io"
	"net"
	"sync"

	"log"
)

type Session struct {
	views []View
	lock  *sync.Mutex
}

func NewSession(views chan View, forwards chan Forwarder) *Session {
	session := Session{
		views: []View{},
		lock:  new(sync.Mutex),
	}

	go func() {
		for {
			view, ok := <-views
			if !ok {
				break
			}
			go session.handleView(view)
		}
	}()
	go func() {
		for {
			forward, ok := <-forwards
			if !ok {
				break
			}
			go session.handleForwarder(forward)
		}
	}()

	return &session
}

func (session *Session) Broadcast(msg []byte) (err error) {
	session.lock.Lock()
	for _, view := range session.views {
		_, err = view.Write(msg)
		if err != nil {
			return
		}
	}
	session.lock.Unlock()

	return
}

func (session *Session) handleView(view View) {
	session.lock.Lock()
	session.views = append(session.views, view)
	session.lock.Unlock()
}

func (session *Session) handleForwarder(forward Forwarder) {
	ln, err := net.Listen("tcp", forward.Request.Address())
	if err != nil {
		log.Print("Could not listen on :8080")
		return
	}
	for {
		cl, err := ln.Accept()
		if err != nil {
			log.Print("Could not accept connection")
			return
		}

		tunn, err := forward.Connect()
		if err != nil {
			log.Print("Could not open remote connection")
			cl.Close()
			continue
		}

		closer := func() {
			cl.Close()
			tunn.Close()
		}

		var once sync.Once
		go func() {
			io.Copy(tunn, cl)
			once.Do(closer)
		}()
		go func() {
			io.Copy(cl, tunn)
			once.Do(closer)
		}()
	}

	// FIXME: this needs closing
	ln.Close()
}

package tunnel

import (
	"sync"

	"log"

	"github.com/jedevc/AppArea/forward"
)

type Session struct {
	views    []View
	forwards []forward.Forwarder

	messages [][]byte

	lock *sync.Mutex
}

func NewSession(views chan View, forwards chan forward.Forwarder) *Session {
	session := Session{
		views:    []View{},
		lock:     new(sync.Mutex),
		messages: make([][]byte, 0),
	}

	var once sync.Once

	go func() {
		for {
			view, ok := <-views
			if !ok {
				break
			}
			go session.handleView(view)
		}

		once.Do(session.Close)
	}()
	go func() {
		for {
			forward, ok := <-forwards
			if !ok {
				break
			}
			go session.handleForwarder(forward)
		}

		once.Do(session.Close)
	}()

	return &session
}

func (session *Session) Broadcast(msg []byte) (err error) {
	session.messages = append(session.messages, msg)

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

func (session *Session) Close() {
	session.lock.Lock()
	for _, forward := range session.forwards {
		forward.Close()
	}
	session.forwards = nil
	session.views = nil
	session.lock.Unlock()
}

func (session *Session) handleView(view View) {
	session.lock.Lock()
	session.views = append(session.views, view)
	for _, message := range session.messages {
		view.Write(message)
	}
	session.lock.Unlock()
}

func (session *Session) handleForwarder(forward forward.Forwarder) {
	err := forward.ListenAndServe()
	if err != nil {
		log.Print(err)
		return
	}

	session.lock.Lock()
	session.forwards = append(session.forwards, forward)
	session.lock.Unlock()

	session.Broadcast([]byte("Listening on " + forward.ListenerAddress()))
}

package tunnel

import (
	"sync"

	"log"
)

type Session struct {
	views    []View
	forwards []Forwarder
	lock     *sync.Mutex
}

func NewSession(views chan View, forwards chan Forwarder) *Session {
	session := Session{
		views: []View{},
		lock:  new(sync.Mutex),
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
	session.lock.Unlock()
}

func (session *Session) handleForwarder(forward Forwarder) {
	session.lock.Lock()
	session.forwards = append(session.forwards, forward)
	session.lock.Unlock()

	err := forward.ListenAndServe()
	if err != nil {
		log.Print(err)
		return
	}
}

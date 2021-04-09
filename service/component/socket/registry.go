package socket

import (
	"sync"
	"time"
)

const (
	// Time allowed to write a broadcastMessage to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong broadcastMessage from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum broadcastMessage size allowed from peer.
	maxMessageSize = 512
)

type DTOMessage struct {
	Type string
	Data interface{}
}

type Registry struct {
	Register   chan *Socket
	Unregister chan *Socket
	Sockets    *sync.Map
}

func BuildAndRunSocketRegistry(registerHandler, unregisterHandler func(s *Socket)) *Registry {
	registry := &Registry{
		Register:   make(chan *Socket),
		Unregister: make(chan *Socket),
		Sockets:    &sync.Map{},
	}

	go registry.run(registerHandler, unregisterHandler)

	return registry
}

func (registry *Registry) run(registerHandler, unregisterHandler func(s *Socket)) {
	for {
		select {
		case s := <-registry.Register: //new connection
			registry.Sockets.Store(s.ID, s)
			registerHandler(s)
		case s := <-registry.Unregister:
			registry.Sockets.Delete(s.ID)
			unregisterHandler(s)
			break
		}
	}
}
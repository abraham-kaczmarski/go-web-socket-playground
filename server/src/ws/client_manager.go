package ws

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrEventNotSupported = errors.New("this event type is not supported")
)

type ClientManager struct {
	sync.RWMutex

	clients  clientList
	handlers map[string]EventHandler
}

func NewClientManager() *ClientManager {
	m := &ClientManager{
		clients: make(clientList),
	}

	m.setUpEventHandlers()

	return m
}

func (m *ClientManager) AddClient(c *Client) {
	m.Lock()
	defer m.Unlock()

	m.clients[c] = true
}

func (m *ClientManager) setUpEventHandlers() {
	m.handlers[EventSendMessage] = func(e Event, c *Client) error {
		fmt.Println(e)
		return nil
	}
}

func (m *ClientManager) routeEvent(event Event, c *Client) error {
	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			return err
		}

		return nil
	}

	return ErrEventNotSupported
}

func (m *ClientManager) removeClient(c *Client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.clients[c]; ok {
		c.connection.Close()
		delete(m.clients, c)
	}
}

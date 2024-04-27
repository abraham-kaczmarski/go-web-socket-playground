package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var (
	pongWait     = 10 * time.Second
	pingInterval = pongWait * 9 / 10
)

type clientList map[*Client]bool

type Client struct {
	connection *websocket.Conn
	manager    *ClientManager
	egress     chan []byte
}

func NewClient(conn *websocket.Conn, mgr *ClientManager) *Client {
	return &Client{
		connection: conn,
		manager:    mgr,
		egress:     make(chan []byte),
	}
}

func (c *Client) HandleConnection() {
	c.readMessages()
}

func (c *Client) handlePong(message string) error {
	return c.connection.SetReadDeadline((time.Now().Add(pongWait)))
}

func (c *Client) readMessages() {
	defer func() {
		c.manager.removeClient(c)
	}()

	c.connection.SetReadLimit(512)

	if err := c.connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println(err)
		return
	}

	c.connection.SetPongHandler(c.handlePong)

	for {
		_, payload, err := c.connection.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break
		}

		var request Event

		if err := json.Unmarshal(payload, &request); err != nil {
			log.Printf("error marshalling message: %v", err)
			continue
		}

		if err := c.manager.routeEvent(request, c); err != nil {
			log.Printf("error handling message %v", err)
		}
	}
}

func (c *Client) writeMessages() {
	ticker := time.NewTicker(pingInterval)

	defer func() {
		ticker.Stop()
		c.manager.removeClient(c)
	}()

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("connection closed: ", err)
				}
				return
			}

			data, err := json.Marshal(message)

			if err != nil {
				log.Println(err)
				continue
			}

			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println(err)
			}
		case <-ticker.C:
			if err := c.connection.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Println("writemsg: ", err)
				return
			}
		}
	}
}

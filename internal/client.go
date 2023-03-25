package internal

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// readPump pumps messages from the websocket connection to the hub.
func (s subscription) readPump() {
	c := s.conn
	defer func() {
		H.unregister <- s
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		var readPayload MessagePayload

		err := c.ws.ReadJSON(&readPayload)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				fmt.Printf("error: %v", err.Error())
			}
			break
		}
		m := message{readPayload, s.room}
		H.broadcast <- m
	}
}

// write writes a message with the given message type and payload.
func (c *connection) write(payload MessagePayload) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteJSON(payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (s *subscription) writePump() {
	c := s.conn
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(MessagePayload{})
				return
			}
			if err := c.write(message); err != nil {
				log.Fatalf("error: %v", err)
				return
			}
		case <-ticker.C:
			if err := c.write(MessagePayload{}); err != nil {
				log.Fatalf("error: %v", err)
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServeWs(w http.ResponseWriter, r *http.Request, roomId string) {
	fmt.Print(roomId)
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	c := &connection{send: make(chan MessagePayload), ws: ws}
	s := subscription{c, roomId}
	H.register <- s
	go s.writePump()
	go s.readPump()
}

package internal

import "github.com/gorilla/websocket"

type (
	// WebSocket Payload
	MessagePayload struct {
		User    string `json:"user"`
		Message string `json:"message"`
	}

	// connection is an middleman between the websocket connection and the hub.
	connection struct {
		// The websocket connection.
		ws *websocket.Conn

		// Buffered channel of outbound messages.
		send chan []byte
	}

	//hub maintains the set of active connections and broadcasts messages to the connections.
	hub struct {
		// Registered connections.
		rooms map[string]map[*connection]bool

		// Inbound messages from the connections.
		broadcast chan message

		// Register requests from the connections.
		register chan subscription

		// Unregister requests from connections.
		unregister chan subscription
	}

	subscription struct {
		conn *connection
		room string
	}

	message struct {
		data []byte
		room string
	}
)

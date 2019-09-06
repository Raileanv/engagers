package main

import "encoding/json"

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case newClient := <-h.register:
			logger.Infoln( "[WEBSOCKET] Register new user")
			h.clients[newClient] = true
		case client := <-h.unregister:
			logger.Infoln( "[WEBSOCKET] Unregister user")
			delete(h.clients, client)
			close(client.send)
		case message := <-h.broadcast:
			for client := range h.clients {
				if client.sessionId == message.Client.sessionId {
					message.SelfSended = false
					if client == message.Client {
						message.SelfSended = true
					}
					logger.Infoln( "[WEBSOCKET]", message)
					msg, _ := json.Marshal(message)
					select {
					case client.send <- msg:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
		}
	}
}

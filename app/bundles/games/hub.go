// Package games handles everything related to our game
package games

import (
	"fmt"

	"github.com/gorilla/websocket"
)

// ClientManager Manages all connections to the game
type ClientManager struct {
	clients      map[*Client]bool
	broadcast    chan *Message
	addClient    chan *Client
	removeClient chan *Client
}

// Client is every person connecting to the game
type Client struct {
	Conn *websocket.Conn
	send chan *Message
	Type string
}

// Start the manager
func (manager *ClientManager) Start(g *Game) {
	for {
		select {
		case conn := <-manager.addClient:
			manager.clients[conn] = true
			// p := NewPlayer("player", conn)
		case conn := <-manager.removeClient:
			if _, ok := manager.clients[conn]; ok {
				delete(manager.clients, conn)
				close(conn.send)
			}
		case message := <-manager.broadcast:
			for conn := range manager.clients {
				select {
				case conn.send <- message:
				default:
					close(conn.send)
					delete(manager.clients, conn)
				}
			}
		}
	}
}

// WriteJSON to the client
func (c *Client) WriteJSON() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.WriteJSON(message)
		}
	}
}

func (c *Client) read() {
	defer func() {
		MainGame.ClientManager.removeClient <- c
		c.Conn.Close()
	}()

	for {
		m := &Message{}
		err := c.Conn.ReadJSON(m)
		if err != nil {
			MainGame.ClientManager.removeClient <- c
			c.Conn.Close()
			break
		}
		fmt.Println(m)
		// msg := <-MainGame.ClientManager.broadcast
		// fmt.Println(msg)
	}
}

package games

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
	"github.com/krishamoud/game/app/common/db"
)

const channelBufSize = 100

type Client struct {
	id     string
	ws     *websocket.Conn
	ch     chan *[]byte
	doneCh chan bool
	server *Game
	player *Player
}

// NewClient initializes a new Client struct with given websocket.
func NewClient(ws *websocket.Conn, server *Game) *Client {
	if ws == nil {
		panic("ws cannot be nil")
	}

	ch := make(chan *[]byte, channelBufSize)
	doneCh := make(chan bool)

	return &Client{
		id:     db.RandomID(12),
		ws:     ws,
		ch:     ch,
		doneCh: doneCh,
		server: server,
	}
}

// Conn returns client's websocket.Conn struct.
func (c *Client) Conn() *websocket.Conn {
	return c.ws
}

// SendMessage sends game state to the client.
func (c *Client) SendMessage(bytes *[]byte) {
	select {
	case c.ch <- bytes:
	default:
		// c.server.monitor.AddDroppedMessage()
	}
}

// Done sends done message to the Client which closes the conection.
func (c *Client) Done() {
	c.doneCh <- true
}

// Listen Write and Read request via chanel
func (c *Client) Listen() {
	go c.listenWrite()
	c.listenRead()
}

// Listen write request via chanel
func (c *Client) listenWrite() {
	defer func() {
		err := c.ws.Close()
		if err != nil {
			log.Println("Error:", err.Error())
		}
	}()

	log.Println("Listening write to client")
	for {
		select {

		case bytes := <-c.ch:
			// before := time.Now()
			err := c.ws.WriteMessage(websocket.TextMessage, *bytes)
			// after := time.Now()

			if err != nil {
				log.Println(err)
			} else {
				// elapsed := after.Sub(before)
				// c.server.monitor.AddSendTime(elapsed)
			}

		case <-c.doneCh:
			c.doneCh <- true
			return
		}
	}
}

func (c *Client) listenRead() {
	defer func() {
		err := c.ws.Close()
		if err != nil {
			log.Println("Error:", err.Error())
		}
	}()

	log.Println("Listening read from client")
	for {
		select {

		case <-c.doneCh:
			c.doneCh <- true
			return

		default:
			c.readFromWebSocket()
		}
	}
}

func (c *Client) readFromWebSocket() {
	_, data, err := c.ws.ReadMessage()
	if err != nil {
		log.Println(err)

		c.doneCh <- true
		// c.server.eventsDispatcher.FireUserLeft(&events.UserLeft{ClientID: c.id})
		// } else if messageType != websocket.BinaryMessage {
		// 	log.Println("Non binary message recived, ignoring")
	} else {
		m := &Message{}
		json.Unmarshal(data, &m)
		c.server.dispatch(m, c.player)
		// c.unmarshalUserInput(data)
	}
}

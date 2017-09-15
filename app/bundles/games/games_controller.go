// Package games handles everything related to our game
package games

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/krishamoud/game/app/common/controller"
)

// Controller struct
type Controller struct {
	common.Controller
}

// Connection maintains the socket and type
type Connection struct {
	Conn *websocket.Conn
	Type string
}

// Websocket upgrade to push logs
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Connect starts the user connection to the game
func (c *Controller) Connect(w http.ResponseWriter, r *http.Request) {
	// upgrade the connection for websockets
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}
	cn := &Connection{
		Conn: conn,
		Type: r.FormValue("type"),
	}
	setupConnection(cn)
}

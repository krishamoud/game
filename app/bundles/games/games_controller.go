// Package games handles everything related to our game
package games

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gorilla/websocket"
	"github.com/krishamoud/game/app/common/controller"
)

// Controller struct
type Controller struct {
	common.Controller
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
	// Upgrade connection
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return
	}
	if err, id := MainGame.Hub.EventsCollector.Add(conn); err != nil {
		log.Printf("Failed to add connection %v", err)
		conn.Close()
	} else {
		fmt.Println(id)
		p := MainGame.Hub.AddPlayer(conn)
		MainGame.Players[id] = p

	}
}

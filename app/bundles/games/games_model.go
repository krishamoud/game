// Package games handles everything related to our game
package games

import (
	"container/list"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/krishamoud/game/app/common/conf"
	"github.com/krishamoud/game/app/common/quadtree"
	"github.com/krishamoud/game/app/common/utils"
)

var rawEmptyObj = json.RawMessage(`{}`)

const (
	player = "player"
)

var EventsCollector *Hub
var err error

// MainGame is the single exported game the server runs until I create a GameManager
var MainGame = &Game{
	Players:    make(map[int]*Player),
	Users:      list.New(),
	Food:       list.New(),
	Ballistics: list.New(),
	mu:         new(sync.Mutex),
	Hub:        EventsCollector,
	ClientManager: &ClientManager{
		clients:      make(map[*Client]bool),
		broadcast:    make(chan *Message),
		addClient:    make(chan *Client),
		removeClient: make(chan *Client),
	},
	Sockets: make(map[string]*Client),
	Quadtree: &quadtree.Quadtree{
		Bounds: quadtree.Bounds{
			X:      0,
			Y:      0,
			Width:  c.GameWidth,
			Height: c.GameHeight,
		},
		MaxObjects: 200,
		MaxLevels:  7,
		Level:      0,
		Objects:    make([]quadtree.Bounds, 0),
		Nodes:      make([]quadtree.Quadtree, 0),
	},
}

// Message is a websocket message
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

var c = conf.AppConf
var initMassLog = utils.Log(float64(c.DefaultPlayerMass), float64(c.SlowBase))

func (g *Game) dispatch(msg *Message, p *Player) {
	switch msg.Type {
	case "gotit":
		data := make(map[string]interface{})
		_ = json.Unmarshal(msg.Data, &data)
		p.Name = data["name"].(string)
		p.ScreenHeight = data["screenHeight"].(float64)
		p.ScreenWidth = data["screenWidth"].(float64)
		fmt.Println("got it")
		g.gotIt(p)
		break
	case "pingcheck":
		p.Emit("pongcheck", rawEmptyObj)
		break
	case "windowResized":
		data := make(map[string]interface{})
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			panic(err)
		}
		w := float64(data["screenWidth"].(float64))
		h := float64(data["screenHeight"].(float64))
		p.WindowResize(w, h)
		break
	case "respawn":
		g.SpliceUser(p.ID)
		p.Emit("welcome", rawEmptyObj)
		fmt.Println("[INFO] User " + p.Name + " respawned!")
		break
	case "disconnect":
		g.SpliceUser(p.ID)
		g.RemovePlayerConnection(p)
		fmt.Println("[INFO] User " + p.Name + " disconnected!")
		g.Broadcast(p.ID, "playerDisconnect", rawEmptyObj)
		break
	case "0":
		p.LastHeartbeat = time.Now()
		point := make(map[string]interface{})
		err := json.Unmarshal(msg.Data, &point)
		if err != nil {
			panic(err)
		}
		if point["x"].(float64) != p.Point.X || point["y"].(float64) != p.Point.Y {
			p.Target = &utils.Point{
				X: point["x"].(float64),
				Y: point["y"].(float64),
			}
		}
		break
	case "2":
		p.Fire(g)
	}
}

func (g *Game) gotIt(p *Player) {
	if !utils.ValidNickname(p.Name) {
		p.Conn.Conn.WriteMessage(websocket.TextMessage, []byte("kick"))
		g.RemovePlayerConnection(p)
	} else {
		fmt.Println("[INFO] Player " + p.Name + " connected!")
		g.AddPlayerConnection(p)

		radius := utils.MassToRadius(c.DefaultPlayerMass)
		position := utils.RandomPosition(radius)

		p.Point.X = position.X
		p.Point.Y = position.Y
		p.Target.X = 0
		p.Target.Y = 0
		if p.Type == "player" {
			cells := []*Cell{
				&Cell{
					Mass:   c.DefaultPlayerMass,
					Point:  &utils.Point{X: position.X, Y: position.Y},
					Radius: radius,
				},
			}
			p.Cells = cells
			p.MassTotal = c.DefaultPlayerMass
		}
		p.Hue = rand.Intn(360)
		p.LastHeartbeat = time.Now()
		p.Scale = 1
		p.ClipSize = 10
		p.ShotsLeft = p.ClipSize
		p.MassCurrent = p.MassTotal
		g.PushUser(p)
		var n = struct {
			Name string `json:"name"`
		}{
			p.Name,
		}
		b, _ := json.MarshalIndent(&n, "", "\t")
		g.Emit("playerJoin", b)
		var gd = struct {
			GameWidth  float64 `json:"gameWidth"`
			GameHeight float64 `json:"gameHeight"`
		}{
			c.GameWidth,
			c.GameHeight,
		}
		data, _ := json.MarshalIndent(&gd, "", "\t")
		p.Emit("gameSetup", data)
		fmt.Println("Total Players:", g.Users.Len())
	}
}

func (g *Game) removeFood(toRem int) {
	for toRem > 0 {
		g.PopFood()
		toRem--
	}
}

func (g *Game) userMass() float64 {
	var total float64
	for e := g.Users.Front(); e != nil; e = e.Next() {
		u := e.Value.(*Player)
		total += u.MassTotal
	}
	return total
}

// Package games handles everything related to our game
package games

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/krishamoud/game/app/common/conf"
	"github.com/krishamoud/game/app/common/db"
	"github.com/krishamoud/game/app/common/quadtree"
	"github.com/krishamoud/game/app/common/utils"
)

var rawEmptyObj = json.RawMessage(`{}`)

const (
	player = "player"
)

// MainGame is the single exported game the server runs until I create a GameManager
var MainGame = &Game{
	Users:   []*Player{},
	FoodArr: []*Food{},
	Sockets: make(map[string]*Connection),
	mu:      new(sync.Mutex),
	Quadtree: &quadtree.Quadtree{
		Bounds: quadtree.Bounds{
			X:      0,
			Y:      0,
			Width:  c.GameWidth,
			Height: c.GameHeight,
		},
		MaxObjects: 10,
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

func setupConnection(cn *Connection) {
	radius := utils.MassToRadius(c.DefaultPlayerMass)
	position := utils.RandomPosition(radius)
	cells := []*Cell{}
	var shape string
	var massTotal float64
	if cn.Type == player {
		if len(MainGame.Users)&1 == 1 {
			shape = circle
		} else {
			shape = square
		}
		cell := &Cell{
			Mass: c.DefaultPlayerMass,
			Point: &utils.Point{
				X: position.X,
				Y: position.Y,
			},
			Radius: radius,
		}
		cells = append(cells, cell)
		massTotal = c.DefaultPlayerMass
	}
	currentPlayer := &Player{
		ID:            db.RandomID(12),
		Point:         position,
		W:             c.DefaultPlayerMass,
		H:             c.DefaultPlayerMass,
		Cells:         cells,
		MassTotal:     massTotal,
		Hue:           rand.Intn(360),
		Type:          cn.Type,
		LastHeartbeat: time.Now(),
		Target:        &utils.Point{X: 0, Y: 0},
		Conn:          cn,
		mu:            new(sync.Mutex),
		Shape:         shape,
	}

	for {
		m := &Message{}
		err := cn.Conn.ReadJSON(m)
		if err != nil {
			return
		}
		dispatch(m, currentPlayer)
	}
}

func dispatch(msg *Message, p *Player) {
	switch msg.Type {
	case "gotit":
		data := make(map[string]interface{})
		_ = json.Unmarshal(msg.Data, &data)
		p.Name = data["name"].(string)
		p.ScreenHeight = data["screenHeight"].(float64)
		p.ScreenWidth = data["screenWidth"].(float64)
		gotIt(p, data)
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
		if i := MainGame.FindDuplicatePlayer(p.ID); i != -1 {
			MainGame.SpliceUser(i)
		}
		p.Emit("welcome", rawEmptyObj)
		fmt.Println("[INFO] User " + p.Name + " respawned!")
		break
	case "disconnect":
		if i := MainGame.FindDuplicatePlayer(p.ID); i > -1 {
			MainGame.SpliceUser(i)
			MainGame.RemovePlayerConnection(p)
		}
		fmt.Println("[INFO] User " + p.Name + " disconnected!")
		MainGame.Broadcast(p.ID, "playerDisconnect", rawEmptyObj)
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
	case "1":

	}
}

func gotIt(p *Player, msg map[string]interface{}) {
	fmt.Println(msg)
	if MainGame.FindDuplicatePlayer(p.ID) != -1 {
		fmt.Println("[INFO] Player ID is already connected, kicking")
		MainGame.RemovePlayerConnection(p)
	} else if !utils.ValidNickname(p.Name) {
		p.Conn.Conn.WriteMessage(websocket.TextMessage, []byte("kick"))
		MainGame.RemovePlayerConnection(p)
	} else {
		fmt.Println("[INFO] Player " + p.Name + " connected!")
		MainGame.AddPlayerConnection(p)

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
		MainGame.PushUser(p)
		var n = struct {
			Name string `json:"name"`
		}{
			p.Name,
		}
		b, _ := json.MarshalIndent(&n, "", "\t")
		MainGame.Emit("playerJoin", b)
		var gd = struct {
			GameWidth  float64 `json:"gameWidth"`
			GameHeight float64 `json:"gameHeight"`
		}{
			c.GameWidth,
			c.GameHeight,
		}
		data, _ := json.MarshalIndent(&gd, "", "\t")
		p.Emit("gameSetup", data)
		fmt.Println("Total Players:", len(MainGame.Users))
	}
}

func removeFood(toRem int) {
	for toRem > 0 {
		MainGame.PopFood()
		toRem--
	}
}

func moveMass(mass *Mass) {
	deg := math.Atan2(float64(mass.Target.Y), float64(mass.Target.X))
	deltaY := mass.Speed * math.Sin(deg)
	deltaX := mass.Speed * math.Cos(deg)
	mass.Speed -= 0.5
	if mass.Speed < 0 {
		mass.Speed = 0
	}
	mass.Point.Y += deltaY
	mass.Point.X += deltaX

	borderCalc := mass.Radius + 5
	if mass.Point.X > c.GameWidth-borderCalc {
		mass.Point.X = c.GameWidth - borderCalc
	}
	if mass.Point.Y > c.GameHeight-borderCalc {
		mass.Point.Y = c.GameHeight - borderCalc
	}

	if mass.Point.X < borderCalc {
		mass.Point.X = borderCalc
	}
	if mass.Point.Y < borderCalc {
		mass.Point.Y = borderCalc
	}
}

func userMass() float64 {
	var total float64
	for _, u := range MainGame.Users {
		total += u.MassTotal
	}
	return total
}

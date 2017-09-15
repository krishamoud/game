// Package games handles everything related to our game
package games

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/Tarliton/collision2d"
	"github.com/krishamoud/game/app/common/db"
	"github.com/krishamoud/game/app/common/quadtree"
	"github.com/krishamoud/game/app/common/utils"
)

// Game holds the state for the entire game
type Game struct {
	Users    []*Player
	FoodArr  []*Food
	Sockets  map[string]*Connection
	Quadtree *quadtree.Quadtree
	mu       *sync.Mutex
}

// PushFood adds food to a one of the food arrays
func (g *Game) PushFood(f *Food) {
	g.FoodArr = append(g.FoodArr, f)
}

// PopFood removes one food
func (g *Game) PopFood() {
	g.FoodArr = g.FoodArr[1:]
}

// SpliceFood removes food at an index
func (g *Game) SpliceFood(i int) {
	if i < len(g.FoodArr) {
		g.FoodArr = append(g.FoodArr[:i], g.FoodArr[i+1:]...)
	}
}

// AddPlayerConnection adds the players socket to the game
func (g *Game) AddPlayerConnection(p *Player) {
	g.Sockets[p.ID] = p.Conn
}

// RemovePlayerConnection removes the players socke to the game
func (g *Game) RemovePlayerConnection(p *Player) {
	g.Sockets[p.ID].Conn.Close()
	delete(g.Sockets, p.ID)
}

// MoveLoop ticks every player
func (g *Game) MoveLoop() {
	for _, p := range g.Users {
		g.tickPlayer(p)
	}
}

// GameLoop balances the mass
func (g *Game) GameLoop() {
	g.balanceMass()
}

// SendUpdates updates all clients to the current game state
func (g *Game) SendUpdates() {
	for _, p := range g.Users {
		if p.Point.X == 0 {
			p.Point.X = c.GameWidth / 2
		}
		if p.Point.Y == 0 {
			p.Point.Y = c.GameHeight / 2
		}
		visibleFood := p.VisibleFood()
		visiblePlayers := p.VisibleCells()
		var m = struct {
			Players     []*Player `json:"players"`
			VisibleFood []*Food   `json:"visibleFood"`
		}{
			visiblePlayers,
			visibleFood,
		}
		data, _ := json.MarshalIndent(&m, "", "\t")
		p.Emit("serverTellPlayerMove", data)
		// visibleM
	}
}

// MoveInterval moves the game alone
func (g *Game) MoveInterval() {
	ticker := time.NewTicker(1000 / 60 * time.Millisecond)
	quit := make(chan struct{})
	func() {
		for {
			select {
			case <-ticker.C:
				// do stuff
				g.MoveLoop()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

// GameInterval runs GameLoop at 60hz
func (g *Game) GameInterval() {
	ticker := time.NewTicker(1000 * time.Millisecond)
	quit := make(chan struct{})
	func() {
		for {
			select {
			case <-ticker.C:
				// do stuff
				g.GameLoop()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

// UpdateInterval runs SendUpdates in an interval dependent on network latency
func (g *Game) UpdateInterval() {
	n := time.Duration(c.NetworkUpdateFactor)
	ticker := time.NewTicker(1000 / n * time.Millisecond)
	quit := make(chan struct{})
	func() {
		for {
			select {
			case <-ticker.C:
				// do stuff
				g.SendUpdates()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (g *Game) tickPlayer(p *Player) {
	// hbDuration := time.Duration(c.MaxHeartBeatInterval) * time.Millisecond
	if false {
		// if p.LastHeartbeat.Before(time.Now().Add(-hbDuration)) {
		fmt.Println("Kicking for inactivity")
		str := "Last heartbeat recieved over " + string(c.MaxHeartBeatInterval) + " ms ago"
		var m = struct {
			Msg string `jsong:"msg"`
		}{
			str,
		}
		body, _ := json.MarshalIndent(&m, "", "\t")
		p.Emit("kick", body)
		g.RemovePlayerConnection(p)
		if i := g.FindDuplicatePlayer(p.ID); i > -1 {
			g.SpliceUser(i)
		}
	}
	p.SetCollider()
	p.movePlayer()
	p.SetSpeed()
	g.RefreshQTree()
	b := quadtree.Bounds{
		X:      p.Point.X,
		Y:      p.Point.Y,
		Width:  utils.MassToRadius(p.MassTotal),
		Height: utils.MassToRadius(p.MassTotal),
	}
	collidablePoints := g.Quadtree.Retrieve(b)
	for _, col := range collidablePoints {
		switch col.P {
		case "food":
			if len(p.Cells) != 0 && col.Idx < len(g.FoodArr) {
				fc := g.FoodArr[col.Idx].Col
				if p.Shape == circle {
					if p.CheckCircleCollision(fc) {
						p.Cells[0].Mass += c.FoodMass
						p.MassTotal += c.FoodMass
						p.Cells[0].Radius = utils.MassToRadius(p.Cells[0].Mass)
						g.SpliceFood(col.Idx)
					}
				} else {
					if p.CheckBoxCollision(fc) {
						p.Cells[0].Mass += c.FoodMass
						p.MassTotal += c.FoodMass
						p.Cells[0].Radius = utils.MassToRadius(p.Cells[0].Mass)
						g.SpliceFood(col.Idx)
					}
				}
			}
		}
	}
}

// RefreshQTree clears and reinserts the quadtree nodes
func (g *Game) RefreshQTree() {
	g.Quadtree.Clear()
	for _, u := range g.Users {
		b := quadtree.Bounds{
			X:      u.Point.X,
			Y:      u.Point.Y,
			Width:  utils.MassToRadius(u.MassTotal),
			Height: utils.MassToRadius(u.MassTotal),
			P:      "user",
		}
		g.Quadtree.Insert(b)
	}
	for i, f := range g.FoodArr {
		b := quadtree.Bounds{
			X:      f.Col.Pos.X,
			Y:      f.Col.Pos.Y,
			Width:  f.Radius,
			Height: f.Radius,
			P:      "food",
			Idx:    i,
		}
		g.Quadtree.Insert(b)
	}
}

// FindDuplicatePlayer finds the playerID and returns its index, -1 if not found
func (g *Game) FindDuplicatePlayer(playerID string) int {
	for i, p := range g.Users {
		if p.ID == playerID {
			return i
		}
	}
	return -1
}

// PushUser adds a User to the Game.Users slice
func (g *Game) PushUser(u *Player) {
	g.Users = append(g.Users, u)
}

// SpliceUser removes a User from the Game.Users slice
func (g *Game) SpliceUser(i int) {
	g.Users = append(g.Users[:i], g.Users[i+1:]...)
}

// Emit sends websocket messages to every player in the game
func (g *Game) Emit(msg string, body json.RawMessage) {
	for _, p := range g.Users {
		p.mu.Lock()
		cn := p.Conn.Conn
		message := &Message{
			Type: msg,
			Data: body,
		}
		cn.WriteJSON(message)
		p.mu.Unlock()
	}
}

// Broadcast sends websocket messages to every player except the playerID that called it
func (g *Game) Broadcast(pID string, msg string, body json.RawMessage) {
	for _, p := range g.Users {
		if pID != p.ID {
			p.mu.Lock()
			cn := p.Conn.Conn
			message := &Message{
				Type: msg,
				Data: body,
			}
			cn.WriteJSON(message)
			p.mu.Lock()
		}
	}
}
func (g *Game) addFood(toAdd int) {
	radius := utils.MassToRadius(c.FoodMass)
	for toAdd > 0 {
		position := utils.RandomPosition(radius)
		pos := collision2d.NewVector(position.X, position.Y)
		g.PushFood(&Food{
			ID:     db.RandomID(12),
			Point:  position,
			Radius: radius,
			Mass:   rand.Float64() + 2,
			Hue:    rand.Intn(360),
			Col:    collision2d.NewCircle(pos, radius),
		})
		toAdd--
	}
}
func (g *Game) balanceMass() {
	totalMass := float64(len(MainGame.FoodArr))*c.FoodMass + userMass()
	massDiff := c.GameMass - totalMass
	maxFoodDiff := c.MaxFood - float64(len(MainGame.FoodArr))
	foodDiff := massDiff/c.FoodMass - maxFoodDiff
	foodToAdd := math.Min(float64(foodDiff), float64(maxFoodDiff))
	foodToRemove := -math.Max(float64(foodDiff), float64(maxFoodDiff))
	if foodToAdd > 0 {
		g.addFood(int(foodToAdd))
	} else if foodToRemove > 0 {
		removeFood(int(foodToRemove))
	}
}

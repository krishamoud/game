// Package games handles everything related to our game
package games

import (
	"container/list"
	"encoding/json"
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
	Users         *list.List
	Food          *list.List
	Ballistics    *list.List
	ClientManager *ClientManager
	Sockets       map[string]*Client
	Quadtree      *quadtree.Quadtree
	mu            *sync.Mutex
}

// PushFood adds food to a one of the food arrays
func (g *Game) PushFood(f *Food) {
	g.Food.PushFront(f)
}

// PopFood removes one food
func (g *Game) PopFood() {
	g.Food.Remove(g.Food.Front())
}

// SpliceFood removes food at an index
func (g *Game) SpliceFood(id string) {
	for e := g.Food.Front(); e != nil; e = e.Next() {
		// do something with e.Value
		if e.Value.(*Food).ID == id {
			g.Food.Remove(e)
		}
	}
}

// AddPlayerConnection adds the players socket to the game
func (g *Game) AddPlayerConnection(p *Player) {
	g.Sockets[p.ID] = p.Conn
}

// RemovePlayerConnection removes the players socke to the game
func (g *Game) RemovePlayerConnection(p *Player) {
	// g.Sockets[p.ID].Conn.Close()
	// delete(g.Sockets, p.ID)
}

// MoveLoop ticks every player
func (g *Game) MoveLoop() {
	for e := g.Users.Front(); e != nil; e = e.Next() {
		p := e.Value.(*Player)
		g.tickPlayer(p)
	}
	for e := g.Ballistics.Front(); e != nil; e = e.Next() {
		b := e.Value.(*Ballistic)
		b.Update(g)
	}
	for e := g.Food.Front(); e != nil; e = e.Next() {
		f := e.Value.(*Food)
		f.Update(g)
	}
	g.RefreshQTree()
}

// SendUpdates updates all clients to the current game state
func (g *Game) SendUpdates() {
	for e := g.Users.Front(); e != nil; e = e.Next() {
		p := e.Value.(*Player)
		visibleFood := p.VisibleFood(g)
		visiblePlayers := p.VisibleCells(g)
		visibleBallistics := p.VisibleBallistics(g)
		var m = struct {
			Players           []*Player    `json:"players"`
			VisibleFood       []*Food      `json:"visibleFood"`
			VisibleBallistics []*Ballistic `json:"visibleBallistics"`
		}{
			visiblePlayers,
			visibleFood,
			visibleBallistics,
		}
		data, _ := json.MarshalIndent(&m, "", "\t")
		p.Emit("serverTellPlayerMove", data)
	}
}

// GameInterval runs GameLoop at 60hz
func (g *Game) GameInterval() {
	n := time.Duration(c.NetworkUpdateFactor)
	updateTicker := time.NewTicker(1000 / n * time.Millisecond)
	quit := make(chan struct{})
	func() {
		for {
			select {
			case <-updateTicker.C:
				g.MoveLoop()
				g.SendUpdates()
				g.balanceMass()
			case <-quit:
				updateTicker.Stop()
				return
			}
		}
	}()
}

func (g *Game) tickPlayer(p *Player) {
	col := p.GetCollisions(g)
	pColl := p.GetPlayerCollisions(col)
	p.checkHeartbeat(g)
	p.SetCollider()
	p.movePlayer(pColl)
	p.reload()
	p.CheckCollisions(col, g)
	p.CheckKillPlayer(g)
}

// RefreshQTree clears and reinserts the quadtree nodes
func (g *Game) RefreshQTree() {
	g.Quadtree.Clear()
	b := quadtree.Bounds{}
	for e := g.Users.Front(); e != nil; e = e.Next() {
		u := e.Value.(*Player)
		var w, h float64
		if u.Shape == circle {
			w = utils.MassToRadius(u.MassTotal)
			h = utils.MassToRadius(u.MassTotal)
			b.X = u.Point.X
			b.Y = u.Point.Y
		} else {
			w = u.W
			h = u.H
			b.X = u.Point.X
			b.Y = u.Point.Y
		}
		b.Width = w
		b.Height = h
		b.ID = u.ID
		b.Mass = u.MassTotal
		b.P = "user"
		b.Obj = u
		g.Quadtree.Insert(b)
	}
	for e := g.Food.Front(); e != nil; e = e.Next() {
		f := e.Value.(*Food)
		b := quadtree.Bounds{
			X:      f.Col.Pos.X,
			Y:      f.Col.Pos.Y,
			Width:  f.Radius,
			Height: f.Radius,
			P:      "food",
			ID:     f.ID,
			Obj:    f,
		}
		g.Quadtree.Insert(b)
	}
	for e := g.Ballistics.Front(); e != nil; e = e.Next() {
		b := e.Value.(*Ballistic)
		bnd := quadtree.Bounds{
			X:      b.Point.X,
			Y:      b.Point.Y,
			Width:  b.Radius,
			Height: b.Radius,
			P:      "ballistic",
			ID:     b.ID,
			Obj:    b,
		}
		g.Quadtree.Insert(bnd)
	}
}

// PushUser adds a User to the Game.Users list
func (g *Game) PushUser(u *Player) {
	g.Users.PushFront(u)
}

// SpliceUser removes a User from the Game.Users list
func (g *Game) SpliceUser(id string) {
	for e := g.Users.Front(); e != nil; e = e.Next() {
		p := e.Value.(*Player)
		if p.ID == id {
			g.Users.Remove(e)
			return
		}
	}
}

// RemoveBallistic removes a ballistic from the game
func (g *Game) RemoveBallistic(id string) {
	for e := g.Ballistics.Front(); e != nil; e = e.Next() {
		b := e.Value.(*Ballistic)
		if b.ID == id {
			g.Ballistics.Remove(e)
			b = nil
			return
		}
	}
}

// GetBallistic returns a ballistic index
func (g *Game) GetBallistic(id string) *Ballistic {
	var b *Ballistic
	for e := g.Ballistics.Front(); e != nil; e = e.Next() {
		temp := e.Value.(*Ballistic)
		if temp.ID == id {
			return temp
		}
	}
	return b
}

// Emit sends websocket messages to every player in the game
func (g *Game) Emit(msg string, body json.RawMessage) {
	for e := g.Users.Front(); e != nil; e = e.Next() {
		p := e.Value.(*Player)
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
	for e := g.Users.Front(); e != nil; e = e.Next() {
		p := e.Value.(*Player)
		if pID != p.ID {
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
}
func (g *Game) addFood(toAdd int) {
	radius := utils.MassToRadius(c.FoodMass)
	for toAdd > 0 {
		position := utils.RandomPosition(radius)
		pos := collision2d.NewVector(position.X, position.Y)
		id := db.RandomID(12)
		f := &Food{
			ID:     id,
			Point:  position,
			Radius: radius,
			Mass:   c.FoodMass,
			Hue:    rand.Intn(360),
			Col:    collision2d.NewCircle(pos, radius),
		}
		g.PushFood(f)
		toAdd--
	}
}
func (g *Game) balanceMass() {
	totalMass := float64(g.Food.Len())*c.FoodMass + g.userMass()
	massDiff := c.GameMass - totalMass
	maxFoodDiff := c.MaxFood - float64(g.Food.Len())
	foodDiff := massDiff/c.FoodMass - maxFoodDiff
	foodToAdd := math.Min(float64(foodDiff), float64(maxFoodDiff))
	foodToRemove := -math.Max(float64(foodDiff), float64(maxFoodDiff))
	if foodToAdd > 0 {
		g.addFood(int(foodToAdd))
	} else if foodToRemove > 0 {
		g.removeFood(int(foodToRemove))
	}
}

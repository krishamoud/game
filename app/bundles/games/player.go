// Package games handles everything related to our game
package games

import (
	"encoding/json"
	"math"
	"sync"
	"time"

	"github.com/Tarliton/collision2d"
	"github.com/krishamoud/game/app/common/utils"
)

const (
	circle = "circle"
	square = "square"
)

// Player controls an individual player state
type Player struct {
	ID                string             `json:"id"`
	Name              string             `json:"name"`
	Point             *utils.Point       `json:"point"`
	W                 float64            `json:"w"`
	H                 float64            `json:"h"`
	Cells             []*Cell            `json:"cells"`
	MassTotal         float64            `json:"massTotal"`
	Hue               int                `json:"hue"`
	Type              string             `json:"type"`
	LastHeartbeat     time.Time          `json:"lastHeartBeat"`
	Target            *utils.Point       `json:"target"`
	LastSplit         time.Time          `json:"lastSplit"`
	Conn              *Connection        `json:"conn"`
	ScreenWidth       float64            `json:"screenWidth"`
	ScreenHeight      float64            `json:"screenHeight"`
	Shape             string             `json:"shape"`
	Circle            collision2d.Circle `json:"circle"`
	Box               collision2d.Box    `json:"box"`
	triangleStartTime time.Time
	mu                *sync.Mutex
}

// SpliceCells removes cells from the player
func (p *Player) SpliceCells(i int) {
	p.Cells = append(p.Cells[:i], p.Cells[i+1:]...)
}

// PushCells adds cells to the player
func (p *Player) PushCells(cell *Cell) {
	p.Cells = append(p.Cells, cell)
}

// WindowResize resizes the window visible to the player
func (p *Player) WindowResize(w, h float64) {
	p.ScreenHeight = h
	p.ScreenWidth = w
}

// VisibleFood returns all food the player can see based on their window size
func (p *Player) VisibleFood() []*Food {
	vf := []*Food{}
	if p.Type == "player" {
		for _, f := range MainGame.FoodArr {
			if f.Point.X > p.Point.X-p.ScreenWidth &&
				f.Point.X < p.Point.X+p.ScreenWidth &&
				f.Point.Y > p.Point.Y-p.ScreenHeight &&
				f.Point.Y < p.Point.Y+p.ScreenHeight {
				vf = append(vf, f)
			}
		}
		return vf
	}
	return []*Food{}
}

// VisibleCells returns the player cells visible based on the player window size
func (p *Player) VisibleCells() []*Player {
	vc := []*Player{}
	for _, u := range MainGame.Users {
		for _, c := range u.Cells {
			if c.Point.X+c.Radius > p.Point.X-p.ScreenWidth/2-20 &&
				c.Point.X-c.Radius < p.Point.X+p.ScreenWidth/2+20 &&
				c.Point.Y+c.Radius > p.Point.Y-p.ScreenHeight/2-20 &&
				c.Point.Y-c.Radius < p.Point.Y+p.ScreenHeight/2+20 {
				if u.ID != p.ID {
					p := &Player{
						ID:        u.ID,
						Point:     u.Point,
						Cells:     u.Cells,
						MassTotal: u.MassTotal,
						Hue:       u.Hue,
						Name:      u.Name,
						Shape:     u.Shape,
						W:         u.W,
						H:         u.H,
					}
					vc = append(vc, p)
				} else {
					p := &Player{
						Point:     u.Point,
						Cells:     u.Cells,
						MassTotal: u.MassTotal,
						Hue:       u.Hue,
						Shape:     u.Shape,
						W:         u.W,
						H:         u.H,
					}
					vc = append(vc, p)
				}
			}
		}
	}
	return vc
}

// CheckCircleCollision checks if the player collided with a circle
func (p *Player) CheckCircleCollision(circle collision2d.Circle) bool {
	col, _ := collision2d.TestCircleCircle(circle, p.Circle)
	return col
}

// CheckBoxCollision checks if the player collided with a box
func (p *Player) CheckBoxCollision(circle collision2d.Circle) bool {
	col, _ := collision2d.TestCirclePolygon(circle, p.Box.ToPolygon())
	return col
}

// Emit sends a websocket message to this player
func (p *Player) Emit(msg string, body json.RawMessage) {
	p.mu.Lock()
	cn := p.Conn.Conn
	message := &Message{
		Type: msg,
		Data: body,
	}
	cn.WriteJSON(message)
	p.mu.Unlock()
}

// SetCollider sets the collider every frame to fit the expanding size
func (p *Player) SetCollider() {
	r := utils.MassToRadius(p.MassTotal)
	p.W = r * 2
	p.H = r * 2
	if p.Shape == "circle" {
		p.Circle = collision2d.NewCircle(collision2d.Vector{
			X: p.Point.X,
			Y: p.Point.Y,
		}, r)
	} else {
		p.Box = collision2d.NewBox(collision2d.Vector{
			X: p.Point.X,
			Y: p.Point.Y,
		}, p.W, p.H)
	}
}

// SetSpeed sets a default speed to 6.25 if a cell is new
func (p *Player) SetSpeed() {
	for _, cl := range p.Cells {
		if cl.Speed == 0 {
			cl.Speed = 6.25
		}
	}
}

func (p *Player) movePlayer() {
	var x, y float64
	for i, cl := range p.Cells {
		target := &utils.Point{
			X: p.Point.X - cl.Point.X + p.Target.X,
			Y: p.Point.Y - cl.Point.Y + p.Target.Y,
		}
		dist := utils.GetHypotenuse(target.X, target.Y)
		deg := math.Atan2(float64(target.Y), float64(target.X))
		slowDown := float64(1)
		if cl.Speed <= 6.25 {
			slowDown = utils.Log(float64(cl.Mass), c.SlowBase) - initMassLog + 1
		}

		deltaX := cl.Speed * math.Cos(deg) / float64(slowDown)
		deltaY := cl.Speed * math.Sin(deg) / float64(slowDown)
		if cl.Speed > 6.25 {
			cl.Speed -= 0.5
		}
		if dist < float64(50+cl.Radius) {
			deltaY *= dist / float64((50 + cl.Radius))
			deltaX *= dist / float64((50 + cl.Radius))
		}
		cl.Point.Y += deltaY
		cl.Point.X += deltaX
		for j := range p.Cells {
			if j != i && p.Cells[i] != nil {
				distance := utils.GetDistance(p.Cells[i].Point, p.Cells[j].Point)
				radiusTotal := p.Cells[i].Radius + p.Cells[j].Radius
				if distance < float64(radiusTotal) {
					mergeDuration := time.Duration(c.MergeTimer)
					if p.LastSplit.Before(time.Now().Add(-mergeDuration)) {
						if p.Cells[i].Point.X < p.Cells[j].Point.X {
							p.Cells[i].Point.X--
						} else if p.Cells[i].Point.X > p.Cells[i].Point.X {
							p.Cells[i].Point.X++
						}

						if p.Cells[i].Point.Y < p.Cells[j].Point.Y {
							p.Cells[i].Point.Y--
						} else if p.Cells[i].Point.Y > p.Cells[i].Point.Y {
							p.Cells[i].Point.Y++
						}
					} else if distance < float64(radiusTotal)/1.75 {
						p.Cells[i].Mass += p.Cells[i].Mass
						p.Cells[i].Radius = utils.MassToRadius(p.Cells[i].Mass)
						p.SpliceCells(j)
					}
				}
			}
			if len(p.Cells) > i {
				borderCalc := p.Cells[i].Radius / 3
				if p.Shape == "circle" {
					if p.Cells[i].Point.X > c.GameWidth-borderCalc {
						p.Cells[i].Point.X = c.GameWidth - borderCalc
					}
					if p.Cells[i].Point.Y > c.GameHeight-borderCalc {
						p.Cells[i].Point.Y = c.GameHeight - borderCalc
					}

					if p.Cells[i].Point.X < borderCalc {
						p.Cells[i].Point.X = borderCalc
					}
					if p.Cells[i].Point.Y < borderCalc {
						p.Cells[i].Point.Y = borderCalc
					}
				} else if p.Shape == "square" {
					if p.Cells[i].Point.X > c.GameWidth-p.W {
						p.Cells[i].Point.X = c.GameWidth - p.W
					}
					if p.Cells[i].Point.Y > c.GameHeight-p.H {
						p.Cells[i].Point.Y = c.GameHeight - p.H
					}

					if p.Cells[i].Point.X < 0 {
						p.Cells[i].Point.X = 1
					}
					if p.Cells[i].Point.Y < 0 {
						p.Cells[i].Point.Y = 1
					}
				}

				x += p.Cells[i].Point.X
				y += p.Cells[i].Point.Y
			}
		}
	}
	p.Point.X = x / float64(len(p.Cells))
	p.Point.Y = y / float64(len(p.Cells))
}

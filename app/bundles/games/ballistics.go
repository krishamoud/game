// Package games handles everything related to our game
package games

import (
	"math"

	"github.com/Tarliton/collision2d"
	"github.com/krishamoud/game/app/common/db"
	"github.com/krishamoud/game/app/common/utils"
)

// Ballistic is the base projectile class
type Ballistic struct {
	ID       string       `json:"id"`
	PlayerID string       `json:"playerId"`
	Speed    float64      `json:"speed"`
	Point    *utils.Point `json:"point"`
	Radius   float64      `json:"radius"`
	Shape    string       `json:"shape"`
	W        float64      `json:"w"`
	Hue      int          `json:"hue"`
	Degree   float64      `json:"angle"`
	Circle   collision2d.Circle
	Box      collision2d.Box
	Mass     float64
	Distance float64
}

// NewBallistic will generate a new ballistic from a player
func NewBallistic(id string, speed, mass float64, point *utils.Point, deg float64, dist float64, shape string, hue int) *Ballistic {
	radius := utils.MassToRadius(mass)
	if shape == "circle" {
		return &Ballistic{
			ID:       db.RandomID(12),
			PlayerID: id,
			Speed:    speed,
			Point:    point,
			Mass:     mass,
			Radius:   radius,
			Degree:   deg,
			Distance: dist,
			Circle:   collision2d.NewCircle(collision2d.NewVector(point.X, point.Y), radius),
			Shape:    shape,
			W:        radius * 2,
			Hue:      hue,
		}
	}
	width := utils.MassToWidth(mass)
	return &Ballistic{
		ID:       db.RandomID(12),
		PlayerID: id,
		Speed:    speed,
		Point:    point,
		Mass:     mass,
		Radius:   radius,
		Degree:   deg,
		Distance: dist,
		Box: collision2d.NewBox(collision2d.Vector{
			X: point.X,
			Y: point.Y,
		}, width, width),
		Shape: shape,
		W:     utils.MassToWidth(mass),
		Hue:   hue,
	}
}

// Update moves the ballistic
func (b *Ballistic) Update(g *Game) {
	if b.Distance < 0 {
		g.RemoveBallistic(b.ID)
		return
	}
	deltaY := b.Speed * math.Sin(b.Degree)
	deltaX := b.Speed * math.Cos(b.Degree)
	b.Point.Y += deltaY
	b.Point.X += deltaX
	b.Circle.Pos.X += deltaX
	b.Circle.Pos.Y += deltaY
	b.Box.Pos.X += deltaX
	b.Box.Pos.Y += deltaY
	b.Distance -= utils.GetHypotenuse(deltaX, deltaY)
}

// IsCircle determines if the ballistic is a circle or not
func (b *Ballistic) IsCircle() bool {
	return b.Shape == "circle"
}

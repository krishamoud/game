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
	Mass     float64
	Degree   float64
	Distance float64
	circle   collision2d.Circle
}

// NewBallistic will generate a new ballistic from a player
func NewBallistic(id string, speed, mass float64, point *utils.Point, deg float64, dist float64) *Ballistic {
	radius := utils.MassToRadius(mass) * 0.5
	return &Ballistic{
		ID:       db.RandomID(12),
		PlayerID: id,
		Speed:    speed,
		Point:    point,
		Mass:     mass,
		Radius:   radius,
		Degree:   deg,
		Distance: dist,
		circle:   collision2d.NewCircle(collision2d.NewVector(point.X, point.Y), radius),
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
	b.circle.Pos.X += deltaX
	b.circle.Pos.Y += deltaY
	b.Distance -= utils.GetHypotenuse(deltaX, deltaY)
}

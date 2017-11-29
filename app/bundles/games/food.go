// Package games handles everything related to our game
package games

import (
	"math"

	"github.com/Tarliton/collision2d"
	"github.com/krishamoud/game/app/common/utils"
)

// Food increases player mass
type Food struct {
	ID       string       `json:"id"`
	Point    *utils.Point `json:"point"`
	Hue      int          `json:"hue"`
	Radius   float64      `json:"radius"`
	Mass     float64      `json:"mass"`
	Col      collision2d.Circle
	PlayerID string
	Speed    float64
	Angle    float64
}

// Update moves the ballistic
func (f *Food) Update(g *Game) {
	if f.Speed <= 0 {
		return
	}
	deltaY := f.Speed * math.Sin(f.Angle)
	deltaX := f.Speed * math.Cos(f.Angle)
	f.Point.Y += deltaY
	f.Point.X += deltaX
	f.Col.Pos.X += deltaX
	f.Col.Pos.Y += deltaY
	if f.Point.X > c.GameWidth-f.Radius {
		f.Point.X = c.GameWidth - f.Radius
		f.Col.Pos.X = c.GameWidth - f.Radius
		f.Speed = 0
	}
	if f.Point.Y > c.GameHeight-f.Radius {
		f.Point.Y = c.GameHeight - f.Radius
		f.Col.Pos.Y = f.Radius
		f.Speed = 0
	}

	if f.Point.X < f.Radius {
		f.Point.X = f.Radius
		f.Col.Pos.X = f.Radius
		f.Speed = 0
	}
	if f.Point.Y < f.Radius {
		f.Point.Y = f.Radius
		f.Col.Pos.Y = f.Radius
		f.Speed = 0
	}

	f.Speed -= 0.1
}

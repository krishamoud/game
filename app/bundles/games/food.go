// Package games handles everything related to our game
package games

import (
	"github.com/Tarliton/collision2d"
	"github.com/krishamoud/game/app/common/utils"
)

// Food increases player mass
type Food struct {
	ID     string       `json:"id"`
	Point  *utils.Point `json:"point"`
	Hue    int          `json:"hue"`
	Radius float64      `json:"radius"`
	Mass   float64      `json:"mass"`
	Col    collision2d.Circle
}

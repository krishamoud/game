// Package games handles everything related to our game
package games

import "github.com/krishamoud/game/app/common/utils"

// Mass is the discarded mass from players
type Mass struct {
	Point  *utils.Point
	Target *utils.Point
	Speed  float64
	Radius float64
}

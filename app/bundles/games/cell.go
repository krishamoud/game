// Package games handles everything related to our game
package games

import "github.com/krishamoud/game/app/common/utils"

// Cell is the player body
type Cell struct {
	Point  *utils.Point `json:"cell"`
	Radius float64      `json:"radius"`
	Mass   float64      `json:"mass"`
	Speed  float64      `json:"speed"`
}

// Package utils is a package that handles all utility functions
package utils

import (
	crypto "crypto/rand"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strconv"
	"time"

	"github.com/krishamoud/game/app/common/conf"
)

var cfg = conf.AppConf

// Point is a position in xy space
type Point struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Radius float64 `json:"radius"`
}

// Color is a struct with border and fill properties
type Color struct {
	Fill   string
	Border string
}

// ValidNickname returns a true if a nickname is valid false otherwise
func ValidNickname(nickname string) bool {
	return true
}

// MassToWidth returns a width based on a mass
func MassToWidth(mass float64) float64 {
	return 4 + mass*6
}

// MassToRadius determines the radius based on mass
func MassToRadius(mass float64) float64 {
	return 4 + math.Sqrt(float64(mass))*6
}

// GetDistance returns the distance between two points
func GetDistance(p1, p2 *Point) float64 {
	return math.Sqrt(math.Pow(float64(p2.X-p1.X), 2) + math.Pow(float64(p2.Y-p1.Y), 2))
}

// GetHypotenuse returns the hypetenuse of a tryangle or distance between two points x,y
func GetHypotenuse(x, y float64) float64 {
	return math.Sqrt(math.Pow(float64(y), 2) + math.Pow(float64(x), 2))
}

// DegreesToRadians returns radians from DegreesToRadians
func DegreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// GetRadianAngle returns the angle between two points
func GetRadianAngle(p1, p2 *Point) float64 {
	return math.Atan2(p2.Y-p1.Y, p2.X-p1.X)
}

// RandomInRange generates a random position in a RandomInRange
func RandomInRange(from, to float64) float64 {
	dif := big.NewFloat(to - from)
	difInt, _ := dif.Int(nil)
	p, _ := crypto.Int(crypto.Reader, difInt)
	f := new(big.Float).SetInt(p)
	f2, _ := f.Float64()
	return f2
}

// RandomPosition generates a random position within the field of play
func RandomPosition(radius float64) *Point {
	x := RandomInRange(0, cfg.GameWidth)
	y := RandomInRange(0, cfg.GameHeight)
	return &Point{
		X: x,
		Y: y,
	}
}

// Log returns log(n) / log(base) if base exists else returns log(n)
func Log(n, base float64) float64 {
	if base != 0 {
		return math.Log(n) / math.Log(base)
	}
	return math.Log(n)
}

// UniformPosition distributes returns a single point that is evenly distributed
func UniformPosition(points []*Point, radius float64) *Point {
	var bestCandidate *Point
	var maxDistance float64
	var numberOfCandidates = 10
	if len(points) == 0 {
		return RandomPosition(radius)
	}
	for i := 0; i < numberOfCandidates; i++ {
		var minDistance = math.MaxFloat64
		candidate := RandomPosition(radius)
		for _, p := range points {
			distance := GetDistance(candidate, p)
			if distance < minDistance {
				minDistance = distance
			}
		}

		if minDistance > maxDistance {
			bestCandidate = candidate
			maxDistance = minDistance
		} else {
			return RandomPosition(radius)
		}
	}
	return bestCandidate
}

// RandomColor generates a random fill and stroke color
func RandomColor() *Color {
	rand.Seed(time.Now().Unix())
	num := rand.Intn(1<<24) | 0
	cstr := fmt.Sprintf("%05d", num)
	hexN, _ := strconv.ParseInt(cstr, 10, 64)
	hexC := "#" + strconv.FormatInt(hexN, 16)
	r, _ := strconv.ParseInt(hexC[1:3], 16, 64)
	r -= 32
	g, _ := strconv.ParseInt(hexC[3:5], 16, 64)
	g -= 32
	b, _ := strconv.ParseInt(hexC[5:], 16, 64)
	b -= 32
	if r <= 0 {
		r = 0
	}
	if g <= 0 {
		g = 0
	}
	if b <= 0 {
		b = 0
	}
	borderN := (1 << 24) + (r << 16) + (g << 8) + b
	borderC := "#" + strconv.FormatInt(borderN, 16)[1:]
	return &Color{
		Border: borderC,
		Fill:   hexC,
	}
}

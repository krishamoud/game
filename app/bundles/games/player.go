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

const (
	circle = "circle"
	square = "square"
)

// Player controls an individual player state
type Player struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Point         *utils.Point       `json:"point"`
	W             float64            `json:"w"`
	H             float64            `json:"h"`
	Cells         []*Cell            `json:"cells"`
	MassTotal     float64            `json:"massTotal"`
	MassCurrent   float64            `json:"massCurrent"`
	Hue           int                `json:"hue"`
	Type          string             `json:"type"`
	LastHeartbeat time.Time          `json:"lastHeartBeat"`
	Target        *utils.Point       `json:"target"`
	LastSplit     time.Time          `json:"lastSplit"`
	Conn          *Client            `json:"conn"`
	ScreenWidth   float64            `json:"screenWidth"`
	ScreenHeight  float64            `json:"screenHeight"`
	Shape         string             `json:"shape"`
	Circle        collision2d.Circle `json:"circle"`
	Box           collision2d.Box    `json:"box"`
	EyeAngle      float64            `json:"eyeAngle"`
	Scale         float64            `json:"scale"`
	EyeLength     float64            `json:"eyeLength"`
	ClipSize      int                `json:"clipSize"`
	ShotsLeft     int                `json:"shotsLeft"`
	msgChan       chan string
	lastShot      time.Time
	mu            *sync.Mutex
	sprinting     bool
	sprintStart   time.Time
	invinc        bool
	invincStart   time.Time
}

// NewPlayer returns a new instance of a player
func NewPlayer(t string, cn *Client) *Player {
	radius := utils.MassToRadius(c.DefaultPlayerMass)
	position := utils.RandomPosition(radius)
	cells := []*Cell{}
	var shape string
	var massTotal float64
	if t == player {
		if MainGame.Users.Len()&1 == 1 {
			shape = circle
		} else {
			shape = square
		}
		cell := &Cell{
			Mass: c.DefaultPlayerMass,
			Point: &utils.Point{
				X: position.X,
				Y: position.Y,
			},
			Radius: radius,
		}
		cells = append(cells, cell)
		massTotal = c.DefaultPlayerMass
	}
	currentPlayer := &Player{
		ID:            db.RandomID(12),
		Point:         position,
		W:             c.DefaultPlayerMass,
		H:             c.DefaultPlayerMass,
		Cells:         cells,
		MassTotal:     massTotal,
		Hue:           rand.Intn(360),
		Type:          cn.Type,
		LastHeartbeat: time.Now(),
		Target:        &utils.Point{X: 0, Y: 0},
		Conn:          cn,
		mu:            new(sync.Mutex),
		Shape:         shape,
		msgChan:       make(chan string),
	}
	return currentPlayer
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
func (p *Player) VisibleFood(g *Game) []*Food {
	vf := []*Food{}
	div := math.Min(p.ScreenWidth/4, p.ScreenHeight/4)
	scale := div / p.W
	scaledW := p.ScreenWidth / scale
	scaledH := p.ScreenHeight / scale
	count := 0
	for e := g.Food.Front(); e != nil; e = e.Next() {
		f := e.Value.(*Food)
		if f.Point.X > p.Point.X-scaledW/2 &&
			f.Point.X < p.Point.X+scaledW/2 &&
			f.Point.Y > p.Point.Y-scaledH/2 &&
			f.Point.Y < p.Point.Y+scaledH/2 {
			// if count > 200 {
			// 	continue
			// }
			vf = append(vf, f)
			count++
		}
	}
	return vf
}

// VisibleBallistics returns all ballistics the player can see based on their window size
func (p *Player) VisibleBallistics(g *Game) []*Ballistic {
	vb := []*Ballistic{}
	div := math.Min(p.ScreenWidth/4, p.ScreenHeight/4)
	scale := div / p.W
	scaledW := p.ScreenWidth / scale
	scaledH := p.ScreenHeight / scale
	count := 0
	for e := g.Ballistics.Front(); e != nil; e = e.Next() {
		b := e.Value.(*Ballistic)
		if b.Point.X > p.Point.X-scaledW/2 &&
			b.Point.X < p.Point.X+scaledW/2 &&
			b.Point.Y > p.Point.Y-scaledH/2 &&
			b.Point.Y < p.Point.Y+scaledH/2 {
			if count > 200 {
				continue
			}
			vb = append(vb, b)
			count++
		}
	}
	return vb
}

// VisibleCells returns the player cells visible based on the player window size
func (p *Player) VisibleCells(g *Game) []*Player {
	vc := []*Player{}
	div := math.Max(p.ScreenWidth/4, p.ScreenHeight/4)
	scale := div / p.W
	scaledW := p.ScreenWidth / scale
	scaledH := p.ScreenHeight / scale
	for e := g.Users.Front(); e != nil; e = e.Next() {
		u := e.Value.(*Player)
		if u.Shape == circle {
			for _, c := range u.Cells {
				if c.Point.X+c.Radius > p.Point.X-scaledW/2-40 &&
					c.Point.X-c.Radius < p.Point.X+scaledW/2+40 &&
					c.Point.Y+c.Radius > p.Point.Y-scaledH/2-40 &&
					c.Point.Y-c.Radius < p.Point.Y+scaledH/2+40 {
					if u.ID != p.ID {
						pl := &Player{
							ID:          u.ID,
							Point:       u.Point,
							Cells:       u.Cells,
							MassTotal:   u.MassTotal,
							MassCurrent: u.MassCurrent,
							Hue:         u.Hue,
							Name:        u.Name,
							Shape:       u.Shape,
							W:           u.W,
							H:           u.H,
							EyeAngle:    u.EyeAngle,
							EyeLength:   u.EyeLength,
						}
						vc = append(vc, pl)
					} else {
						pl := &Player{
							Point:       u.Point,
							Cells:       u.Cells,
							MassTotal:   u.MassTotal,
							MassCurrent: u.MassCurrent,
							Hue:         u.Hue,
							Shape:       u.Shape,
							W:           u.W,
							H:           u.H,
							EyeAngle:    u.EyeAngle,
							EyeLength:   u.EyeLength,
						}
						vc = append(vc, pl)
					}
				}
			}
		} else {
			if u.Point.X+u.W/2 > p.Point.X-scaledW/2-40 &&
				u.Point.X-u.W/2 < p.Point.X+scaledW/2+40 &&
				u.Point.Y+u.W/2 > p.Point.Y-scaledH/2-40 &&
				u.Point.Y-u.W/2 < p.Point.Y+scaledH/2+40 {
				if u.ID != p.ID {
					p := &Player{
						ID:          u.ID,
						Point:       u.Point,
						Cells:       u.Cells,
						MassTotal:   u.MassTotal,
						MassCurrent: u.MassCurrent,
						Hue:         u.Hue,
						Name:        u.Name,
						Shape:       u.Shape,
						W:           u.W,
						H:           u.H,
						EyeAngle:    u.EyeAngle,
						EyeLength:   u.EyeLength,
					}
					vc = append(vc, p)
				} else {
					p := &Player{
						Point:       u.Point,
						Cells:       u.Cells,
						MassTotal:   u.MassTotal,
						MassCurrent: u.MassCurrent,
						Hue:         u.Hue,
						Shape:       u.Shape,
						W:           u.W,
						H:           u.H,
						EyeAngle:    u.EyeAngle,
						EyeLength:   u.EyeLength,
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
	if p.Shape == circle {
		p.Circle = collision2d.NewCircle(collision2d.Vector{
			X: p.Point.X,
			Y: p.Point.Y,
		}, r)
	} else {
		p.Box = collision2d.NewBox(collision2d.Vector{
			X: p.Point.X - p.W/2,
			Y: p.Point.Y - p.H/2,
		}, p.W, p.H)
	}
}

// SetSpeed sets a default speed to 6.25 if a cell is new
func (p *Player) SetSpeed() {
	for _, cl := range p.Cells {
		if cl.Speed == 0 {
			cl.Speed = 5.00
		}
	}
}

// GetCollisions gets all collideable things in a game
func (p *Player) GetCollisions(g *Game) []quadtree.Bounds {
	b := quadtree.Bounds{}
	if p.Shape == circle {
		b = quadtree.Bounds{
			X:      p.Point.X,
			Y:      p.Point.Y,
			Width:  utils.MassToRadius(p.MassTotal),
			Height: utils.MassToRadius(p.MassTotal),
		}
	} else if p.Shape == square {
		b = quadtree.Bounds{
			X:      p.Point.X - p.W/2,
			Y:      p.Point.Y - p.H/2,
			Width:  p.W,
			Height: p.H,
		}
	}
	return g.Quadtree.Retrieve(b)
}

// AddMass adds mass and increases size accordingly
func (p *Player) AddMass(m float64) {
	if p.MassCurrent >= p.MassTotal {
		p.MassCurrent += m
		p.Cells[0].Mass += m
		p.MassTotal += m
		p.Cells[0].Radius = utils.MassToRadius(p.Cells[0].Mass)
		if p.MassCurrent > p.MassTotal {
			p.MassCurrent = p.MassTotal
		}
	} else {
		p.MassCurrent += m
	}
}

// RemoveMass adds mass and increases size accordingly
func (p *Player) RemoveMass(m float64) {
	p.Cells[0].Mass -= m
	p.MassTotal -= m
	p.Cells[0].Radius = utils.MassToRadius(p.Cells[0].Mass)
}

// FoodCollision Checks the collisions between food and players
func (p *Player) FoodCollision(f *Food) bool {
	if p.Shape == circle {
		return p.CheckCircleCollision(f.Col) && f.PlayerID != p.ID
	}
	return p.CheckBoxCollision(f.Col) && f.PlayerID != p.ID
}

// CheckCollisions checks if the player has collided with something
func (p *Player) CheckCollisions(collidablePoints []quadtree.Bounds, g *Game) {
	for _, col := range collidablePoints {
		switch col.P {
		case "food":
			f := &Food{}
			for e := g.Food.Front(); e != nil; e = e.Next() {
				// do something with e.Value
				if col.ID == e.Value.(*Food).ID {
					f = e.Value.(*Food)
					continue
				}
			}
			if p.FoodCollision(f) {
				p.AddMass(f.Mass)
				g.SpliceFood(col.ID)
			}
			break
		case "ballistic":
			b := g.GetBallistic(col.ID)
			if b != nil && b.PlayerID != p.ID {
				p.BallisticCollision(b, g)
			}
			break
		default:
			break
		}
	}
}

// BallisticCollision checks if a player collides with a ballistic
func (p *Player) BallisticCollision(b *Ballistic, g *Game) {
	bloodTotal := 0.1 * p.MassTotal
	bloodLeak := (b.Mass / p.MassTotal) * bloodTotal
	dmg := b.Mass
	if dmg > p.MassTotal {
		dmg *= 0.1
	}
	if p.IsCircle() {
		if ok, col := collision2d.TestCircleCircle(p.Circle, b.circle); ok {
			p.Leak(col.OverlapV, bloodLeak, b.Degree, g)
			p.MassCurrent -= dmg
			g.RemoveBallistic(b.ID)
		}
	} else {
		if ok, col := collision2d.TestPolygonCircle(p.Box.ToPolygon(), b.circle); ok {
			p.Leak(col.OverlapV, bloodLeak, b.Degree, g)
			p.MassCurrent -= dmg
			g.RemoveBallistic(b.ID)
		}
	}
}

// Leak spills blood on the map
func (p *Player) Leak(col collision2d.Vector, bloodMass float64, angle float64, g *Game) {
	var m float64
	for bloodMass > 0 {
		if bloodMass < 1 {
			m = bloodMass
		} else {
			m = 1
		}
		v := collision2d.NewVector(p.Cells[0].Point.X-col.X, p.Cells[0].Point.Y-col.Y)
		r := utils.MassToRadius(m)
		f := &Food{
			ID: db.RandomID(12),
			Point: &utils.Point{
				X: p.Cells[0].Point.X - col.X,
				Y: p.Cells[0].Point.Y - col.Y,
			},
			Hue:      p.Hue,
			Radius:   r,
			Mass:     m,
			Col:      collision2d.NewCircle(v, r),
			PlayerID: p.ID,
			Speed:    5,
			Angle:    angle * math.Pi,
		}
		g.PushFood(f)
		bloodMass--
	}
}

// CheckKillPlayer checks to see if we should kill the player
func (p *Player) CheckKillPlayer(g *Game) {
	if p.MassCurrent < 0 {
		p.Explode(g)
		p.KillMessage()
		g.RemovePlayerConnection(p)
		g.SpliceUser(p.ID)
	}
}

// Explode turns the player to mush and spreads their mass in a radius
func (p *Player) Explode(g *Game) {
	bloodTotal := 0.9 * p.MassTotal
	var m float64
	for i := float64(0); i < bloodTotal; i++ {
		rand.Seed(time.Now().Unix())
		angle := rand.Float64() * math.Pi * 2
		s := rand.Float64() * 10
		v := collision2d.NewVector(p.Cells[0].Point.X, p.Cells[0].Point.Y)
		r := utils.MassToRadius(m)
		f := &Food{
			ID: db.RandomID(12),
			Point: &utils.Point{
				X: p.Cells[0].Point.X,
				Y: p.Cells[0].Point.Y,
			},
			Hue:      p.Hue,
			Radius:   r,
			Mass:     1,
			Col:      collision2d.NewCircle(v, r),
			PlayerID: p.ID,
			Speed:    s,
			Angle:    angle,
		}
		g.PushFood(f)
	}
}

// PlayerObstructions checks if two players have collided
func (p *Player) PlayerObstructions(cols []quadtree.Bounds) ([]*Player, []*Player) {
	bigger := []*Player{}
	smaller := []*Player{}
	for _, v := range cols {
		u := v.Obj.(*Player)
		if u.MassTotal > p.MassTotal {
			bigger = append(bigger, u)
		} else {
			smaller = append(smaller, u)
		}
	}
	return bigger, smaller
}

// SquarePlayerCollision returns a if a square player collided with another player
func (p *Player) SquarePlayerCollision(u *Player) (bool, collision2d.Response) {
	if p.EqualShape(u) {
		return collision2d.TestPolygonPolygon(p.Box.ToPolygon(), u.Box.ToPolygon())
	}
	return collision2d.TestPolygonCircle(p.Box.ToPolygon(), u.Circle)
}

// CirclePlayerCollision returns if a circle player collided with another player
func (p *Player) CirclePlayerCollision(u *Player) (bool, collision2d.Response) {
	if p.EqualShape(u) {
		return collision2d.TestCircleCircle(p.Circle, u.Circle)
	}
	return collision2d.TestCirclePolygon(p.Circle, u.Box.ToPolygon())
}

// Obstructed returns true if a player can't move in a direction
func (p *Player) Obstructed(players []*Player) {
	var col bool
	res := collision2d.Response{}
	for _, u := range players {
		if p.IsCircle() {
			col, res = p.CirclePlayerCollision(u)
		} else {
			col, res = p.SquarePlayerCollision(u)
		}
		if col {
			p.Cells[0].Point.X -= res.OverlapV.X
			p.Cells[0].Point.Y -= res.OverlapV.Y
		}
	}
}

// ChangeInvinc sets invinc true or false and changes the time as well
func (p *Player) ChangeInvinc() {
	p.invinc = !p.invinc
	p.invincStart = time.Now()
}

// IsCircle returns true if shape is circle
func (p *Player) IsCircle() bool {
	return p.Shape == circle
}

// IsSquare returns true if shape is square
func (p *Player) IsSquare() bool {
	return p.Shape == square
}

// EqualShape returns true if two shapes are the same
func (p *Player) EqualShape(u *Player) bool {
	return p.Shape == u.Shape
}

// ChangeShape changes the player from one shape to the other
func (p *Player) ChangeShape() {
	p.ChangeInvinc()
	if p.Shape == circle {
		p.Shape = square
	} else {
		p.Shape = circle
	}
}

// CheckAmmo returns a bool if the player needs to reload or not
func (p *Player) CheckAmmo() bool {
	return p.ClipSize > p.ShotsLeft
}

// reload reloads the players weapon
func (p *Player) reload() {
	now := time.Now()
	t := p.lastShot.Add(time.Second)
	if now.After(t) && p.CheckAmmo() {
		p.ShotsLeft++
		p.lastShot = now
	}
}

// Invincible returns a bool depending on if a player is Invincible
func (p *Player) Invincible() bool {
	t := time.Millisecond * 500
	return p.invinc && time.Since(p.invincStart) < t
}

// Bigger returns a bool if a player is big enough to eat another player
func (p *Player) Bigger(u *Player) bool {
	return p.MassTotal > u.MassTotal
}

// KillMessage creates and sends an RIP message to a user
func (p *Player) KillMessage() {
	str := "You were eaten"
	var m = struct {
		Msg string `jsong:"msg"`
	}{
		str,
	}
	body, _ := json.MarshalIndent(&m, "", "\t")
	p.Emit("RIP", body)
}

// StartSprinting decreases the players mass and sets the sprint value
func (p *Player) StartSprinting(g *Game) {
	if p.MassTotal <= c.DefaultPlayerMass*1.2 || p.sprinting {
		return
	}

	// Start sprinting
	p.sprintStart = time.Now()
	p.sprinting = true

	// Lose 20% of current mass
	massLost := p.MassTotal * 0.2
	p.MassTotal -= massLost
	r := utils.MassToRadius(p.MassTotal)
	if len(p.Cells) > 0 {
		p.Cells[0].Radius = r
		p.Cells[0].Mass -= massLost
	}

	// push the food into the game
	for massLost > 0 {
		var foodMass float64
		if massLost < 10 {
			foodMass = massLost
		} else {
			foodMass = 10
		}

		// create the food radius
		radius := utils.MassToRadius(foodMass)

		// Create a random point to help distribute the food a little
		rx := utils.RandomInRange(-radius*2, radius*2)
		ry := utils.RandomInRange(-radius*2, radius*2)

		// Create the point where food will appear
		position := &utils.Point{
			X: p.Point.X - rx,
			Y: p.Point.Y - ry,
		}

		// create a vector for the collider
		colPoint := collision2d.NewVector(position.X, position.Y)

		g.PushFood(&Food{
			ID:       db.RandomID(12),
			Point:    position,
			Radius:   radius,
			Mass:     foodMass,
			Hue:      p.Hue,
			Col:      collision2d.NewCircle(colPoint, radius),
			PlayerID: p.ID,
		})

		massLost -= 10
	}
}

// Fire shoots your weapon
func (p *Player) Fire(g *Game) {
	tp := utils.Point{
		X: p.Point.X,
		Y: p.Point.Y,
	}
	target := &utils.Point{
		X: tp.X + p.Target.X,
		Y: tp.Y + p.Target.Y,
	}
	deg := math.Atan2(target.Y-tp.Y, target.X-tp.X)
	d2 := deg + utils.DegreesToRadians(15)
	d3 := deg - utils.DegreesToRadians(15)
	p1 := &utils.Point{
		X: tp.X + math.Cos(deg)*(p.W/2),
		Y: tp.Y + math.Sin(deg)*(p.H/2),
	}
	p2 := &utils.Point{
		X: tp.X + math.Cos(d2)*(p.W/2),
		Y: tp.Y + math.Sin(d2)*(p.H/2),
	}
	p3 := &utils.Point{
		X: tp.X + math.Cos(d3)*(p.W/2),
		Y: tp.Y + math.Sin(d3)*(p.H/2),
	}
	var baseSpeed float64 = 15
	var b1, b2, b3 *Ballistic
	w := p.W
	dist := 8 * w
	pID := p.ID
	mass := p.MassTotal * 0.1
	if p.ShotsLeft > 0 {
		p.lastShot = time.Now()
		if p.Shape == circle {
			mass = mass / 3
			b1 = NewBallistic(pID, baseSpeed, mass, p1, deg, dist)
			b2 = NewBallistic(pID, baseSpeed, mass, p2, d2, dist)
			b3 = NewBallistic(pID, baseSpeed, mass, p3, d3, dist)
			g.Ballistics.PushFront(b1)
			g.Ballistics.PushFront(b2)
			g.Ballistics.PushFront(b3)
		} else {
			b1 = NewBallistic(pID, baseSpeed, mass, p1, deg, dist)
			g.Ballistics.PushFront(b1)
		}
		p.ShotsLeft--
	}
}

// ShouldSprint returns true if the player should sprintStart
func (p *Player) ShouldSprint() bool {
	t := p.sprintStart
	s := time.Millisecond * 1500
	if p.sprinting && time.Since(t) < s {
		return true
	}
	p.sprinting = false
	return false
}

// GetPlayerCollisions returns only player colliders
func (p *Player) GetPlayerCollisions(col []quadtree.Bounds) []quadtree.Bounds {
	b := []quadtree.Bounds{}
	for _, v := range col {
		if v.P == "user" && v.ID != p.ID {
			b = append(b, v)
		}
	}
	return b
}

func (p *Player) movePlayer(cols []quadtree.Bounds) {
	var x, y float64
	for i, cl := range p.Cells {
		target := &utils.Point{
			X: p.Point.X - cl.Point.X + p.Target.X,
			Y: p.Point.Y - cl.Point.Y + p.Target.Y,
		}
		bp, _ := p.PlayerObstructions(cols)
		var dist float64
		inv := p.Invincible()
		if !inv {
			p.invinc = false
			dist = utils.GetHypotenuse(target.X, target.Y)
		}

		deg := math.Atan2(float64(target.Y), float64(target.X))
		p.EyeAngle = deg
		cl.Speed = 5.00
		if p.ShouldSprint() || inv {
			cl.Speed = 7.25
		}
		deltaX := cl.Speed * math.Cos(deg)
		deltaY := cl.Speed * math.Sin(deg)
		if dist < cl.Radius/3 {
			deltaY *= dist / float64((cl.Radius / 3))
			deltaX *= dist / float64((cl.Radius / 3))
		}
		s := deltaX / (cl.Speed * math.Cos(deg))
		p.EyeLength = s
		cl.Point.Y += deltaY
		cl.Point.X += deltaX
		p.Obstructed(bp)
		borderCalc := p.Cells[i].Radius / 3
		if p.Shape == circle {
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
		} else if p.Shape == square {
			if p.Cells[i].Point.X > c.GameWidth-p.W/2 {
				p.Cells[i].Point.X = c.GameWidth - p.W/2
			}
			if p.Cells[i].Point.Y > c.GameHeight-p.H/2 {
				p.Cells[i].Point.Y = c.GameHeight - p.H/2
			}

			if p.Cells[i].Point.X <= 0 {
				p.Cells[i].Point.X = 0
			}
			if p.Cells[i].Point.Y <= 0 {
				p.Cells[i].Point.Y = 0
			}
		}

		x += p.Cells[i].Point.X
		y += p.Cells[i].Point.Y
	}
	p.Point.X = x / float64(len(p.Cells))
	p.Point.Y = y / float64(len(p.Cells))
}

func (p *Player) checkHeartbeat(g *Game) {
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
		g.SpliceUser(p.ID)
	}
}

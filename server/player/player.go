package player

import (
	"fmt"
	"github.com/aimjel/nitrate/server/network"
	"github.com/aimjel/nitrate/server/world/entity"
	"math"
	"sync/atomic"

	"github.com/aimjel/nitrate/server/world"
)

// Player implements a player entity
type Player struct {
	wrld *world.World

	dimension *world.Dimension

	gameMode atomic.Int32

	x, y, z atomic.Uint64

	yaw, pitch atomic.Uint32

	pose atomic.Int32

	Session *network.Session

	Handler Handler
}

func (p *Player) Interact(target entity.Entity) {
	fmt.Println("player touched", target)
}

func (p *Player) Attack(e entity.Entity) {
	p.Session.Knockback(0, 4000, 4000)
}

func New(wrld *world.World) *Player {
	return &Player{
		wrld:      wrld,
		dimension: wrld.OverWorld(),
	}
}

func (p *Player) GameMode() int {
	return int(GameMode(p.gameMode.Load()))
}

func (p *Player) Dimension() *world.Dimension {
	return p.dimension
}

func (p *Player) Move(x, y, z float64) {
	p.x.Store(math.Float64bits(x))
	p.y.Store(math.Float64bits(y))
	p.z.Store(math.Float64bits(z))

	if p.Handler != nil {
		p.Handler.Move(p)
	}

	//todo add anti cheat checks
}

func (p *Player) Rotate(yaw, pitch float32) {
	p.yaw.Store(math.Float32bits(yaw))
	p.pitch.Store(math.Float32bits(pitch))
}

func (p *Player) Position() (x, y, z float64) {
	x = math.Float64frombits(p.x.Load())
	y = math.Float64frombits(p.y.Load())
	z = math.Float64frombits(p.z.Load())
	return
}

func (p *Player) Rotation() (yaw, pitch float32) {
	yaw = math.Float32frombits(p.yaw.Load())
	pitch = math.Float32frombits(p.pitch.Load())
	return
}

func (p *Player) Pose() int32 {
	return p.pose.Load()
}

func (p *Player) SetPose(pose int32) {
	p.pose.Store(pose)
}

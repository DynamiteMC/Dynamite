package pos

import (
	"math"
	"sync/atomic"
)

type EntityPosition struct {
	x, y, z,
	yaw, pitch atomic.Value
	onGround *atomic.Bool
}

func NewEntityPosition() *EntityPosition {
	var e EntityPosition
	e.SetX(0)
	e.SetY(0)
	e.SetZ(0)
	e.SetYaw(0)
	e.SetPitch(0)

	e.onGround = &atomic.Bool{}
	return &e
}

func (pos *EntityPosition) X() float64 {
	return pos.x.Load().(float64)
}

func (pos *EntityPosition) Y() float64 {
	return pos.y.Load().(float64)
}

func (pos *EntityPosition) Z() float64 {
	return pos.z.Load().(float64)
}

func (pos *EntityPosition) Yaw() float32 {
	return pos.yaw.Load().(float32)
}

func (pos *EntityPosition) Pitch() float32 {
	return pos.pitch.Load().(float32)
}

func (pos *EntityPosition) OnGround() bool {
	return pos.onGround.Load()
}

func (pos *EntityPosition) SetX(x float64) {
	pos.x.Store(x)
}

func (pos *EntityPosition) SetY(y float64) {
	pos.y.Store(y)
}

func (pos *EntityPosition) SetZ(z float64) {
	pos.z.Store(z)
}

func (pos *EntityPosition) SetPosition(x, y, z float64) {
	pos.x.Store(x)
	pos.y.Store(y)
	pos.z.Store(z)
}

func (pos *EntityPosition) SetRotation(y, p float32) {
	pos.yaw.Store(y)
	pos.pitch.Store(p)
}

func (pos *EntityPosition) All() (x, y, z float64, yaw, pitch float32, ong bool) {
	return pos.X(), pos.Y(), pos.Z(), pos.Yaw(), pos.Pitch(), pos.OnGround()
}

func (pos *EntityPosition) SetAll(x, y, z float64, yaw, pitch float32, ong bool) {
	pos.SetX(x)
	pos.SetY(y)
	pos.SetZ(z)

	pos.SetYaw(yaw)
	pos.SetPitch(pitch)

	pos.SetOnGround(ong)
}

func (pos *EntityPosition) SetYaw(y float32) {
	pos.yaw.Store(y)
}

func (pos *EntityPosition) SetPitch(p float32) {
	pos.pitch.Store(p)
}

func (pos *EntityPosition) SetOnGround(ong bool) {
	pos.onGround.Store(ong)
}

func DegreesToAngle(degrees float32) byte {
	return byte(math.Round(float64(degrees) * (256.0 / 360.0)))
}

func PositionIsValid(x, y, z float64) bool {
	return !math.IsNaN(x) && !math.IsNaN(y) && !math.IsNaN(z) &&
		!math.IsInf(x, 0) && !math.IsInf(y, 0) && !math.IsInf(z, 0)
}

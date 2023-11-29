package pos

import (
	"fmt"
	"math"

	"github.com/dynamitemc/dynamite/util/atomic"
)

type EntityPosition struct {
	x, y, z    *atomic.Value[float64]
	yaw, pitch *atomic.Value[float32]
	onGround   *atomic.Value[bool]
}

func NewEntityPosition(x, y, z float64, yaw, pitch float32, ong bool) *EntityPosition {
	var e EntityPosition
	e.x = atomic.NewValue(x)
	e.y = atomic.NewValue(y)
	e.z = atomic.NewValue(z)
	e.yaw = atomic.NewValue(yaw)
	e.pitch = atomic.NewValue(pitch)
	e.onGround = atomic.NewValue(ong)
	return &e
}

func (pos *EntityPosition) X() float64 {
	return pos.x.Get()
}

func (pos *EntityPosition) Y() float64 {
	return pos.y.Get()
}

func (pos *EntityPosition) Z() float64 {
	return pos.z.Get()
}

func (pos *EntityPosition) Yaw() float32 {
	return pos.yaw.Get()
}

func (pos *EntityPosition) Pitch() float32 {
	return pos.pitch.Get()
}

func (pos *EntityPosition) OnGround() bool {
	return pos.onGround.Get()
}

func (pos *EntityPosition) SetX(x float64) {
	pos.x.Set(x)
}

func (pos *EntityPosition) SetY(y float64) {
	pos.y.Set(y)
}

func (pos *EntityPosition) SetZ(z float64) {
	pos.z.Set(z)
}

func (pos *EntityPosition) SetPosition(x, y, z float64) {
	pos.x.Set(x)
	pos.y.Set(y)
	pos.z.Set(z)
}

func (pos *EntityPosition) SetRotation(y, p float32) {
	pos.yaw.Set(y)
	pos.pitch.Set(p)
}

func (pos *EntityPosition) Position() (x, y, z float64) {
	return pos.x.Get(), pos.y.Get(), pos.z.Get()
}

func (pos *EntityPosition) Rotation() (yaw, pitch float32) {
	return pos.yaw.Get(), pos.pitch.Get()
}

func (pos *EntityPosition) SetYaw(y float32) {
	pos.yaw.Set(y)
}

func (pos *EntityPosition) SetPitch(p float32) {
	pos.pitch.Set(p)
}

func (pos *EntityPosition) SetOnGround(ong bool) {
	pos.onGround.Set(ong)
}

func (pos *EntityPosition) String() string {
	return fmt.Sprintf("(%f %f %f | %f%f)", pos.X(), pos.Y(), pos.Z(), pos.Yaw(), pos.Pitch())
}

func DegreesToAngle(degrees float32) byte {
	return byte(math.Round(float64(degrees) * (256.0 / 360.0)))
}

func PositionIsValid(x, y, z float64) bool {
	return !math.IsNaN(x) && !math.IsNaN(y) && !math.IsNaN(z) &&
		!math.IsInf(x, 0) && !math.IsInf(y, 0) && !math.IsInf(z, 0)
}

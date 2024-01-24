package network

import (
	"github.com/aimjel/nitrate/server/world"
	"github.com/aimjel/nitrate/server/world/entity"
)

// Player describes functions the session can use to update and retrieve states
type Player interface {

	// Dimension returns the dimension the player is currently in
	Dimension() *world.Dimension

	// Position returns the absolute position's of the player
	Position() (float64, float64, float64)

	// Rotation returns the player's absolute yaw and pitch values.
	// The value returned uses mojang's weird format, See DegreesToAngle
	Rotation() (float32, float32)

	// Move moves the player coordinates passed
	Move(x, y, z float64)

	// Rotate rotates the player's head yaw and pitch
	Rotate(yaw, pitch float32)

	// GameMode returns the player's game-mode.
	// See player.GameMode to convert integers into game-mode types
	GameMode() int

	entity.Interactor

	entity.Attacker
}

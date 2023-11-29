package player

import (
	"math"

	"github.com/aimjel/minecraft/packet"
	"github.com/dynamitemc/dynamite/logger"
	"github.com/dynamitemc/dynamite/server/entity/pos"
	"github.com/dynamitemc/dynamite/server/enum"
	"github.com/google/uuid"
)

func (p *Player) handleCenterChunk(oldx, oldz, newx, newz float64) {
	oldChunkX := int(math.Floor(oldx / 16))
	oldChunkZ := int(math.Floor(oldz / 16))

	newChunkX := int(math.Floor(newx / 16))
	newChunkZ := int(math.Floor(newz / 16))

	if newChunkX != oldChunkX || newChunkZ != oldChunkZ {
		p.SendPacket(&packet.SetCenterChunk{
			ChunkX: int32(newChunkX),
			ChunkZ: int32(newChunkZ),
		})
	}
}

func (p *Player) HandleMovement(packetid int32, x1, y1, z1 float64, ya, pi float32, ong bool, teleport bool) {
	oldx, oldy, oldz := p.Position.X(), p.Position.Y(), p.Position.Z()
	p.Position.SetPosition(x1, y1, z1)
	p.Position.SetRotation(ya, pi)
	p.Position.SetOnGround(ong)
	p.handleCenterChunk(oldx, oldy, x1, y1)
	if y1 > p.HighestY.Get() {
		p.HighestY.Set(y1)
	} else {
		if g := p.GameMode.Get(); g == enum.GameModeSurvival || g == enum.GameModeAdventure {
			d := p.HighestY.Get() - y1
			if d > 3 {
				if b := p.OnBlock(); b != nil && b.EncodedName() != "minecraft:air" && b.EncodedName() != "minecraft:water" {
					d -= 3
					if int(d) > 0 {
						p.Damage(float32(d), enum.DamageTypeFall)
					}
					p.HighestY.Set(0)
				}
			}
		}
	}

	logger.Println(p.Position)

	distance := math.Sqrt((x1-oldx)*(x1-oldx) + (y1-oldy)*(y1-oldy) + (z1-oldz)*(z1-oldz))
	/*if distance > 100 && !teleport {
		p.Teleport(oldx, oldy, oldz, yaw, pitch)
		p.logger.Info("%s moved too quickly!", p.Name())
		return
	}*/
	if !pos.PositionIsValid(x1, y1, z1) {
		p.Disconnect(p.lang.Translate("disconnect.invalid_player_movement", nil))
		return
	}

	if distance > 8 {
		packetid = 0
		return
	}

	yaw, pitch := pos.DegreesToAngle(ya), pos.DegreesToAngle(pi)

	var pk packet.Packet
	headRotationPacket := &packet.EntityHeadRotation{
		EntityID: p.EntityID(),
		HeadYaw:  yaw,
	}
	var sendHeadRotation bool
	switch packetid {
	case 0x14: // position
		pk = &packet.EntityPosition{
			EntityID: p.EntityID(),
			X:        int16(((x1 * 32) - oldx*32) * 128),
			Y:        int16(((y1 * 32) - oldy*32) * 128),
			Z:        int16(((z1 * 32) - oldz*32) * 128),
			OnGround: ong,
		}
	case 0x15: // position + rotation
		pk = &packet.EntityPositionRotation{
			EntityID: p.EntityID(),
			X:        int16(((x1 * 32) - oldx*32) * 128),
			Y:        int16(((y1 * 32) - oldy*32) * 128),
			Z:        int16(((z1 * 32) - oldz*32) * 128),
			Yaw:      yaw,
			Pitch:    pitch,
			OnGround: ong,
		}
		sendHeadRotation = true
	case 0x16: // rotation
		pk = &packet.EntityRotation{
			EntityID: p.EntityID(),
			Yaw:      yaw,
			Pitch:    pitch,
			OnGround: ong,
		}
		sendHeadRotation = true
	default:
		pk = &packet.TeleportEntity{
			EntityID: p.EntityID(),
			X:        x1,
			Y:        y1,
			Z:        z1,
			Yaw:      yaw,
			Pitch:    pitch,
			OnGround: ong,
		}
	}

	p.playerController.Range(func(u uuid.UUID, pl *Player) bool {
		if p.Session.UUID() == u {
			return true
		}
		if !pl.InView(p.Position.X(), p.Position.Y(), p.Position.Z()) {
			pl.DespawnEntity(p.EntityID())
			return true
		}

		if !pl.IsSpawned(p.EntityID()) {
			pl.SpawnPlayer(p)
			return true
		}

		pl.SendPacket(pk)
		if sendHeadRotation {
			pl.SendPacket(headRotationPacket)
		}
		return true
	})
}

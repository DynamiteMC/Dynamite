package player

import (
	"fmt"
	"strings"

	"github.com/aimjel/minecraft/chat"
	"github.com/aimjel/minecraft/packet"
	"github.com/aimjel/minecraft/protocol/types"
	"github.com/dynamitemc/dynamite/logger"
	"github.com/dynamitemc/dynamite/server/controller"
	"github.com/dynamitemc/dynamite/server/entity"
	"github.com/dynamitemc/dynamite/server/enum"
	"github.com/dynamitemc/dynamite/server/registry"
	"github.com/dynamitemc/dynamite/server/world"
	"github.com/dynamitemc/dynamite/server/world/chunk"

	"github.com/google/uuid"
)

func (p *Player) BroadcastAnimation(animation uint8) {
	p.playerController.Range(func(u uuid.UUID, pl *Player) bool {
		if !pl.IsSpawned(p.EntityID()) || u == p.Session.UUID() {
			return true
		}

		pl.SendPacket(&packet.EntityAnimation{
			EntityID:  p.EntityID(),
			Animation: animation,
		})
		return true
	})
}

func (p *Player) BreakBlock(pos int64) {
	b := p.Dimension().Block(world.ParsePosition(pos))
	bid, _ := chunk.GetBlockId(b)

	p.playerController.Range(func(_ uuid.UUID, pl *Player) bool {
		if !pl.IsSpawned(p.EntityID()) {
			return true
		}

		pl.SendPacket(&packet.WorldEvent{Event: enum.WorldEventBlockBreak, Location: types.Position(pos), Data: int32(bid)})
		pl.SendPacket(&packet.BlockUpdate{
			Location: types.Position(pos),
		})

		return true
	})
}

func (p *Player) Attack(entityId int32) {
	e := findEntity(p.entityController, p.playerController, entityId)
	if e == nil {
		return
	}
	x, y, z := p.Position.X(), p.Position.Y(), p.Position.Z()
	soundId := int32(519)
	if pl, ok := e.(*Player); ok {
		if pl.GameMode.Get() == enum.GameModeCreative {
			return
		}

		health := pl.Health.Get()
		pl.SetHealth(health - 1)
		pl.SendPacket(&packet.DamageEvent{
			EntityID:        entityId,
			SourceTypeID:    enum.DamageTypePlayerAttack,
			SourceCauseID:   p.EntityID() + 1,
			SourceDirectID:  p.EntityID() + 1,
			SourcePositionX: &x,
			SourcePositionY: &y,
			SourcePositionZ: &z,
		})
	} else {
		entity, ok := e.(entity.LivingEntity)
		if !ok {
			return
		}
		sound, ok := registry.SoundEvent.Get(fmt.Sprintf("minecraft:entity.%s.hurt", strings.TrimPrefix(entity.Type(), "minecraft:")))
		if ok {
			soundId = sound
		}
	}

	p.playerController.Range(func(_ uuid.UUID, pl *Player) bool {
		if !pl.IsSpawned(entityId) {
			return true
		}
		pl.SendPacket(&packet.DamageEvent{
			EntityID:        entityId,
			SourceTypeID:    enum.DamageTypePlayerAttack,
			SourceCauseID:   p.EntityID() + 1,
			SourceDirectID:  p.EntityID() + 1,
			SourcePositionX: &x,
			SourcePositionY: &y,
			SourcePositionZ: &z,
		})
		pl.SendPacket(&packet.EntitySoundEffect{
			Category: enum.SoundCategoryAmbient,
			SoundID:  soundId,
			EntityID: entityId,
			Seed:     world.RandomSeed(),
			Volume:   1,
			Pitch:    1,
		})
		return true
	})
}

func (p *Player) Despawn() {
	p.playerController.Range(func(_ uuid.UUID, pl *Player) bool {
		if !pl.IsSpawned(p.EntityID()) {
			return true
		}
		pl.DespawnEntity(p.EntityID())
		return true
	})
}

func (p *Player) BroadcastGamemode() {
	gm := int32(p.GameMode.Get())

	p.playerController.Range(func(_ uuid.UUID, pl *Player) bool {
		pl.SendPacket(&packet.PlayerInfoUpdate{
			Actions: 0x04,
			Players: []types.PlayerInfo{{
				UUID:     p.Session.UUID(),
				GameMode: gm,
			}},
		})
		return true
	})
}

func (p *Player) sendEquipment(pl *Player) {
	slots := make(map[int8]packet.Slot)
	inv := p.Inventory
	sel := p.Inventory.SelectedSlot.Get()

	for _, s := range inv.Export() {
		switch s.Slot {
		case int8(sel):
			s, _ := s.ToPacketSlot()
			slots[0] = s
		case -106:
			s, _ := s.ToPacketSlot()
			slots[1] = s
		case 100:
			s, _ := s.ToPacketSlot()
			slots[2] = s
		case 101:
			s, _ := s.ToPacketSlot()
			slots[3] = s
		case 102:
			s, _ := s.ToPacketSlot()
			slots[4] = s
		case 103:
			s, _ := s.ToPacketSlot()
			slots[5] = s
		}
	}

	for s, i := range slots {
		pl.SendPacket(&packet.SetEquipment{
			EntityID: p.EntityID(),
			Slot:     s,
			Item:     i,
		})
	}
}

func (p *Player) BroadcastMetadataInArea(pk *packet.SetEntityMetadata) {
	p.playerController.Range(func(u uuid.UUID, pl *Player) bool {
		if !pl.IsSpawned(p.EntityID()) {
			return true
		}
		pl.SendPacket(pk)
		return true
	})
}

func (p *Player) broadcastMetadataGlobal(pk *packet.SetEntityMetadata) {
	p.playerController.Range(func(u uuid.UUID, pl *Player) bool {
		if u == p.Session.UUID() {
			return true
		}
		pl.SendPacket(pk)
		return true
	})
}

func findEntity(
	entities *controller.Controller[int32, entity.Entity],
	players *controller.Controller[uuid.UUID, *Player],
	id int32,
) interface{} {
	if _, p := players.Range(func(_ uuid.UUID, p *Player) bool {
		return p.entityID != id
	}); p != nil {
		return p
	} else {
		e, ok := entities.Get2(id)
		if !ok {
			return nil
		}
		return e
	}
}

func findEntityByUUID(
	entities *controller.Controller[int32, entity.Entity],
	players *controller.Controller[uuid.UUID, *Player],
	id [16]byte,
) interface{} {
	if _, p := players.Range(func(_ uuid.UUID, p *Player) bool {
		return p.Session.UUID() != id
	}); p != nil {
		return p
	}
	if _, e := entities.Range(func(_ int32, e entity.Entity) bool {
		return e.UUID() != id
	}); e != nil {
		return e
	}
	return nil
}

func globalMessage(logger *logger.Logger, players *controller.Controller[uuid.UUID, *Player], message chat.Message) {
	players.Range(func(_ uuid.UUID, p *Player) bool {
		if p.ClientSettings().ChatMode == 2 {
			return true
		}
		p.SendPacket(&packet.SystemChatMessage{
			Message: message,
		})
		return true
	})
	logger.Print(message)
}

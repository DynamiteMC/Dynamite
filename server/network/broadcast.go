package network

import (
	"github.com/aimjel/minecraft/protocol/metadata"
	"sync"

	"github.com/aimjel/minecraft/protocol"

	"github.com/aimjel/minecraft/packet"
	"github.com/aimjel/minecraft/protocol/types"
)

type Broadcast struct {
	sessions map[string]*Session

	mu sync.RWMutex
}

func NewBroadcast() *Broadcast {
	return &Broadcast{
		sessions: make(map[string]*Session),
		mu:       sync.RWMutex{},
	}
}

const (
	sendPosition byte = 1 << iota
	sendRotation
)

func (b *Broadcast) BroadcastMovement(c *Session, dx, dy, dz int16, yawAngle, pitchAngle, sendWhat byte) {
	var pk packet.Packet
	var sendHeadRotation bool

	switch sendWhat {
	case sendPosition | sendRotation:
		//player moved and rotated
		pk = &packet.EntityPositionRotation{
			EntityID: c.eid,
			X:        dx,
			Y:        dy,
			Z:        dz,
			Yaw:      yawAngle,
			Pitch:    pitchAngle,
		}
		sendHeadRotation = true
	case sendPosition:
		//player moved
		pk = &packet.EntityPosition{
			EntityID: c.eid,
			X:        dx,
			Y:        dy,
			Z:        dz,
		}
	case sendRotation:
		// player rotated
		pk = &packet.EntityRotation{
			EntityID: c.eid,
			Yaw:      yawAngle,
			Pitch:    pitchAngle,
		}
		sendHeadRotation = true
	}

	b.mu.RLock()
	defer b.mu.RUnlock()
	x, y, z := c.state.Position()
	for _, pl := range b.sessions {
		if pl.conn.UUID() == c.conn.UUID() {
			continue
		}
		if !pl.InView(x, y, z) {
			// player is not in the field of view, despawn if spawned
			pl.Despawn(c)
			continue
		}
		if !pl.IsSpawned(c) {
			// player is not spawned, spawn
			pl.conn.SendPacket(&packet.SpawnPlayer{
				EntityID:   c.eid,
				PlayerUUID: c.conn.UUID(),
				X:          x,
				Y:          y,
				Z:          z,
				Yaw:        yawAngle,
				Pitch:      pitchAngle,
			})
			pl.spawnEntity(c.eid)
			continue
		}
		pl.conn.SendPacket(pk)
		if sendHeadRotation {
			pl.conn.SendPacket(&packet.EntityHeadRotation{EntityID: c.eid, HeadYaw: yawAngle})
		}
	}
}

func (b *Broadcast) AddSession(c *Session) {
	self := types.PlayerInfo{
		UUID:       c.conn.UUID(),
		Name:       c.conn.Name(),
		Properties: c.conn.Properties(),
		GameMode:   int32(c.state.GameMode()),
		Listed:     true,
	}

	b.mu.RLock()
	players := make([]types.PlayerInfo, 0, len(b.sessions)+1)
	players = append(players, self)
	for _, sesh := range b.sessions {

		//notifies online players of the new session
		sesh.conn.SendPacket(&packet.PlayerInfoUpdate{
			Actions: byte(protocol.AddPlayer | protocol.UpdateGameMode | protocol.UpdateListed),
			Players: []types.PlayerInfo{self},
		})

		players = append(players, types.PlayerInfo{
			UUID:       sesh.conn.UUID(),
			Name:       sesh.conn.Name(),
			Properties: sesh.conn.Properties(),
			GameMode:   int32(sesh.state.GameMode()),
			Listed:     true,
		})
	}
	b.mu.RUnlock()

	b.mu.Lock()
	b.sessions[c.conn.Name()] = c
	b.mu.Unlock()

	//notifies the new session of all the online players, including itself(used for skin data etc)
	c.conn.SendPacket(&packet.PlayerInfoUpdate{
		Actions: byte(protocol.AddPlayer | protocol.UpdateGameMode | protocol.UpdateListed),
		Players: players,
	})
}

func (b *Broadcast) RemoveSessions(sessions ...*Session) {
	uuids := make([][16]byte, 0, len(sessions))
	eids := make([]int32, 0, len(sessions))
	b.mu.Lock()
	for _, s := range sessions {
		uuids = append(uuids, s.conn.UUID())
		eids = append(eids, s.eid)
		delete(b.sessions, s.conn.Name())
	}
	b.mu.Unlock()

	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, sesh := range b.sessions {
		sesh.conn.SendPacket(&packet.PlayerInfoRemove{UUIDs: uuids})
		sesh.conn.SendPacket(&packet.DestroyEntities{EntityIds: eids})
	}
}

func (b *Broadcast) BroadcastEntityMetaData(c *Session, meta metadata.MetaData) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, s := range b.sessions {
		if !s.IsSpawned(c) {
			continue
		}
		s.conn.SendPacket(&packet.SetEntityMetadata{
			EntityID: c.eid,
			MetaData: meta,
		})
	}
}

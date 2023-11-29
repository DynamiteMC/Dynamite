package server

import (
	"github.com/aimjel/minecraft/packet"
	"github.com/dynamitemc/dynamite/server/entity"
	"github.com/dynamitemc/dynamite/server/entity/pos"
	"github.com/dynamitemc/dynamite/server/player"
	"github.com/dynamitemc/dynamite/server/registry"
	"github.com/dynamitemc/dynamite/server/world"
	"github.com/dynamitemc/dynamite/server/world/chunk"
	"github.com/google/uuid"
)

func (srv *Server) FindEntity(id int32) interface{} {
	if p := srv.FindPlayerByID(id); p != nil {
		return p
	} else {
		e, ok := srv.Entities.Get2(id)
		if !ok {
			return nil
		}
		return e
	}
}

func (srv *Server) FindEntityByUUID(id [16]byte) interface{} {
	if _, p := srv.Players.Range(func(_ uuid.UUID, p *player.Player) bool {
		return p.Session.UUID() != id
	}); p != nil {
		return p
	}
	if _, e := srv.Entities.Range(func(_ int32, e entity.Entity) bool {
		return e.UUID() != id
	}); e != nil {
		return e
	}
	return nil
}

func (srv *Server) SpawnEntity(typ string, x, y, z float64) {
	id, ok := registry.EntityType.Get(typ)
	if !ok {
		return
	}
	e := entity.CreateEntity(srv.Entities, srv.NewID(), chunk.Entity{
		Pos:  []float64{x, y, z},
		Id:   typ,
		UUID: world.ByteUUIDToIntUUID(uuid.New()),
	}, srv.World.Overworld())
	p := &packet.SpawnEntity{
		EntityID: e.EntityID(),
		UUID:     e.UUID(),
		X:        x,
		Y:        y,
		Z:        z,
		Type:     id,
	}

	srv.Players.Range(func(_ uuid.UUID, pl *player.Player) bool {
		if !pl.InView(x, y, z) {
			return true
		}
		pl.SpawnEntity(p)
		return true
	})
}

func (srv *Server) SetEntityPosition(id int32, x, y, z float64) {
	e := srv.FindEntity(id)
	if e == nil {
		return
	}
	if e, ok := e.(entity.Entity); ok {
		prevX, prevY, prevZ := e.Position()

		// todo check distance
		dx := int16((x*32 - prevX*32) * 128)
		dy := int16((y*32 - prevY*32) * 128)
		dz := int16((z*32 - prevZ*32) * 128)

		e.SetPosition(x, y, z)

		ong := e.OnGround()
		srv.Players.Range(func(_ uuid.UUID, pl *player.Player) bool {
			if !pl.IsSpawned(id) {
				return true
			}
			pl.SendPacket(&packet.EntityPosition{
				EntityID: id,
				OnGround: ong,
				X:        dx,
				Y:        dy,
				Z:        dz,
			})
			return true
		})
	}
}

func (srv *Server) TeleportEntity(id int32, x, y, z float64) {
	e := srv.FindEntity(id)
	if e == nil {
		return
	}
	if e, ok := e.(entity.Entity); ok {
		e.SetPosition(x, y, z)

		ya, pi := e.Rotation()
		yaw, pitch := pos.DegreesToAngle(ya), pos.DegreesToAngle(pi)
		ong := e.OnGround()
		srv.Players.Range(func(_ uuid.UUID, pl *player.Player) bool {
			if !pl.IsSpawned(id) {
				return true
			}
			pl.SendPacket(&packet.TeleportEntity{
				EntityID: id,
				OnGround: ong,
				X:        x,
				Y:        y,
				Z:        z,
				Yaw:      yaw,
				Pitch:    pitch,
			})
			return true
		})
	}
}

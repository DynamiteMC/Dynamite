package server

import (
	"github.com/aimjel/minecraft/chat"
	"github.com/dynamitemc/dynamite/server/player"
	"github.com/google/uuid"

	"github.com/aimjel/minecraft/packet"
)

func (srv *Server) GlobalMessage(message chat.Message) {
	srv.Players.Range(func(_ uuid.UUID, p *player.Player) bool {
		if p.ClientSettings().ChatMode == 2 {
			return true
		}
		p.SendPacket(&packet.SystemChatMessage{
			Message: message,
		})
		return true
	})
	srv.Logger.Print(message)
}

func (srv *Server) OperatorMessage(message chat.Message) {
	srv.Players.Range(func(_ uuid.UUID, p *player.Player) bool {
		if p.ClientSettings().ChatMode == 2 || !p.Operator.Get() {
			return true
		}
		p.SendPacket(&packet.SystemChatMessage{
			Message: message,
		})
		return true
	})
	srv.Logger.Print(message)
}

func (srv *Server) playerlistRemove(players ...[16]byte) {
	srv.Players.Range(func(_ uuid.UUID, p *player.Player) bool {
		p.SendPacket(&packet.PlayerInfoRemove{UUIDs: players})
		return true
	})
}

package handler

import (
	"github.com/aimjel/minecraft/chat"
	"github.com/dynamitemc/dynamite/server/enum"
	"github.com/dynamitemc/dynamite/server/player"
)

func TeleportToEntity(state *player.Player, uuid [16]byte) {
	if state.GameMode.Get() != enum.GameModeSpectator {
		state.Disconnect(chat.NewMessage("Yo how do you do dat without gamemode spectator?"))
		return
	}
	state.TeleportToEntity(uuid)
}

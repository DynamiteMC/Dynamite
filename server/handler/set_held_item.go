package handler

import "github.com/dynamitemc/dynamite/server/player"

func SetHeldItem(state *player.Player, heldItem int16) {
	state.Inventory.SelectedSlot.Set(int32(heldItem))
}

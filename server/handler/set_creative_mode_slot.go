package handler

import (
	"github.com/aimjel/minecraft/chat"
	"github.com/aimjel/minecraft/packet"
	"github.com/dynamitemc/dynamite/server/enum"
	"github.com/dynamitemc/dynamite/server/inventory"
	"github.com/dynamitemc/dynamite/server/item"
	"github.com/dynamitemc/dynamite/server/player"
)

func SetCreativeModeSlot(state *player.Player, slot int16, data packet.Slot) {
	if state.GameMode.Get() != enum.GameModeCreative {
		state.Disconnect(chat.NewMessage("bruh cant use the creative button without creative"))
		return
	}
	s := inventory.NetworkSlotToDataSlot(slot)
	if !data.Present {
		if s, ok := state.Inventory.Slot(s); ok {
			state.PreviousSelectedSlot.Set(s)
		}
		//state.ClearItem(s)
	} else {
		i, _ := item.PacketSlotToItem(s, data)
		state.Inventory.SetSlot(s, i)
	}
}

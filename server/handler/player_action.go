package handler

import (
	"github.com/aimjel/minecraft/packet"
	"github.com/dynamitemc/dynamite/server/enum"
	"github.com/dynamitemc/dynamite/server/player"
)

func PlayerAction(state *player.Player, pk *packet.PlayerActionServer) {
	switch pk.Status {
	case enum.PlayerActionStartedDigging:
		if state.GameMode.Get() == enum.GameModeCreative {
			state.BreakBlock(int64(pk.Location))
		}
		state.BroadcastMetadataInArea(&packet.SetEntityMetadata{
			EntityID: state.EntityID(),
			Pose:     &dg,
		})
		//go controller.BroadcastDigging(pk.Location)
	case enum.PlayerActionCancelledDigging, enum.PlayerActionFinishedDigging:
		if pk.Status == enum.PlayerActionFinishedDigging {
			state.BreakBlock(int64(pk.Location))
		}
		state.BroadcastMetadataInArea(&packet.SetEntityMetadata{
			EntityID: state.EntityID(),
			Pose:     &st,
		})
	case enum.PlayerActionDropItemStack, enum.PlayerActionDropItem:
		if s, ok := state.Inventory.HeldItem(); ok {
			state.PreviousSelectedSlot.Set(s)
			state.Inventory.DeleteSlot(int8(s.Slot))
		}
		//controller.DropSlot()
	}
}

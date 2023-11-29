package session

import (
	"net"

	"github.com/aimjel/minecraft/packet"
	"github.com/aimjel/minecraft/protocol/types"
)

type Session interface {
	SendPacket(pk packet.Packet) error
	ReadPacket() (packet.Packet, error)
	Close(err error)
	Name() string
	UUID() [16]byte
	Properties() []types.Property
	RemoteAddr() net.Addr

	/*SetHealth(health float32, food int32, foodSaturation float32)
	GameEvent(event uint8, value float32)
	EntityEvent(entityId int32, event int8)
	PlayerInfoUpdate(actions uint8, players []types.PlayerInfo)
	Respawn(dimension *world.Dimension, gameMode uint8, dataKept uint8, deathDimensionName string, deathLocation uint64, partialCooldown int32)
	SetContainerContent(windowId uint8, stateId int32, slots []packet.Slot)
	SetCenterChunk(x, z int32)
	SetDefaultSpawnPosition(pos types.Position, angle float32)
	PluginMessage(channel string, data []byte)
	SystemChatMessage()*/
}

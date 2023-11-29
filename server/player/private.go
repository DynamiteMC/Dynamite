package player

import (
	"bytes"

	"github.com/aimjel/minecraft/chat"
	"github.com/aimjel/minecraft/packet"
	"github.com/aimjel/minecraft/protocol/types"
	"github.com/dynamitemc/dynamite/server/enum"
	"github.com/dynamitemc/dynamite/server/world"
	"github.com/google/uuid"
)

func (p *Player) SendPacket(pk packet.Packet) error {
	return p.Session.SendPacket(pk)
}

func (p *Player) EntityID() int32 {
	return p.entityID
}

func (p *Player) ClientSettings() clientInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.clientInfo
}

func (p *Player) SetClientSettings(pk *packet.ClientSettings) {
	p.mu.Lock()
	p.clientInfo.Locale = pk.Locale
	//don't set view distance but server controls it
	p.clientInfo.ChatMode = pk.ChatMode
	p.clientInfo.ChatColors = pk.ChatColors
	p.clientInfo.DisplayedSkinParts = pk.DisplayedSkinParts
	p.clientInfo.MainHand = pk.MainHand
	p.clientInfo.DisableTextFiltering = pk.DisableTextFiltering
	p.clientInfo.AllowServerListings = pk.AllowServerListings
	p.mu.Unlock()

	p.BroadcastMetadataInArea(&packet.SetEntityMetadata{
		DisplayedSkinParts: &pk.DisplayedSkinParts,
	})
}

func (p *Player) Properties() []types.Property {
	return p.Session.Properties()
}

func (p *Player) Dimension() *world.Dimension {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.dimension
}

func (p *Player) SetDimension(d *world.Dimension) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.dimension = d
}

func (p *Player) SetHealth(health float32) {
	if health < 0 {
		health = 0
	}

	p.Health.Set(health)
	food, saturation := p.FoodLevel.Get(), p.FoodSaturation.Get()
	p.Session.SendPacket(&packet.SetHealth{Health: health, Food: food, FoodSaturation: saturation})
	/*p.broadcastMetadataGlobal(&packet.SetEntityMetadata{
		EntityID: p.EntityID(),
		Health:   &health,
	})*/
	if health <= 0 {
		p.Kill("died :skull:")
	}
}

func (p *Player) SavedAbilities() world.Abilities {
	return p.data.Abilities
}

func (p *Player) SetGameMode(gm byte) {
	p.GameMode.Set(gm)

	p.Session.SendPacket(&packet.GameEvent{Event: enum.GameEventChangeGamemode, Value: float32(gm)})
	p.BroadcastGamemode()
}

func (p *Player) SetOperator(op bool) {
	p.Operator.Set(op)
	v := enum.EntityStatusPlayerOpPermissionLevel0
	if op {
		v = enum.EntityStatusPlayerOpPermissionLevel4
	}

	p.Session.SendPacket(&packet.EntityEvent{EntityID: p.entityID, Status: v})

}
func (p *Player) SetSessionID(id [16]byte, pk, ks []byte, expires int64) {
	p.sessionID.Set(id)
	p.publicKey.Set(bytes.Clone(pk))
	p.keySignature.Set(bytes.Clone(ks))
	p.expires.Set(expires)

	p.playerController.Range(func(_ uuid.UUID, pl *Player) bool {
		pl.Session.SendPacket(&packet.PlayerInfoUpdate{
			Actions: 0x02,
			Players: []types.PlayerInfo{
				{
					UUID:          p.Session.UUID(),
					ChatSessionID: id,
					PublicKey:     bytes.Clone(pk),
					KeySignature:  bytes.Clone(ks),
					ExpiresAt:     expires,
				},
			},
		})
		return true
	})
}

func (p *Player) SessionID() (id [16]byte, pk, ks []byte, expires int64) {
	return p.sessionID.Get(), p.publicKey.Get(), p.keySignature.Get(), p.expires.Get()
}

/*// SetSkin allows you to set the player's skin.
func (p *Player) SetSkin(url string) {
	var textures types.TexturesProperty
	b, _ := base64.StdEncoding.DecodeString(p.session.Properties()[0].Value)
	json.Unmarshal(b, &textures)
	textures.Textures.Skin.URL = url

	t, _ := json.Marshal(textures)

	d := base64.StdEncoding.EncodeToString(t)

	p.session.Properties()[0].Signature = ""
	p.session.Properties()[0].Value = d
}

// SetCape allows you to set the player's cape.
func (p *Player) SetCape(url string) {
	var textures types.TexturesProperty
	b, _ := base64.StdEncoding.DecodeString(p.session.Properties()[0].Value)
	json.Unmarshal(b, &textures)
	textures.Textures.Cape.URL = url

	t, _ := json.Marshal(textures)

	d := base64.StdEncoding.EncodeToString(t)

	p.session.Properties()[0].Signature = ""
	p.session.Properties()[0].Value = d
}*/

func (p *Player) SetDisplayName(name *chat.Message) {
	p.playerController.Range(func(_ uuid.UUID, pl *Player) bool {
		pl.Session.SendPacket(&packet.PlayerInfoUpdate{
			Actions: 0x20,
			Players: []types.PlayerInfo{
				{
					UUID:        p.Session.UUID(),
					DisplayName: name,
				},
			},
		})
		return true
	})
	p.BroadcastMetadataInArea(&packet.SetEntityMetadata{
		EntityID:            p.entityID,
		CustomName:          name,
		IsCustomNameVisible: point(name != nil),
	})
}

package player

import (
	"fmt"
	"slices"
	"strings"

	"github.com/aimjel/minecraft/chat"
	"github.com/aimjel/minecraft/packet"
	"github.com/dynamitemc/dynamite/server/enum"
	"github.com/dynamitemc/dynamite/server/lang/placeholder"

	"github.com/google/uuid"
)

func (p *Player) HandleChat(pk *packet.ChatMessageServer) {
	if !p.config.Chat.Enable {
		return
	}
	if !p.HasPermissions([]string{"server.chat"}) {
		return
	}

	prefix, suffix := p.GetPrefixSuffix()

	net := chat.NewMessage(prefix + p.Name() + suffix).WithSuggestCommandClickEvent(fmt.Sprintf("/msg %s", p.Name()))

	if !p.config.Chat.Secure {
		if !p.config.Chat.Colors || !p.HasPermissions([]string{"server.chat.colors"}) {
			// strip colors
			sp := strings.Split(pk.Message, "")
			for i, c := range sp {
				if c == "&" {
					if sp[i+1] != " " {
						sp = slices.Delete(sp, i, i+2)
					}
				}
			}
			pk.Message = strings.Join(sp, "")
		}
		if p.config.Chat.Format == "" {
			p.logger.Print(chat.Message{
				Text: point("<"),
				Extra: []chat.Message{
					net,
					{
						Text: point(">"),
					},
					{
						Text: point(" " + pk.Message),
					},
				},
			})
			p.playerController.Range(func(_ uuid.UUID, pl *Player) bool {
				if pl.ClientSettings().ChatMode != 0 {
					return true
				}
				pl.SendPacket(&packet.DisguisedChatMessage{
					Message:      chat.NewMessage(pk.Message),
					ChatTypeName: net,
				})
				return true
			})
		} else {
			globalMessage(p.logger, p.playerController, chat.NewMessage(placeholder.New(map[string]string{"message": pk.Message}, p.PlaceholderContext).Parse(p.config.Chat.Format)))
		}
	} else {
		p.logger.Print(chat.Message{
			Text: point("<"),
			Extra: []chat.Message{
				net,
				{
					Text: point(">"),
				},
				{
					Text: point(" " + pk.Message),
				},
			},
		})
		p.playerController.Range(func(_ uuid.UUID, pl *Player) bool {
			if pl.ClientSettings().ChatMode != 0 {
				return true
			}
			var pr = p.PreviousMessages()
			for _, msg := range pr {
				if !pl.IsMessageCached([256]byte(msg.Signature)) {
					msg.MessageID = -1
				}
			}
			pl.SendPacket(&packet.PlayerChatMessage{
				Sender:           p.UUID(),
				MessageSignature: pk.Signature,
				Index:            pl.Index(),
				Message:          pk.Message,
				Timestamp:        pk.Timestamp,
				Salt:             pk.Salt,
				NetworkName:      net,
				PreviousMessages: pr,
			})
			pl.CacheMessage(pk.Signature)
			return true
		})
		p.AddMessage(pk.Signature)
	}
}

func (p *Player) Whisper(pl *Player, msg string, timestamp, salt int64, sig []byte) {
	prefix, suffix := p.GetPrefixSuffix()
	prefix1, suffix1 := pl.GetPrefixSuffix()
	tgt := chat.NewMessage(prefix1 + pl.Name() + suffix1)
	p.SendPacket(&packet.PlayerChatMessage{
		Sender:  p.UUID(),
		Message: msg,
		//MessageSignature:  sig,
		Salt:              salt,
		Timestamp:         timestamp,
		ChatType:          enum.ChatTypeMsgCommandIncoming,
		NetworkName:       chat.NewMessage(prefix + p.Name() + suffix),
		NetworkTargetName: &tgt,
	})
}

func point[T any](t T) *T {
	return &t
}

package handler

import (
	"fmt"
	"strings"

	"github.com/aimjel/minecraft/chat"
	"github.com/aimjel/minecraft/packet"
	"github.com/dynamitemc/dynamite/logger"
	"github.com/dynamitemc/dynamite/logger/color"
	"github.com/dynamitemc/dynamite/server/commands"
	"github.com/dynamitemc/dynamite/server/player"
	"github.com/google/uuid"
)

func ChatCommandPacket(state *player.Player, graph *commands.Graph, log *logger.Logger, content string, timestamp, salt int64, sigs []packet.Argument) {
	log.Info(color.FromChat(chat.NewMessage(fmt.Sprintf("[%s] Player %s (%s) issued server command /%s", state.IP(), state.Session.Name(), uuid.UUID(state.Session.UUID()).String(), content))))
	args := strings.Split(content, " ")
	cmd := args[0]
	var command *commands.Command
	for _, c := range graph.Commands {
		if c == nil {
			continue
		}
		if c.Name == cmd {
			command = c
		}

		for _, a := range c.Aliases {
			if a == cmd {
				command = c
			}
		}
	}
	if command == nil || !state.HasPermissions(command.RequiredPermissions) {
		state.SendMessage(chat.NewMessage(fmt.Sprintf("§cUnknown or incomplete command, see below for error\n§n%s§r§c§o<--[HERE]", content)))
		return
	}
	ctx := commands.CommandContext{
		Command:            command,
		Arguments:          args[1:],
		Executor:           state,
		FullCommand:        content,
		ArgumentSignatures: sigs,
		Salt:               salt,
		Timestamp:          timestamp,
	}
	command.Execute(ctx)
}

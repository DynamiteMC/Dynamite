package core_commands

import (
	"strings"

	"github.com/dynamitemc/dynamite/server/commands"
	"github.com/dynamitemc/dynamite/server/lang/placeholder"
)

var ban_cmd = &commands.Command{
	Name:                "ban",
	RequiredPermissions: []string{"server.command.ban"},
	Arguments: []commands.Argument{
		commands.NewEntityArg("player", commands.EntityPlayerOnly),
		commands.NewStrArg("reason", commands.GreedyPhrase),
	},
	Execute: func(ctx commands.CommandContext) {
		if len(ctx.Arguments) == 0 {
			ctx.Incomplete()
			return
		}
		server := getServer(ctx.Executor)
		playerName := ctx.Arguments[0]
		player := server.FindPlayer(playerName)
		if player == nil {
			ctx.Error("No player was found")
			return
		}

		reason := server.Lang.Translate("disconnect.banned", nil)
		if len(ctx.Arguments) > 1 {
			reason = server.Lang.Translate("disconnect.banned.reason", placeholder.New(map[string]string{"reason": strings.Join(ctx.Arguments[1:], " ")}, player.PlaceholderContext))
		}
		server.Ban(player.Name(), player.UUID().String(), strings.Join(ctx.Arguments[1:], " "))
		player.Disconnect(reason)
	},
}

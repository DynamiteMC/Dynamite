package core_commands

import (
	"github.com/dynamitemc/dynamite/server/commands"
	"github.com/dynamitemc/dynamite/server/lang/placeholder"
)

var unban_cmd = &commands.Command{
	Name:                "unban",
	Aliases:             []string{"pardon"},
	RequiredPermissions: []string{"server.command.unban"},
	Arguments: []commands.Argument{
		commands.NewStrArg("player", commands.SingleWord),
	},
	Execute: func(ctx commands.CommandContext) {
		if len(ctx.Arguments) == 0 {
			ctx.Incomplete()
			return
		}
		playerName := ctx.Arguments[0]
		server := getServer(ctx.Executor)

		server.Unban(playerName)
		ctx.Reply(server.Lang.Translate("commands.pardon.success", placeholder.New(map[string]string{
			"player": playerName,
		}, server.PlaceholderContext)))
	},
}

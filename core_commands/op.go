package core_commands

import (
	"github.com/dynamitemc/dynamite/server/commands"
)

var op_cmd = &commands.Command{
	Name:                "op",
	RequiredPermissions: []string{"server.command.op"},
	Arguments: []commands.Argument{
		commands.NewEntityArg("player", commands.EntityPlayerOnly),
	},
	Execute: func(ctx commands.CommandContext) {
		if len(ctx.Arguments) == 0 {
			ctx.Incomplete()
			return
		}
		playerName := ctx.Arguments[0]
		server := getServer(ctx.Executor)
		player := server.FindPlayer(playerName)
		if player == nil {
			ctx.Error("No player was found")
			return
		}
		server.MakeOperator(player)
		player.SendCommands(server.GetCommandGraph())
		ctx.Reply(server.Lang.Translate("commands.op.success", player.PlaceholderContext))
	},
}

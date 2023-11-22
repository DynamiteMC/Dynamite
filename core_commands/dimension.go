package core_commands

import (
	"github.com/dynamitemc/dynamite/server"
	"github.com/dynamitemc/dynamite/server/commands"
	"github.com/dynamitemc/dynamite/server/lang/placeholder"
	"github.com/dynamitemc/dynamite/server/player"
)

var dimension_cmd = &commands.Command{
	Name:                "dimension",
	RequiredPermissions: []string{"server.command.dimension"},
	Arguments: []commands.Argument{
		commands.NewDimensionArg("dimension"),
	},
	Execute: func(ctx commands.CommandContext) {
		if len(ctx.Arguments) == 0 {
			ctx.Incomplete()
			return
		}
		if p, ok := ctx.Executor.(*player.Player); ok {
			p.Respawn(p.Server.(*server.Server).World.GetDimension(ctx.Arguments[0]))
			ctx.Reply(p.Server.(*server.Server).Lang.Translate("commands.dimension.success", placeholder.New(map[string]string{
				"dimension": ctx.Arguments[0],
			}, p.PlaceholderContext)))
		} else {
			ctx.Error("This command can only be used by players")
			return
		}
	},
}

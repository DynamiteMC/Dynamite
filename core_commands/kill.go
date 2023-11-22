package core_commands

import (
	"github.com/aimjel/minecraft/chat"
	"github.com/dynamitemc/dynamite/server"
	"github.com/dynamitemc/dynamite/server/commands"
	"github.com/dynamitemc/dynamite/server/player"
)

var kill_cmd = &commands.Command{
	Name:                "kill",
	RequiredPermissions: []string{"server.command.kill"},
	Arguments: []commands.Argument{
		commands.NewEntityArg("player", commands.EntityPlayerOnly),
	},
	Execute: func(ctx commands.CommandContext) {
		var pl *player.Player
		if len(ctx.Arguments) == 0 {
			if p, ok := ctx.Executor.(*player.Player); !ok {
				ctx.Incomplete()
				return
			} else {
				pl = p
			}
		} else {
			p := getServer(ctx.Executor).FindPlayer(ctx.Arguments[0])
			if p == nil {
				ctx.Error("No player was found")
				return
			}
			pl = p
		}
		name := pl.Name()
		pl.Kill(name + " was killed")
		ctx.Reply(pl.Server.(*server.Server).Lang.Translate("commands.kill.success.single", pl.PlaceholderContext))
		pl.Server.(*server.Server).GlobalMessage(chat.NewMessage(name + " was killed"))
	},
}

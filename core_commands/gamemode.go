package core_commands

import (
	"fmt"
	"strings"

	"github.com/dynamitemc/dynamite/server"
	"github.com/dynamitemc/dynamite/server/commands"
	"github.com/dynamitemc/dynamite/server/lang/placeholder"
	"github.com/dynamitemc/dynamite/server/player"
)

func pascalify(str string) (res string) {
	sp := strings.Split(str, " ")
	for _, w := range sp {
		spp := strings.Split(w, "")
		res += strings.ToUpper(spp[0]) + strings.ToLower(strings.Join(spp[1:], ""))
	}
	return
}

var gamemode_cmd = &commands.Command{
	Name:                "gamemode",
	RequiredPermissions: []string{"server.command.gamemode"},
	Arguments: []commands.Argument{
		commands.NewGamemodeArg("mode"),
		commands.NewEntityArg("player", commands.EntityPlayerOnly),
	},
	Execute: func(ctx commands.CommandContext) {
		if len(ctx.Arguments) == 0 {
			ctx.Incomplete()
			return
		}
		gm := player.Gamemode(ctx.Arguments[0])
		if gm == -1 {
			ctx.ErrorHere(fmt.Sprintf("Unknown game mode: %s", ctx.Arguments[0]))
			return
		}
		var pl *player.Player
		if len(ctx.Arguments) == 1 {
			if p, ok := ctx.Executor.(*player.Player); !ok {
				ctx.Incomplete()
				return
			} else {
				pl = p
			}
		} else {
			p := getServer(ctx.Executor).FindPlayer(ctx.Arguments[1])
			if p == nil {
				ctx.Error("No player was found")
				return
			}
			pl = p
		}
		if int(pl.GameMode()) == gm {
			return
		}
		pl.SetGameMode(byte(gm))
		ph := placeholder.New(map[string]string{"gamemode": pascalify(ctx.Arguments[0])}, pl.PlaceholderContext)
		msg := pl.Server.(*server.Server).Lang.Translate("commands.gamemode.success.other", ph)
		if exe, ok := ctx.Executor.(*player.Player); ok && pl.UUID() == exe.UUID() {
			msg = pl.Server.(*server.Server).Lang.Translate("commands.gamemode.success.self", ph)
		}
		ctx.Reply(msg)
	},
}

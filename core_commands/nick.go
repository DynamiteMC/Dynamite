package core_commands

import (
	"strings"

	"github.com/aimjel/minecraft/chat"
	"github.com/dynamitemc/dynamite/server/commands"
	"github.com/dynamitemc/dynamite/server/player"
	"github.com/google/uuid"
)

var nick_cmd = &commands.Command{
	Name:                "nick",
	RequiredPermissions: []string{"server.command.nick"},
	Arguments: []commands.Argument{
		commands.NewEntityArg("player", commands.EntityPlayerOnly),
		commands.NewStrArg("nickname", commands.QuotablePhrase),
	},
	Execute: func(ctx commands.CommandContext) {
		srv := getServer(ctx.Executor)
		var name string
		var p *player.Player
		if len(ctx.Arguments) == 0 {
			if pl, ok := ctx.Executor.(*player.Player); ok {
				p = pl
			} else {
				ctx.Incomplete()
				return
			}
		} else {
			p = srv.Players.Find(func(_ uuid.UUID, pl *player.Player) bool {
				return pl.Session.Name() == ctx.Arguments[0]
			})
		}
		if p == nil {
			if pl, ok := ctx.Executor.(*player.Player); ok {
				p = pl
			} else {
				ctx.Incomplete()
				return
			}
		} else {
			if len(ctx.Arguments) >= 1 {
				ctx.Arguments = ctx.Arguments[1:]
			}
		}
		if len(ctx.Arguments) >= 1 {
			name = strings.TrimSuffix(strings.TrimPrefix(strings.Join(ctx.Arguments, " "), `"`), `"`)
		}
		if name == "" {
			p.SetDisplayName(nil)
		} else {
			m := chat.NewMessage(name)
			p.SetDisplayName(&m)
		}
		ctx.Reply(chat.NewMessage("Set nickname."))
	},
}

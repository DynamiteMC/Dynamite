package core_commands

import (
	"fmt"

	"github.com/aimjel/minecraft/chat"
	"github.com/aimjel/minecraft/packet"
	"github.com/dynamitemc/dynamite/server/commands"
	"github.com/dynamitemc/dynamite/server/player"
	"github.com/google/uuid"
)

var list_cmd = &commands.Command{
	Name:                "list",
	RequiredPermissions: []string{"server.command.list"},
	Arguments: []commands.Argument{
		commands.NewStrArg("uuids", commands.SingleWord).SetSuggest(func(ctx commands.SuggestionsContext) {
			ctx.Return([]packet.SuggestionMatch{
				{Match: "uuids"},
			})
		}),
	},
	Execute: func(ctx commands.CommandContext) {
		srv := getServer(ctx.Executor)
		l := srv.Players.Count()
		if l == 0 {
			ctx.Reply(chat.NewMessage("No players online"))
			return
		}
		msg := fmt.Sprintf("There are %d of a max of %d players online: ", l, srv.Config.MaxPlayers)
		var index int
		srv.Players.Range(func(_ uuid.UUID, p *player.Player) bool {
			msg += p.Session.Name()
			if len(ctx.Arguments) == 1 && ctx.Arguments[0] == "uuids" {
				msg += fmt.Sprintf(" (%s)", uuid.UUID(p.Session.UUID()))
			}
			if index != l-1 {
				msg += ", "
			}
			index++
			return true
		})
		ctx.Reply(chat.NewMessage(msg))
	},
}

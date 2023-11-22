package core_commands

import (
	"fmt"

	"github.com/aimjel/minecraft/chat"
	"github.com/dynamitemc/dynamite/server/commands"
	"github.com/dynamitemc/dynamite/server/lang/placeholder"
)

var seed_cmd = &commands.Command{
	Name:                "seed",
	RequiredPermissions: []string{"server.command.seed"},
	Execute: func(ctx commands.CommandContext) {
		server := getServer(ctx.Executor)
		seed := server.World.Seed()
		ctx.Reply(server.Lang.Translate("commands.seed.success", placeholder.New(map[string]string{"seed": fmt.Sprint(seed)}, server.PlaceholderContext)).
			WithCopyToClipboardClickEvent(fmt.Sprint(seed)).
			WithShowTextHoverEvent(chat.NewMessage("Click to Copy to clipboard")))
	},
}

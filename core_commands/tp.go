package core_commands

import (
	"fmt"
	"strconv"

	"github.com/dynamitemc/dynamite/server/commands"
	"github.com/dynamitemc/dynamite/server/lang/placeholder"
	"github.com/dynamitemc/dynamite/server/player"
)

var tp_cmd = &commands.Command{
	Name:                "tp",
	RequiredPermissions: []string{"server.command.op"},
	Aliases:             []string{"teleport"},
	Arguments: []commands.Argument{
		commands.NewEntityArg("targets", commands.EntityPlayerOnly),
		commands.NewEntityArg("destination", commands.EntityPlayerOnly).SetAlternative(commands.NewVector3Arg("location")),
	},
	Execute: func(ctx commands.CommandContext) {
		srv := getServer(ctx.Executor)
		switch len(ctx.Arguments) {
		case 1:
			{
				if exe, ok := ctx.Executor.(*player.Player); !ok {
					ctx.Incomplete()
					return
				} else {
					player := srv.FindPlayer(ctx.Arguments[0])
					x, y, z := player.Position.X(), player.Position.Y(), player.Position.Z()
					yaw, pitch := exe.Position.Yaw(), exe.Position.Pitch()
					exe.Teleport(x, y, z, yaw, pitch)
					ep, es := exe.GetPrefixSuffix()
					pp, ps := player.GetPrefixSuffix()
					ctx.Reply(srv.Lang.Translate("commands.teleport.success.entity.single", placeholder.New(map[string]string{
						"player":         exe.Name(),
						"player_prefix":  ep,
						"player_suffx":   es,
						"player1":        player.Name(),
						"player1_prefix": pp,
						"player1_suffx":  ps,
					})))
				}
			}
		case 2:
			{
				// Teleport player to player
				player1 := srv.FindPlayer(ctx.Arguments[0])
				player2 := srv.FindPlayer(ctx.Arguments[1])
				x, y, z := player2.Position.X(), player2.Position.Y(), player2.Position.Z()
				yaw, pitch := player1.Position.Yaw(), player1.Position.Yaw()
				player1.Teleport(x, y, z, yaw, pitch)

				ep, es := player1.GetPrefixSuffix()
				pp, ps := player2.GetPrefixSuffix()
				ctx.Reply(srv.Lang.Translate("commands.teleport.success.entity.single", placeholder.New(map[string]string{
					"player":         player1.Name(),
					"player_prefix":  ep,
					"player_suffx":   es,
					"player1":        player2.Name(),
					"player1_prefix": pp,
					"player1_suffx":  ps,
				})))
			}
		case 3:
			{
				// Teleport executor to coordinates
				if exe, ok := ctx.Executor.(*player.Player); !ok {
					ctx.Incomplete()
				} else {
					x, err := strconv.ParseFloat(ctx.Arguments[0], 64)
					if err != nil {
						ctx.Error("Invalid x position")
						return
					}
					y, err := strconv.ParseFloat(ctx.Arguments[1], 64)
					if err != nil {
						ctx.Error("Invalid y position")
						return
					}
					z, err := strconv.ParseFloat(ctx.Arguments[2], 64)
					if err != nil {
						ctx.Error("Invalid x position")
						return
					}
					yaw, pitch := exe.Position.Yaw(), exe.Position.Pitch()

					exe.Teleport(x, y, z, yaw, pitch)

					prefix, suffix := exe.GetPrefixSuffix()
					ctx.Reply(srv.Lang.Translate("commands.teleport.success.location.single", placeholder.New(
						map[string]string{
							"player":        exe.Name(),
							"player_prefix": prefix,
							"player_suffx":  suffix,
							"x":             fmt.Sprint(x),
							"y":             fmt.Sprint(y),
							"z":             fmt.Sprint(z),
						})))
				}
			}
		case 4:
			{
				// teleport player to coordinates
				player := srv.FindPlayer(ctx.Arguments[0])
				x, err := strconv.ParseFloat(ctx.Arguments[1], 64)
				if err != nil {
					ctx.Error("Invalid x position")
					return
				}
				y, err := strconv.ParseFloat(ctx.Arguments[2], 64)
				if err != nil {
					ctx.Error("Invalid y position")
					return
				}
				z, err := strconv.ParseFloat(ctx.Arguments[3], 64)
				if err != nil {
					ctx.Error("Invalid x position")
					return
				}

				yaw, pitch := player.Position.Yaw(), player.Position.Pitch()
				player.Teleport(x, y, z, yaw, pitch)

				prefix, suffix := player.GetPrefixSuffix()
				ctx.Reply(srv.Lang.Translate("commands.teleport.success.location.single", placeholder.New(
					map[string]string{
						"player":        player.Name(),
						"player_prefix": prefix,
						"player_suffx":  suffix,
						"x":             fmt.Sprint(x),
						"y":             fmt.Sprint(y),
						"z":             fmt.Sprint(z),
					})))
			}
		default:
			ctx.Incomplete()
		}
	},
}

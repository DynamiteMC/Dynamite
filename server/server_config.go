package server

import (
	"errors"
	"os"
	"sync"

	"github.com/aimjel/minecraft"
	"github.com/dynamitemc/dynamite/config"
	"github.com/dynamitemc/dynamite/logger"
	"github.com/dynamitemc/dynamite/server/commands"
	"github.com/dynamitemc/dynamite/server/plugins"
	"github.com/dynamitemc/dynamite/server/world"
	"github.com/dynamitemc/dynamite/util"
)

type ServerConfig config.ServerConfig

func (cfg *ServerConfig) Listen(address string, logger logger.Logger, commandGraph *commands.Graph) (*Server, error) {
	lnCfg := minecraft.ListenConfig{
		Status: minecraft.NewStatus(minecraft.Version{
			Text:     "DynamiteMC 1.20.1",
			Protocol: 763,
		}, cfg.MaxPlayers, cfg.MOTD),
		OnlineMode:           cfg.Online,
		CompressionThreshold: 256,
		Messages: &minecraft.Messages{
			OnlineMode:     cfg.Messages.OnlineMode,
			ProtocolTooNew: cfg.Messages.ProtocolNew,
			ProtocolTooOld: cfg.Messages.ProtocolOld,
		},
	}
	//web.SetMaxPlayers(cfg.MaxPlayers)

	ln, err := lnCfg.Listen(address)
	if err != nil {
		return nil, err
	}
	w, err := world.OpenWorld(util.GetArg("worldpath", "world"))
	if err != nil {
		logger.Error("Failed to load world: %s", err)
		os.Exit(1)
	}
	c := config.ServerConfig(*cfg)
	srv := &Server{
		Config:       &c,
		listener:     ln,
		Logger:       logger,
		world:        w,
		mu:           &sync.RWMutex{},
		Players:      make(map[string]*PlayerController),
		CommandGraph: commandGraph,
		Plugins:      make(map[string]*plugins.Plugin),
	}

	var files = []string{"whitelist.json", "banned_players.json", "ops.json", "banned_ips.json"}
	var addresses = []*[]user{&srv.WhitelistedPlayers, &srv.BannedPlayers, &srv.Operators, &srv.BannedIPs}
	for i, file := range files {
		u, err := loadUsers(file)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}

		*addresses[i] = u
	}

	logger.Debug("Loaded player info")
	return srv, nil
}

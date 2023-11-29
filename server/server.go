package server

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dynamitemc/dynamite/server/block"
	"github.com/dynamitemc/dynamite/server/lang/placeholder"

	"github.com/dynamitemc/dynamite/server/block/pos"
	"github.com/dynamitemc/dynamite/server/config"
	"github.com/dynamitemc/dynamite/server/controller"
	"github.com/dynamitemc/dynamite/server/entity"
	"github.com/dynamitemc/dynamite/server/enum"
	"github.com/dynamitemc/dynamite/server/handler"
	"github.com/dynamitemc/dynamite/server/lang"
	"github.com/dynamitemc/dynamite/server/permission"
	"github.com/dynamitemc/dynamite/server/world/chunk"
	"github.com/dynamitemc/dynamite/server/world/generator/overworld"
	"github.com/google/uuid"
	"golang.org/x/term"

	"github.com/aimjel/minecraft/chat"
	"github.com/aimjel/minecraft/protocol/types"
	"github.com/pelletier/go-toml/v2"

	"github.com/aimjel/minecraft"
	"github.com/aimjel/minecraft/packet"

	//"github.com/dynamitemc/dynamite/web"
	"github.com/dynamitemc/dynamite/logger"
	"github.com/dynamitemc/dynamite/server/commands"
	"github.com/dynamitemc/dynamite/server/player"
	"github.com/dynamitemc/dynamite/server/world"
	"github.com/dynamitemc/dynamite/web"
)

const Version = "1.20.1"

func New(cfg *config.Config, address string, logger *logger.Logger, commandGraph *commands.Graph) (*Server, error) {
	lnCfg := minecraft.ListenConfig{
		Status: minecraft.NewStatus(minecraft.Version{
			Text:     "DynamiteMC 1.20.1",
			Protocol: 763,
		}, cfg.MaxPlayers, cfg.MOTD, true, true),
		OnlineMode:           cfg.Online,
		CompressionThreshold: int32(cfg.CompressionThreshold),
	}

	if cfg.Chat.Secure && !cfg.Online {
		logger.Warn("Secure chat doesn't work on offline mode")
		cfg.Chat.Secure = false
	}
	if cfg.Chat.Secure && cfg.Chat.Format != "" {
		logger.Warn("Secure chat overrides the chat format")
	}
	if cfg.TPS < 20 {
		logger.Warn("TPS must be at least 20")
		cfg.TPS = 20
	}
	if cfg.ResourcePack.Enable && cfg.ResourcePack.URL == "" {
		logger.Warn("Resource pack is enabled but no url is provided")
		cfg.ResourcePack.Enable = false
	}

	//web.SetMaxPlayers(cfg.MaxPlayers)

	ln, err := lnCfg.Listen(address)
	if err != nil {
		return nil, err
	}
	w, err := world.OpenWorld("world", cfg.Superflat)
	if err != nil {
		world.CreateWorld(cfg.Hardcore)
		logger.Error("Failed to load world: %s", err)
		os.Exit(1)
	}
	w.Gamemode = byte(player.Gamemode(cfg.Gamemode))
	srv := &Server{
		Config:       cfg,
		listener:     ln,
		Logger:       logger,
		World:        w,
		Players:      controller.New[uuid.UUID, *player.Player](),
		Entities:     controller.New[int32, entity.Entity](),
		commandGraph: commandGraph,
		Lang:         lang.New("lang.json"),
		PlaceholderContext: placeholder.New(map[string]string{
			"server_host":        cfg.ServerIP,
			"server_port":        fmt.Sprint(cfg.ServerPort),
			"server_max_players": fmt.Sprint(cfg.MaxPlayers),
		}),
	}
	w.Overworld().SetGenerator(&overworld.DefaultGenerator{})

	srv.loadFiles()
	logger.Debug("Loaded player info")

	return srv, nil
}

var idCounter atomic.Int32

type Server struct {
	Config       *config.Config
	Logger       *logger.Logger
	commandGraph *commands.Graph

	Lang *lang.Lang

	PlaceholderContext *placeholder.PlaceholderContext

	WhitelistedPlayers,
	Operators,
	BannedPlayers,
	BannedIPs []user

	listener *minecraft.Listener

	Players  *controller.Controller[uuid.UUID, *player.Player]
	Entities *controller.Controller[int32, entity.Entity]

	World *world.World

	mu sync.RWMutex
}

func (srv *Server) Start() error {
	for {
		conn, err := srv.listener.Accept()
		if err != nil {
			return err
		}
		go srv.handleNewConn(conn)
	}
}

func (srv *Server) NewID() int32 {
	return idCounter.Add(1)
}

func (srv *Server) handleNewConn(conn *minecraft.Conn) {
	if srv.ValidateConn(conn) {
		conn.Close(nil)
		return
	}

	x := conn.UUID()
	uuid := uuid.UUID(x[:])

	data := srv.World.GetPlayerData(uuid.String())

	plyr := player.New(
		srv.Players,
		srv.Entities,
		srv,
		srv.Config,
		srv.Lang,
		srv.PlaceholderContext,
		srv.Logger,
		idCounter.Add(1),
		conn,
		data,
		srv.World.GetDimension(data.Dimension),
		int8(srv.Config.ViewDistance),
		srv.NewID,
	)

	for _, op := range srv.Operators {
		if op.UUID == uuid.String() {
			plyr.SetOperator(true)
		}
	}

	plyr.SendCommands(srv.commandGraph)

	//cntrl.SendPacket(&packet.SetTablistHeaderFooter{
	//	Header: strings.Join(srv.Config.Tablist.Header, "\n"),
	//	Footer: strings.Join(srv.Config.Tablist.Footer, "\n"),
	//})

	srv.addPlayer(plyr)
	plyr.Login(plyr.Dimension())

	plyr.IntitializeData()

	if err := srv.handlePackets(plyr); err != nil {
		srv.Logger.Info("[%s] Player %s (%s) has left the server", conn.RemoteAddr().String(), conn.Name(), uuid)

		srv.GlobalMessage(srv.Lang.Translate("player.leave", plyr.PlaceholderContext))
		srv.Players.Delete(plyr.Session.UUID())
		plyr.Save()
		srv.playerlistRemove(conn.UUID())
		plyr.Despawn()

		//todo consider moving logic of removing player to a separate function
	}
}

func (srv *Server) addPlayer(p *player.Player) {
	p.SendPacket(&packet.ServerData{
		MOTD:               chat.NewMessage(srv.Config.MOTD),
		EnforcesSecureChat: true,
	})
	newPlayer := types.PlayerInfo{
		UUID:       p.Session.UUID(),
		Name:       p.Session.Name(),
		Properties: p.Properties(),
		Listed:     true,
	}

	srv.Players.Set(p.Session.UUID(), p)
	players := make([]types.PlayerInfo, 0, srv.Players.Count())

	srv.Players.Range(func(_ uuid.UUID, pl *player.Player) bool {
		id, pk, ks, ex := pl.SessionID()
		players = append(players, types.PlayerInfo{
			UUID:          pl.Session.UUID(),
			Name:          pl.Session.Name(),
			Properties:    pl.Properties(),
			Listed:        true,
			PublicKey:     bytes.Clone(pk),
			KeySignature:  bytes.Clone(ks),
			ChatSessionID: id,
			ExpiresAt:     ex,
		})
		if pl.Session.UUID() != p.Session.UUID() {
			pl.SendPacket(&packet.PlayerInfoUpdate{
				Actions: 0x01 | 0x08,
				Players: []types.PlayerInfo{newPlayer},
			})
		}
		return true
	})

	//updates the new session's player list
	var actions byte = 0x01 | 0x08
	if srv.Config.Chat.Secure {
		actions |= 0x02
	}
	p.SendPacket(&packet.PlayerInfoUpdate{
		Actions: actions,
		Players: players,
	})

	if srv.Config.Web.Enable {
		web.AddPlayer(p.Session.Name(), uuid.UUID(p.Session.UUID()).String())
	}

	srv.GlobalMessage(srv.Lang.Translate("player.join", p.PlaceholderContext))
}

func (srv *Server) handlePackets(p *player.Player) error {
	ticker := time.NewTicker(25 * time.Second)
	for {
		select {
		case <-ticker.C:
			p.Keepalive()
		default:
		}

		packt, err := p.Session.ReadPacket()
		if err != nil {
			return err
		}

		switch pk := packt.(type) {
		case *packet.PlayerCommandServer:
			handler.PlayerCommand(p, pk.ActionID)
		case *packet.ChatMessageServer:
			handler.ChatMessagePacket(p, pk)
		case *packet.ChatCommandServer:
			handler.ChatCommandPacket(p, srv.commandGraph, srv.Logger, pk.Command, pk.Timestamp, pk.Salt, pk.ArgumentSignatures)
		case *packet.ClientSettings:
			handler.ClientSettings(p, pk)
		case *packet.PlayerPosition, *packet.PlayerPositionRotation, *packet.PlayerRotation:
			handler.PlayerMovement(p, pk)
		case *packet.PlayerActionServer:
			handler.PlayerAction(p, pk)
		case *packet.InteractServer:
			handler.Interact(p, pk)
		case *packet.SwingArmServer:
			handler.SwingArm(p, pk.Hand)
		case *packet.CommandSuggestionsRequest:
			handler.CommandSuggestionsRequest(pk.TransactionId, pk.Text, srv.commandGraph, p)
		case *packet.ClientCommandServer:
			handler.ClientCommand(p, pk.ActionID)
		case *packet.PlayerAbilitiesServer:
			handler.PlayerAbilities(p, pk.Flags)
		case *packet.PlayerSessionServer:
			p.SetSessionID(pk.SessionID, pk.PublicKey, pk.KeySignature, pk.ExpiresAt)
		case *packet.SetHeldItemServer:
			handler.SetHeldItem(p, pk.Slot)
		case *packet.SetCreativeModeSlot:
			handler.SetCreativeModeSlot(p, pk.Slot, pk.ClickedItem)
		case *packet.TeleportToEntityServer:
			handler.TeleportToEntity(p, pk.Player)
		case *packet.ClickContainer:
			handler.ClickContainer(p, pk)
		case *packet.UseItemOnServer:
			handler.UseItemOn(p, pk, srv.SetBlock)
		}
	}
}

func (srv *Server) GetCommandGraph() *commands.Graph {
	return srv.commandGraph
}

func (srv *Server) Reload() error {
	srv.loadFiles()

	config.LoadConfig("config.toml", srv.Config)

	permission.Clear()

	srv.Players.Range(func(_ uuid.UUID, p *player.Player) bool {
		if srv.Config.Whitelist.Enforce && srv.Config.Whitelist.Enable && !srv.IsWhitelisted(p.Session.UUID()) {
			p.Disconnect(srv.Lang.Translate("disconnect.not_whitelisted", nil))
			return true
		}

		p.SetOperator(srv.IsOperator(p.Session.UUID()))
		p.SendCommands(srv.commandGraph)
		return true
	})
	return nil
}

func (srv *Server) FindPlayer(username string) *player.Player {
	_, p := srv.Players.Range(func(_ uuid.UUID, p *player.Player) bool {
		return !strings.EqualFold(p.Session.Name(), username)
	})

	return p
}

func (srv *Server) FindPlayerByID(id int32) *player.Player {
	_, p := srv.Players.Range(func(_ uuid.UUID, p *player.Player) bool {
		return p.EntityID() != id
	})

	return p
}

func (srv *Server) Close() {
	srv.Logger.Info("Stopping server")

	srv.Logger.Info("Saving player data")
	var files = []string{"whitelist.json", "banned_players.json", "ops.json", "banned_ips.json"}
	var lists = [][]user{srv.WhitelistedPlayers, srv.BannedPlayers, srv.Operators, srv.BannedIPs}
	for i, file := range files {
		WritePlayerList(file, lists[i])
	}

	permission.Save()

	srv.Players.Range(func(_ uuid.UUID, p *player.Player) bool {
		p.Disconnect(srv.Lang.Translate("disconnect.server_shutdown", srv.PlaceholderContext))
		p.Save()
		return true
	})

	srv.Logger.Info("Saving world")
	srv.World.Save()

	f, _ := os.OpenFile("config.toml", os.O_RDWR|os.O_CREATE, 0666)
	toml.NewEncoder(f).Encode(srv.Config)
	term.Restore(int(os.Stdin.Fd()), OldState)
	os.Exit(0)
}

func (srv *Server) loadFiles() {
	var files = []string{"whitelist.json", "banned_players.json", "ops.json", "banned_ips.json"}
	var addresses = []*[]user{&srv.WhitelistedPlayers, &srv.BannedPlayers, &srv.Operators, &srv.BannedIPs}
	for i, file := range files {
		u, err := loadUsers(file)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			srv.Logger.Warn("%v loading %v", err, file)
		}

		*addresses[i] = u
	}
}

func (srv *Server) ConsoleCommand(txt string) {
	fmt.Println(txt)
	fmt.Print("\r")
	content := strings.TrimSpace(txt)
	args := strings.Split(content, " ")

	command := srv.GetCommandGraph().FindCommand(args[0])
	if command == nil {
		srv.Logger.Print(chat.NewMessage(fmt.Sprintf("&cUnknown or incomplete command, see below for error\n&n%s&r&c&o<--[HERE]", args[0])))
		return
	}
	command.Execute(commands.CommandContext{
		Command:     command,
		Arguments:   args[1:],
		Executor:    srv,
		FullCommand: content,
	})
}

func (srv *Server) SetBlock(d *world.Dimension, x, y, z int64, b chunk.Block, typ world.SetBlockHandling) {
	if typ > 2 {
		typ = world.SetBlockReplace
	}

	_, isAir := b.(block.Air)

	if typ == world.SetBlockKeep && isAir {
		return
	}
	//d.SetBlock(x, y, z, b)
	cx, cz := int32(x/16), int32(z/16)
	pos := pos.BlockPosition{x, y, z}.Data()
	bid, _ := chunk.GetBlockId(b)

	bu := &packet.BlockUpdate{
		Location: pos,
		BlockID:  int32(bid),
	}

	we := &packet.WorldEvent{Event: enum.WorldEventBlockBreak, Location: pos, Data: int32(bid)}
	srv.Players.Range(func(_ uuid.UUID, pl *player.Player) bool {
		if !pl.IsChunkLoaded(cx, cz) {
			return true
		}
		if typ == world.SetBlockDestroy && isAir {
			pl.SendPacket(we)
		}
		pl.SendPacket(bu)
		return true
	})
}

var OldState *term.State

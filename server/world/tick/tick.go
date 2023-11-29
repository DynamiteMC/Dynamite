package tick

import (
	"time"

	"github.com/MichaelTJones/pcg"
	"github.com/aimjel/minecraft/packet"
	"github.com/dynamitemc/dynamite/server"
	"github.com/dynamitemc/dynamite/server/block"
	"github.com/dynamitemc/dynamite/server/block/pos"
	"github.com/dynamitemc/dynamite/server/entity"
	"github.com/dynamitemc/dynamite/server/player"
	"github.com/dynamitemc/dynamite/server/world"
	"github.com/google/uuid"
)

var pcg32 = pcg.NewPCG32()

type Ticker struct {
	ticker      *time.Ticker
	server      *server.Server
	started     bool
	currentTick uint
	paused      bool
}

func New(srv *server.Server, tps int64) *Ticker {
	return &Ticker{ticker: time.NewTicker(time.Second / time.Duration(tps)), server: srv}
}

func (t *Ticker) Start() {
	if t.started {
		return
	}
	t.started = true
	go t.tickLoop()
}

func (t *Ticker) Restart(tps int64) {
	t.ticker = time.NewTicker(time.Second / time.Duration(tps))
}

func (t *Ticker) Pause() {
	t.paused = true
}

func (t *Ticker) Resume() {
	t.paused = false
}

func (t *Ticker) tickLoop() {
	for range t.ticker.C {
		if t.paused {
			continue
		}
		t.tick(t.currentTick)
		t.currentTick++
	}
}

func (t *Ticker) tick(tick uint) {
	t.server.Entities.RangeNoLock(func(_ int32, e entity.Entity) bool {
		e.Tick(t.server, tick)
		return true
	})

	//TODO: tick blocks again
	if tick%8 == 0 {
		//blockTick(t.server, t.server.World.Overworld(), tick, 3)
	}

	worldAge, dayTime := t.server.World.IncrementTime()
	t.server.Players.RangeNoLock(func(_ uuid.UUID, pl *player.Player) bool {
		//TODO: send chunks again.
		if tick%8 == 0 {
			//pl.SendChunks(pl.Dimension())
			//pl.UnloadChunks()
		}

		pl.SendPacket(&packet.UpdateTime{
			WorldAge:  worldAge,
			TimeOfDay: dayTime,
		})
		return true
	})
}

func blockTick(srv *server.Server, d *world.Dimension, tick uint, rts int) {

	for h, c := range d.Chunks() {
		cx, cz := h.Position()
		for cy, s := range c.Sections() {
			randomTickedBlocks := 0
			for x := 0; x < 16; x++ {
				for y := 0; y < 16; y++ {
					for z := 0; z < 16; z++ {
						b := s.GetBlockAt(x, y, z)
						x1, y1, z1 := int64((int(cx)*16)+x), int64(((cy-4)*16)+y), int64((int(cz)*16)+z)
						pos := pos.BlockPosition{x1, y1, z1}

						if bl, ok := b.(block.Ticker); ok {
							srv.SetBlock(d, x1, y1, z1, bl.Tick(pos, d, tick), world.SetBlockReplace)
						}

						if randomTickedBlocks < rts && randBool() {
							if bl, ok := b.(block.RandomTicker); ok {
								srv.SetBlock(d, x1, y1, z1, bl.RandomTick(pos, d, tick), world.SetBlockReplace)
							}
							randomTickedBlocks++
						}
					}
				}
			}
		}
	}
}

func randBool() bool {
	return pcg32.Random()&0x01 == 0
}

package network

import (
	_ "embed"
	"fmt"
	"github.com/aimjel/nitrate/server/world/entity"
	"math"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aimjel/minecraft"
	"github.com/aimjel/minecraft/packet"
	"github.com/aimjel/minecraft/protocol"
)

var (
	idCounter atomic.Int32

	// entities holds the entity instances on the server(s)
	// mapped by entity id
	entities = map[int32]entity.Entity{}

	// mu protects the entities map
	mu sync.RWMutex

	//go:embed registry.nbt
	registry []byte
)

func addEntity(id int32, e entity.Entity) {
	mu.Lock()
	defer mu.Unlock()
	entities[id] = e
}

type packetHandler func(c *Session, pk packet.Packet)

// Session handles incoming and outgoing packets which controls the player state.
type Session struct {
	conn *minecraft.Conn

	//world *world.World

	state Player

	// The entity ID of the Session
	eid int32

	ViewDistance       int32
	SimulationDistance int32

	b *Broadcast

	spawnedEntities map[int32]struct{}

	handlers map[int32]packetHandler

	mu sync.RWMutex
}

func NewSession(c *minecraft.Conn, p Player, b *Broadcast) *Session {
	return &Session{
		conn:            c,
		state:           p,
		eid:             idCounter.Add(1),
		b:               b,
		spawnedEntities: make(map[int32]struct{}),
		handlers: map[int32]packetHandler{
			0x10: HandlePlayerLeftRightClick,

			0x14: HandlePlayerMovement,
			0x15: HandlePlayerMovement,
			0x16: HandlePlayerMovement,

			0x1E: HandlePlayerCommand,
		},
	}
}

func (s *Session) HandlePackets() error {
	ticker := time.NewTicker(time.Second * 25)
	addEntity(s.eid, s.state)
	for {
		select {
		case <-ticker.C:
			_ = s.conn.SendPacket(&packet.KeepAliveClient{PayloadID: time.Now().UnixMilli()})
		default:
		}

		pk, err := s.conn.ReadPacket()
		if err != nil {
			return err
		}

		if handler, ok := s.handlers[pk.ID()]; ok {
			handler(s, pk)
		} else {
			val := reflect.ValueOf(pk)
			var name string
			switch val.Kind() {
			case reflect.Struct:
				name = val.Type().Name()

			case reflect.Pointer:
				name = val.Elem().Type().Name()
			}
			fmt.Printf("no handle for 0x%x %s\n", pk.ID(), name)
		}
	}
}

func (s *Session) LoginPlay() error {
	dims := []string{s.state.Dimension().Type()}

	s.conn.SendPacket(&packet.JoinGame{
		EntityID:            s.eid,
		IsHardcore:          false, //todo
		GameMode:            uint8(s.state.GameMode()),
		PreviousGameMode:    -1,
		DimensionNames:      dims,
		Registry:            registry,
		DimensionType:       dims[0],
		DimensionName:       "nitrate:secret",
		HashedSeed:          0,
		MaxPlayers:          0,
		ViewDistance:        s.ViewDistance,
		SimulationDistance:  s.ViewDistance,
		ReducedDebugInfo:    false,
		EnableRespawnScreen: false, //todo add game rules
		IsDebug:             false,
		IsFlat:              true, //todo
		DeathDimensionName:  "",   //todo
		DeathLocation:       0,    //todo
		PartialCooldown:     0,    //todo
	})

	s.Teleport(0, 5, 0)

	s.sendSpawnChunks()

	//closes loading screen
	return s.conn.SendPacket(packet.SetDefaultSpawnPosition{})
}

func (s *Session) sendSpawnChunks() error {

	//used for encoding height-map, sections & light data
	data := protocol.GetBuffer(1024 * 10)
	defer protocol.PutBuffer(data)

	x, _, z := s.state.Position()

	chunkX := int32(math.Floor(x / 16))
	chunkZ := int32(math.Floor(z / 16))

	for i := chunkX - s.ViewDistance; i < chunkX+s.ViewDistance; i++ {
		for j := chunkZ - s.ViewDistance; j < chunkZ+s.ViewDistance; j++ {
			//i, j = 0, 0
			c, err := s.state.Dimension().Chunk(i, j)
			if err != nil {
				panic(err)
			}

			c.NetworkEncode(data)

			s.conn.SendPacket(&packet.ChunkData{
				X:    c.X,
				Z:    c.Z,
				Data: data.Bytes(),
			})
			data.Reset()

			//time.Sleep(time.Second * 3)
			//return nil
		}
	}

	return nil
}

func (s *Session) Teleport(x, y, z float64) {
	s.state.Move(x, y, z)
	s.conn.SendPacket(&packet.SyncPlayerPos{
		X: x,
		Y: y,
		Z: z,
	})

	s.conn.SendPacket(&packet.SetCenterChunk{ChunkX: int32(x / 16), ChunkZ: int32(z / 16)})
}

func (s *Session) IsSpawned(p *Session) bool {
	return s.isEntitySpawned(p.eid)
}

func (s *Session) isEntitySpawned(id int32) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.spawnedEntities[id]
	return ok
}

func (s *Session) Despawn(p *Session) {
	if !s.IsSpawned(p) {
		return
	}
	s.unspawnEntity(p.eid)

	s.conn.SendPacket(&packet.DestroyEntities{EntityIds: []int32{p.eid}})
}

// unspawnEntity only removes the entity id from the map
func (s *Session) unspawnEntity(id int32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.spawnedEntities, id)
}

// spawnEntity only adds the entity id to the map
func (s *Session) spawnEntity(id int32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.spawnedEntities[id] = struct{}{}
}

// InView checks if the player is able to see the coordinates specified
func (s *Session) InView(x2, y2, z2 float64) bool {
	x1, y1, z1 := s.state.Position()
	distance := math.Sqrt((x1-x2)*(x1-x2) + (y1-y2)*(y1-y2) + (z1-z2)*(z1-z2))

	return float64(s.ViewDistance)*16 > distance
}

func (s *Session) Knockback(x, y, z int16) {
	s.conn.SendPacket(&packet.SetEntityVelocity{
		EntityID: s.eid,
		X:        x,
		Y:        y,
		Z:        z,
	})
}

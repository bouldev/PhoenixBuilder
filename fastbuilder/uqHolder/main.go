package uqHolder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
)

var Version = [3]byte{0, 0, 3}

type Player struct {
	UUID           uuid.UUID
	EntityUniqueID int64
	//GameModeAfterChange     int32
	LoginTime               time.Time
	LoginTick               uint64
	Username                string
	PlatformChatID          string
	BuildPlatform           int32
	SkinID                  string
	PropertiesFlag          uint32
	CommandPermissionLevel  uint32
	ActionPermissions       uint32
	OPPermissionLevel       uint32
	CustomStoredPermissions uint32
	DeviceID                string
	// only when the player can be seen by bot
	EntityRuntimeID uint64
	Entity          *Entity
	// PlayerUniqueID is a unique identifier of the player. It appears it is not required to fill this field
	// out with a correct value. Simply writing 0 seems to work.
	// PlayerUniqueID int64
}

type PosRepresent struct {
	Position       mgl32.Vec3
	Velocity       mgl32.Vec3
	Pitch          float32
	Yaw            float32
	HeadYaw        float32
	LastUpdateTick uint64
	Rotation       mgl32.Vec3
	MaskedRotation mgl32.Vec3
}

type Entity struct {
	RuntimeID        uint64
	Attributes       []protocol.Attribute
	Metadata         map[uint32]interface{}
	Slots            map[byte]*Equipment
	LastPacketSlot   byte
	OutOfRangeAtTick uint64
	IsPlayer         bool

	LastUpdateTick uint64
	LastPosInfo    PosRepresent

	UniqueID    int64
	EntityType  string
	EntityLinks []protocol.EntityLink
}

type Equipment struct {
	NewItem  protocol.ItemInstance
	Slot     byte
	WindowID byte
}

type GameRule struct {
	CanBeModifiedByPlayer bool
	Value                 interface{}
}
type UQHolder struct {
	VERSION                    string
	ConnectTime                time.Time
	WorldName                  string
	BotRandomID                int64
	BotUniqueID                int64
	BotRuntimeID               uint64
	BotName                    string
	BotIdentity                string
	CompressThreshold          uint16
	CurrentTick                uint64
	WorldGameMode              int32
	WorldDifficulty            uint32
	InventorySlot              map[uint32]protocol.ItemInstance
	playersByUUID              map[[16]byte]*Player
	PlayersByEntityID          map[int64]*Player
	EntitiesByRuntimeID        map[uint64]*Entity
	entitiesByUniqueID         map[int64]*Entity
	Time                       int32
	DayTime                    int32
	DayTimePercent             float32
	OnConnectWoldSpawnPosition protocol.BlockPos
	WorldSpawnPosition         map[int32]protocol.BlockPos
	BotSpawnPosition           map[int32]protocol.BlockPos
	CommandsEnabled            bool
	GameRules                  map[string]*GameRule
	InventoryContent           map[uint32][]protocol.ItemInstance
	PlayerHotBar               packet.PlayerHotBar
	// AvailableCommands   packet.AvailableCommands
	BotPos                PosRepresent
	BotOnGround           bool
	BotHealth             int32
	CommandRelatedEnums   []*packet.UpdateSoftEnum
	displayUnknownPackets bool
	mu                    sync.Mutex
}

func NewUQHolder(BotRuntimeID uint64) *UQHolder {
	uq := &UQHolder{
		VERSION:               fmt.Sprintf("%d.%d.%d", Version[0], Version[1], Version[2]),
		BotRuntimeID:          BotRuntimeID,
		InventorySlot:         map[uint32]protocol.ItemInstance{},
		playersByUUID:         map[[16]byte]*Player{},
		PlayersByEntityID:     map[int64]*Player{},
		WorldSpawnPosition:    map[int32]protocol.BlockPos{},
		BotSpawnPosition:      map[int32]protocol.BlockPos{},
		EntitiesByRuntimeID:   map[uint64]*Entity{},
		entitiesByUniqueID:    map[int64]*Entity{},
		GameRules:             map[string]*GameRule{},
		InventoryContent:      map[uint32][]protocol.ItemInstance{},
		CommandRelatedEnums:   make([]*packet.UpdateSoftEnum, 0),
		displayUnknownPackets: false,
		mu:                    sync.Mutex{},
	}
	go func() {
		t := time.NewTicker(50 * time.Millisecond)
		gcTime := uint64(20 * 60 * 10)
		for {
			<-t.C
			uq.CurrentTick++
			if uq.CurrentTick%gcTime == gcTime-1 {
				uq.gc((uq.CurrentTick + 10) - gcTime)
			}
		}
	}()
	return uq
}

func (uq *UQHolder) UpdateTick(tick uint64) {
	uq.CurrentTick = tick
}

func (uq *UQHolder) DisplayUnKnownPackets(b bool) {
	uq.displayUnknownPackets = b
}

func (uq *UQHolder) gc(deadline uint64) {
	//fmt.Println("gc")
	for _, p := range uq.playersByUUID {
		if p.Entity != nil && p.EntityRuntimeID != 0 {
			if e, hask := uq.EntitiesByRuntimeID[p.EntityRuntimeID]; hask && e.LastUpdateTick < deadline {
				delete(uq.entitiesByUniqueID, e.UniqueID)
				delete(uq.EntitiesByRuntimeID, p.EntityRuntimeID)
			}
			p.Entity = nil
			p.EntityRuntimeID = 0
			p.EntityUniqueID = 0
		}
	}
	gcEID := make([]uint64, 0)
	for rtid, e := range uq.EntitiesByRuntimeID {
		if e.LastUpdateTick < deadline {
			gcEID = append(gcEID, rtid)
		}
	}
	for _, rtid := range gcEID {
		if e, hask := uq.EntitiesByRuntimeID[rtid]; hask && e.LastUpdateTick < deadline {
			delete(uq.entitiesByUniqueID, e.UniqueID)
			delete(uq.EntitiesByRuntimeID, rtid)
		}
	}
}

func (uq *UQHolder) Marshal() []byte {
	buf := bytes.NewBuffer([]byte{Version[0], Version[1], Version[2]})
	compressor := brotli.NewWriter(buf)
	err := json.NewEncoder(compressor).Encode(uq)
	if err != nil {
		panic(err)
	}
	compressor.Close()
	return buf.Bytes()
}

func IsCapable(bs []byte) error {
	if len(bs) < 3 {
		return fmt.Errorf("version length error")
	}
	if bs[0] != Version[0] {
		return fmt.Errorf("version MAJOR mismatch (local=%v,remote=%v)", Version[0], bs[0])
	}
	if bs[1] != Version[1] {
		return fmt.Errorf("version MINOR mismatch (local=%v,remote=%v)", Version[1], bs[1])
	}
	if bs[2] != Version[2] {
		return fmt.Errorf("version Patch mismatch (local=%v,remote=%v)", Version[2], bs[2])
	}
	return nil
}

func (uq *UQHolder) UnMarshal(bs []byte) error {
	if err := IsCapable(bs); err != nil {
		return err
	}
	buf := bytes.NewBuffer(bs[3:])
	decompressor := brotli.NewReader(buf)
	err := json.NewDecoder(decompressor).Decode(uq)
	if err != nil {
		return err
	}
	for _, entity := range uq.EntitiesByRuntimeID {
		uq.entitiesByUniqueID[entity.UniqueID] = entity
	}
	for _, player := range uq.PlayersByEntityID {
		uq.playersByUUID[player.UUID] = player
		if player.EntityRuntimeID != 0 {
			if e, ok := uq.EntitiesByRuntimeID[player.EntityRuntimeID]; ok {
				player.Entity = e
			}
		}
	}
	//botName := uq.GetBotName()
	//fmt.Println(botName)
	return nil
}

func (uq *UQHolder) GetEntityByRuntimeID(EntityRuntimeID uint64) *Entity {
	var e *Entity
	if _e, ok := uq.EntitiesByRuntimeID[EntityRuntimeID]; !ok {
		e = &Entity{
			RuntimeID:      EntityRuntimeID,
			LastPacketSlot: 255,
			Slots:          map[byte]*Equipment{},
			LastUpdateTick: uq.CurrentTick,
			UniqueID:       0,
		}
		uq.EntitiesByRuntimeID[EntityRuntimeID] = e
	} else {
		e = _e
	}
	return e
}

func (uq *UQHolder) GetPlayersByUUID(ud uuid.UUID) *Player {
	if p, ok := uq.playersByUUID[ud]; ok {
		return p
	} else {
		return nil
	}
}

func GetStringContents(s string) []string {
	_s := strings.Split(s, " ")
	for i, c := range _s {
		_s[i] = strings.TrimSpace(c)
	}
	ss := make([]string, 0, len(_s))
	for _, c := range _s {
		if c != "" {
			ss = append(ss, c)
		}
	}
	return ss
}
func ToPlainName(name string) string {
	if strings.Contains(name, ">") {
		name = strings.ReplaceAll(name, ">", " ")
		name = strings.ReplaceAll(name, "<", " ")
	}
	if name != "" {
		names := GetStringContents(name)
		name = names[len(names)-1]
	}
	return name
}

func (uq *UQHolder) GetBotName() string {
	return uq.BotName
}

func (uq *UQHolder) GetBotIdentity() string {
	return uq.BotIdentity
}

func (uq *UQHolder) GetBotUniqueID() int64 {
	return uq.BotUniqueID
}

func (uq *UQHolder) GetBotRuntimeID() uint64 {
	return uq.BotRuntimeID
}

// var recordNoPlayerEntity = false

func (uq *UQHolder) Update(pk packet.Packet) {
	uq.mu.Lock()
	defer uq.mu.Unlock()
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println("UQHolder Update Error: ", r)
			debug.PrintStack()
		}
	}()
	switch p := pk.(type) {
	case *packet.NetworkSettings:
		uq.CompressThreshold = p.CompressionThreshold
	case *packet.InventorySlot:
		uq.InventorySlot[p.Slot] = p.NewItem
	case *packet.PlayerList:
		if p.ActionType == packet.PlayerListActionAdd {
			for _, e := range p.Entries {
				player := &Player{
					UUID:           e.UUID,
					EntityUniqueID: e.EntityUniqueID,
					Username:       ToPlainName(e.Username),
					PlatformChatID: e.PlatformChatID,
					BuildPlatform:  e.BuildPlatform,
					SkinID:         e.Skin.SkinID,
					LoginTime:      time.Now(),
					LoginTick:      uq.CurrentTick,
				}
				uq.playersByUUID[e.UUID] = player
				uq.PlayersByEntityID[e.EntityUniqueID] = player
			}
		} else {
			for _, e := range p.Entries {
				if p, ok := uq.playersByUUID[e.UUID]; ok {
					if p.Entity != nil && p.EntityRuntimeID != 0 {
						if e, hask := uq.EntitiesByRuntimeID[p.EntityRuntimeID]; hask {
							delete(uq.entitiesByUniqueID, e.UniqueID)
							delete(uq.EntitiesByRuntimeID, p.EntityRuntimeID)
						}
						p.Entity = nil
					}
					delete(uq.PlayersByEntityID, p.EntityUniqueID)
					delete(uq.playersByUUID, e.UUID)
				}
			}
		}
	case *packet.AdventureSettings:
		player := uq.PlayersByEntityID[p.PlayerUniqueID]
		if player == nil {
			player = &Player{}
		}
		player.PropertiesFlag = p.Flags
		player.CommandPermissionLevel = p.CommandPermissionLevel
		player.ActionPermissions = p.ActionPermissions
		player.OPPermissionLevel = p.PermissionLevel
		player.CustomStoredPermissions = p.CustomStoredPermissions
	case *packet.SetTime:
		uq.Time = p.Time
		uq.DayTime = p.Time % 24000
		uq.DayTimePercent = float32(uq.DayTime) / 24000.0
	case *packet.SetCommandsEnabled:
		uq.CommandsEnabled = p.Enabled
	case *packet.UpdateAttributes:
		// e := uq.GetEntityByRuntimeID(p.EntityRuntimeID)
		// e.LastUpdateTick = p.Tick
		// e.Attributes = p.Attributes
		// uq.UpdateTick(p.Tick)
	case *packet.GameRulesChanged:
		for _, r := range p.GameRules {
			uq.GameRules[r.Name] = &GameRule{
				CanBeModifiedByPlayer: r.CanBeModifiedByPlayer,
				Value:                 r.Value,
			}
		}
	case *packet.InventoryContent:
		for key, value := range p.Content {
			if value.Stack.ItemType.NetworkID != -1 {
				if uq.InventoryContent[p.WindowID] == nil {
					uq.InventoryContent[p.WindowID] = make([]protocol.ItemInstance, len(p.Content))
				}
				uq.InventoryContent[p.WindowID][key] = value
			}
		}
	case *packet.AvailableCommands:
		// too large
		// uq.AvailableCommands = *p

	case *packet.SetActorData:
		e := uq.GetEntityByRuntimeID(p.EntityRuntimeID)
		e.LastUpdateTick = p.Tick
		e.Metadata = p.EntityMetadata
		uq.UpdateTick(p.Tick)
	case *packet.MovePlayer:
		if p.EntityRuntimeID == uq.BotRuntimeID {
			uq.BotOnGround = p.OnGround
			uq.BotPos = PosRepresent{
				Position:       p.Position,
				Pitch:          p.Pitch,
				Yaw:            p.Yaw,
				HeadYaw:        p.HeadYaw,
				LastUpdateTick: p.Tick,
			}
		}
		e := uq.GetEntityByRuntimeID(p.EntityRuntimeID)
		e.LastPosInfo = PosRepresent{
			Position:       p.Position,
			Pitch:          p.Pitch,
			Yaw:            p.Yaw,
			HeadYaw:        p.HeadYaw,
			LastUpdateTick: p.Tick,
		}
		e.LastUpdateTick = p.Tick
		uq.UpdateTick(p.Tick)
	case *packet.CorrectPlayerMovePrediction:
		uq.BotPos.Position = p.Position
		uq.BotPos.LastUpdateTick = p.Tick
		uq.GetEntityByRuntimeID(uq.BotRuntimeID).LastPosInfo.Position = p.Position
		uq.GetEntityByRuntimeID(uq.BotRuntimeID).LastPosInfo.LastUpdateTick = p.Tick
		uq.GetEntityByRuntimeID(uq.BotRuntimeID).LastUpdateTick = p.Tick
		uq.BotOnGround = p.OnGround
		uq.UpdateTick(p.Tick)

	case *packet.AddPlayer:
		player := uq.PlayersByEntityID[p.EntityUniqueID]
		entity := uq.GetEntityByRuntimeID(p.EntityRuntimeID)
		entity.IsPlayer = true
		entity.LastUpdateTick = uq.CurrentTick
		player.Entity = entity
		entity.LastUpdateTick = uq.CurrentTick
		entity.LastPosInfo.LastUpdateTick = uq.CurrentTick
		entity.LastPosInfo.Position = p.Position
		entity.LastPosInfo.Pitch = p.Pitch
		entity.LastPosInfo.Yaw = p.Yaw
		entity.LastPosInfo.HeadYaw = p.HeadYaw
		player.PropertiesFlag = p.Flags
		player.CommandPermissionLevel = p.CommandPermissionLevel
		player.ActionPermissions = p.ActionPermissions
		player.OPPermissionLevel = p.PermissionLevel
		player.CustomStoredPermissions = p.CustomStoredPermissions
		player.DeviceID = p.DeviceID
	case *packet.MobEquipment:
		entity := uq.GetEntityByRuntimeID(p.EntityRuntimeID)
		entity.Slots[p.InventorySlot] = &Equipment{
			Slot:     p.InventorySlot,
			NewItem:  p.NewItem,
			WindowID: p.WindowID,
		}
		entity.LastUpdateTick = uq.CurrentTick
		entity.LastPacketSlot = p.InventorySlot
	case *packet.SetHealth:
		uq.BotHealth = p.Health
	case *packet.UpdateSoftEnum:
		// uq.CommandRelatedEnums = append(uq.CommandRelatedEnums, p)
	case *packet.AddActor:
		// if !recordNoPlayerEntity {
		// 	return
		// }
		// entity := uq.GetEntityByRuntimeID(p.EntityRuntimeID)
		// entity.IsPlayer = false
		// entity.UniqueID = p.EntityUniqueID
		// uq.entitiesByUniqueID[p.EntityUniqueID] = entity
		// entity.EntityType = p.EntityType
		// entity.LastUpdateTick = uq.CurrentTick
		// entity.LastPosInfo.LastUpdateTick = uq.CurrentTick
		// entity.LastPosInfo.Position = p.Position
		// entity.LastPosInfo.Velocity = p.Velocity
		// entity.LastPosInfo.Pitch = p.Pitch
		// entity.LastPosInfo.Yaw = p.Yaw
		// entity.LastPosInfo.HeadYaw = p.Yaw
		// entity.Attributes = p.Attributes
		// entity.Metadata = p.EntityMetadata
		// entity.EntityLinks = p.EntityLinks

	case *packet.RemoveActor:
		if entity, ok := uq.entitiesByUniqueID[p.EntityUniqueID]; ok {
			rtID := entity.RuntimeID
			if !entity.IsPlayer {
				if _, ok := uq.EntitiesByRuntimeID[rtID]; ok {
					delete(uq.EntitiesByRuntimeID, rtID)
				}
				delete(uq.entitiesByUniqueID, p.EntityUniqueID)
			}
		}
	case *packet.MoveActorDelta:
		// entity := uq.GetEntityByRuntimeID(p.EntityRuntimeID)
		// entity.LastPosInfo.LastUpdateTick = uq.CurrentTick
		// entity.LastPosInfo.Position = p.Position
		// entity.LastPosInfo.Rotation = p.Rotation
		// if x := p.Rotation.X(); x != 0 {
		// 	entity.LastPosInfo.MaskedRotation[0] = x
		// }
		// if y := p.Rotation.Y(); y != 0 {
		// 	entity.LastPosInfo.MaskedRotation[1] = y
		// }
		// if z := p.Rotation.Z(); z != 0 {
		// 	entity.LastPosInfo.MaskedRotation[2] = z
		// }

	case *packet.SetActorMotion:
		// entity := uq.GetEntityByRuntimeID(p.EntityRuntimeID)
		// entity.LastPosInfo.LastUpdateTick = uq.CurrentTick
		// entity.LastPosInfo.Velocity = p.Velocity

	// not fully supported
	case *packet.Respawn:
		if p.EntityRuntimeID == 0 {
			uq.GetEntityByRuntimeID(uq.BotRuntimeID).LastPosInfo.Position = p.Position
		} else {
			if marshal, err := json.Marshal(pk); err == nil {
				fmt.Println("Respawn Data ignored: ", string(marshal))
			}

		}
	// not fully supported, large amount of data
	case *packet.LevelEvent:

	// meaning not clear
	case *packet.SetSpawnPosition:
		if p.SpawnType == packet.SpawnTypePlayer {
			uq.BotSpawnPosition[p.Dimension] = p.Position
			uq.WorldSpawnPosition[p.Dimension] = p.SpawnPosition // not sure
		} else {
			uq.BotSpawnPosition[p.Dimension] = p.Position // not sure
			uq.WorldSpawnPosition[p.Dimension] = p.SpawnPosition
		}
	case *packet.SetDefaultGameType:
		uq.WorldGameMode = p.GameType
	case *packet.SetDifficulty:
		uq.WorldDifficulty = p.Difficulty
	//case *packet.UpdatePlayerGameType:
	// some thing error
	//fmt.Println("mode UpdatePlayerGameType")
	//for _, player := range uq.PlayersByEntityID {
	//	if player.PlayerUniqueID == p.PlayerUniqueID {
	//		player.GameModeAfterChange = p.GameType
	//		fmt.Println("mode changed")
	//	}
	//}
	// meaning not clear
	/*case *packet.PlayerHotBar:
		uq.PlayerHotBar = *p
	// not supported, plan to support
	case *packet.InventoryTransaction:
	// not supported, plan to support
	case *packet.ActorEvent:
	// no plan to support following
	case *packet.LevelChunk:
	case *packet.NetworkChunkPublisherUpdate:
	case *packet.BiomeDefinitionList:
	case *packet.AvailableActorIdentifiers:
	case *packet.CraftingData:
	case *packet.ChunkRadiusUpdated:
	case *packet.LevelSoundEvent:
	case *packet.Animate:
	case *packet.ItemComponent:
	case *packet.CreativeContent:
	case *packet.UpdateBlock:
	case *packet.BlockActorData:
	case *packet.PlayerFog:
	case *packet.Text:
	case *packet.AddItemActor:
	// no need to support
	case *packet.PlayStatus:
	// no need to support
	*/
	default:
		if !uq.displayUnknownPackets {
			break
		}
		marshal, err := json.Marshal(pk)
		if err != nil {
			println(err)
		} else {
			jsonStr := string(marshal)
			if len(jsonStr) < 300 {
				fmt.Println(pk.ID(), " : ", jsonStr)
			} else {
				fmt.Println(pk.ID(), " : ", jsonStr[:300])
			}
		}
	}
}

func (uq *UQHolder) UpdateFromConn(conn *minecraft.Conn) {
	gd := conn.GameData()
	uq.BotUniqueID = gd.EntityUniqueID
	uq.BotRuntimeID = gd.EntityRuntimeID
	uq.ConnectTime = time.Time{} // No longer needed
	uq.WorldName = gd.WorldName
	uq.WorldGameMode = gd.WorldGameMode
	uq.WorldDifficulty = uint32(gd.Difficulty)
	uq.OnConnectWoldSpawnPosition = gd.WorldSpawn
	cd := conn.ClientData()
	uq.BotRandomID = cd.ClientRandomID
	uq.BotName = conn.IdentityData().DisplayName
	uq.BotIdentity = conn.IdentityData().Identity
}

//func main() {
//	TypePool := packet.NewPool()
//	fp, err := os.OpenFile("dump.gob", os.O_RDONLY, 0755)
//	if err != nil {
//		panic(err)
//	}
//	cachedBytes := make([][]byte, 0)
//	err = gob.NewDecoder(fp).Decode(&cachedBytes)
//	if err != nil {
//		panic(err)
//	}
//	holder := NewUQHolder(0)
//	paddingByte := []byte{}
//	safeDecode := func(pktByte []byte) (pkt packet.Packet) {
//		pktID := uint32(pktByte[0])
//		defer func() {
//			if r := recover(); r != nil {
//				fmt.Println(pktID, "decode fail ", pkt)
//			}
//			return
//		}()
//		pkt = TypePool[pktID]()
//		pkt.Unmarshal(protocol.NewReader(bytes.NewReader(
//			bytes.Join([][]byte{pktByte[1:], paddingByte}, []byte{}),
//		), 0))
//		return
//	}
//	for _, pktByte := range cachedBytes {
//      if len(pktByte)==0 {
//			continue
//		}
//		pkt := safeDecode(pktByte)
//		if pkt != nil {
//			holder.Update(pkt)
//		}
//	}
//}

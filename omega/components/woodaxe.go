package components

import (
	"encoding/json"
	"fmt"
	"math"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/mirror/items"
	"phoenixbuilder/omega/defines"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
)

type WoodAxe struct {
	*BasicComponent
	currentRequestUser       string
	currentPlayerPk          *packet.AddPlayer
	currentPlayerKit         defines.PlayerKit
	woodAxeRTID              int32
	woodAxeOn                bool
	currentPos               mgl32.Vec3
	currentYaw, currentPitch float32
	esp                      float64
	nan                      float64
	structureBlock           *StructureBlock
}

type StructureBlock struct {
	Pos          define.CubePos
	pos          protocol.BlockPos
	packetSender func(packet.Packet)
	basicPacket  *packet.StructureBlockUpdate
	operatorUID  int64
}

func NewStructureBlock(Pos define.CubePos, packetSender func(packet.Packet), uid int64) *StructureBlock {
	block := &StructureBlock{Pos: Pos, packetSender: packetSender}
	block.pos = protocol.BlockPos{int32(block.Pos[0]), int32(block.Pos[1]), int32(block.Pos[2])}
	block.operatorUID = uid
	block.basicPacket = &packet.StructureBlockUpdate{
		Position:           block.pos,
		StructureName:      "",
		DataField:          "",
		IncludePlayers:     false,
		ShowBoundingBox:    true,
		StructureBlockType: 1,
		Settings: protocol.StructureSettings{
			PaletteName:               "default",
			IgnoreEntities:            true,
			IgnoreBlocks:              false,
			Size:                      protocol.BlockPos{30, 6, 14},
			Offset:                    protocol.BlockPos{0, 0, 0},
			LastEditingPlayerUniqueID: uid,
			Rotation:                  0,
			Mirror:                    0,
			AnimationMode:             0,
			AnimationDuration:         0,
			Integrity:                 100,
			Seed:                      0,
			Pivot:                     mgl32.Vec3{0, 0, 0},
		},
		RedstoneSaveMode: 0,
		ShouldTrigger:    false,
	}
	return block
}

func (o *StructureBlock) OffBound() {
	o.packetSender(o.basicPacket)
}

func (o *StructureBlock) OnBound() {
	o.packetSender(o.basicPacket)
}

func (o *StructureBlock) IndicateCube(start define.CubePos, end define.CubePos) {
	start, end = sortPos(start, end)
	offset := start.Sub(o.Pos)
	size := end.Sub(start).Add(define.CubePos{1, 1, 1})
	o.basicPacket.Settings.Offset = protocol.BlockPos{int32(offset[0]), int32(offset[1]), int32(offset[2])}
	o.basicPacket.Settings.Size = protocol.BlockPos{int32(size[0]), int32(size[1]), int32(size[2])}
	o.packetSender(o.basicPacket)
}

func sortPos(pa define.CubePos, pb define.CubePos) (start define.CubePos, end define.CubePos) {
	if pa[0] > pb[0] {
		start[0] = pb[0]
		end[0] = pa[0]
	} else {
		start[0] = pa[0]
		end[0] = pb[0]
	}
	if pa[1] > pb[1] {
		start[1] = pb[1]
		end[1] = pa[1]
	} else {
		start[1] = pa[1]
		end[1] = pb[1]
	}
	if pa[2] > pb[2] {
		start[2] = pb[2]
		end[2] = pa[2]
	} else {
		start[2] = pa[2]
		end[2] = pb[2]
	}
	return
}

func (o *WoodAxe) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, o)
	if err != nil {
		panic(err)
	}
	for rtid, desc := range items.RuntimeIDToItemNameMapping {
		if desc.ItemName == "wooden_axe" {
			o.woodAxeRTID = rtid
		}
	}
	o.esp = 0.00001
	o.nan = math.NaN()
}

func (o *WoodAxe) onAnyPacket(pkt packet.Packet) {
	switch pk := pkt.(type) {
	case *packet.LevelChunk:
		break
	case *packet.NetworkChunkPublisherUpdate:
		break
	// case *packet.UpdateBlock:
	// 	break
	case *packet.LevelEvent:
		break
	case *packet.Text:
		break
	case *packet.ActorEvent:
		break
	case *packet.RemoveActor:
		break
	case *packet.MovePlayer:
		break
	case *packet.MoveActorDelta:
		break
	case *packet.SetActorData:
		break
	case *packet.AddPlayer:
		break
	case *packet.SetActorMotion:
		break
	case *packet.SetLastHurtBy:
		break
	case *packet.CommandOutput:
		break
	case *packet.LevelSoundEvent:
		break
	default:
		m, err := json.Marshal(pk)
		if err == nil {
			fmt.Println(pk.ID(), " ", string(m))
		} else {
			fmt.Println(err)
		}
	}
}

type CmdRespDataSetPos struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}
type cmdRespDataSet struct {
	Pos        CmdRespDataSetPos `json:"position"`
	StatusCode int               `json:"statusCode"`
}

func (o *WoodAxe) onInitStructureBlock(pos define.CubePos) {
	uid := o.Frame.GetUQHolder().BotUniqueID
	o.structureBlock = NewStructureBlock(pos, o.Frame.GetGameControl().SendMCPacket, uid)
	o.structureBlock.OnBound()
}

func (o *WoodAxe) onInitWorkSapce(pk *packet.AddPlayer) {
	o.currentPlayerPk = pk
	o.currentPlayerKit = o.Frame.GetGameControl().GetPlayerKit(pk.Username)
	o.currentPlayerKit.Say("建筑师 " + pk.Username + " 已进入小木斧范围内")
	o.onSeeItem(pk.HeldItem.Stack.NetworkID)
	o.Frame.GetGameControl().SendCmdAndInvokeOnResponse("setblock ~~~ structure_block", func(output *packet.CommandOutput) {
		if output.SuccessCount > 0 {
			respData := cmdRespDataSet{}
			if err := json.Unmarshal([]byte(output.DataSet), &respData); err == nil {
				cubePos := define.CubePos{respData.Pos.X, respData.Pos.Y, respData.Pos.Z}
				fmt.Println(cubePos)
				o.onInitStructureBlock(cubePos)
				return
			}
		}
		o.Frame.GetGameControl().SayTo("@a", "小木斧初始化失败")
	})
}

func (o *WoodAxe) onAddPlayer(pkt packet.Packet) {
	pk := pkt.(*packet.AddPlayer)
	if pk.Username == o.currentRequestUser {
		o.onInitWorkSapce(pk)
	}
}

func (o *WoodAxe) onAnimate(pkt packet.Packet) {
	pk := pkt.(*packet.Animate)
	if o.currentPlayerPk != nil && pk.EntityRuntimeID == o.currentPlayerPk.EntityRuntimeID {
		// fmt.Println("animate!")
		if o.woodAxeOn {
			o.onPosInput()
		}
	}
}

func (o *WoodAxe) onSeeMobItem(pk *packet.MobEquipment) {
	if o.currentPlayerPk == nil || pk.EntityRuntimeID != o.currentPlayerPk.EntityRuntimeID {
		return
	}
	rtid := pk.NewItem.Stack.NetworkID
	o.onSeeItem(rtid)
}

func (o *WoodAxe) onSeeItem(rtid int32) {
	if o.woodAxeRTID == rtid {
		o.onWoodAxe()
	} else {
		o.offWoodAxe()
	}
}

func (o *WoodAxe) onWoodAxe() {
	o.currentPlayerKit.Say("小木斧状态: 开")
	o.woodAxeOn = true
}

func (o *WoodAxe) offWoodAxe() {
	o.currentPlayerKit.Say("小木斧状态: 关")
	o.woodAxeOn = false
}

func (o *WoodAxe) computeNextDelta(currentPos float64, lookAtDelta float64) float64 {
	if math.Abs(lookAtDelta) < o.esp {
		return o.nan
	}
	currentBlock := math.Floor(currentPos)
	targetPos := float64(0)
	if lookAtDelta > 0 {
		targetPos = currentBlock + 1
		if math.Abs(targetPos-currentPos) < o.esp {
			targetPos += 1
		}
	} else {
		targetPos = currentBlock
		if math.Abs(targetPos-currentPos) < o.esp {
			targetPos -= 1
		}
	}
	return (targetPos - currentPos) / lookAtDelta
}

func (o *WoodAxe) computeNextPos(currentPos mgl32.Vec3, delta mgl32.Vec3) float64 {
	posDelta := o.nan
	dX := o.computeNextDelta(float64(currentPos[0]), float64(delta[0]))
	if dX != o.nan {
		posDelta = dX
	}
	dY := o.computeNextDelta(float64(currentPos[1]), float64(delta[1]))
	if dY != o.nan && (posDelta == o.nan || math.Abs(dY) < math.Abs(posDelta)) {
		posDelta = dY
	}
	dZ := o.computeNextDelta(float64(currentPos[2]), float64(delta[2]))
	if dZ != o.nan && (posDelta == o.nan || math.Abs(dZ) < math.Abs(posDelta)) {
		posDelta = dZ
	}
	return posDelta
}

func (o *WoodAxe) posToBlockPos(pos mgl32.Vec3) define.CubePos {
	return define.CubePos{int(math.Floor(float64(pos[0]))), int(math.Floor(float64(pos[1]))), int(math.Floor(float64(pos[2])))}
}

func (o *WoodAxe) computeNextXPos(currentPos mgl32.Vec3, delta mgl32.Vec3, numNext int) (nextPoses []define.CubePos) {
	nextPoses = []define.CubePos{}
	for i := 0; i < numNext; i++ {
		d := o.computeNextPos(currentPos, delta)
		if d == o.nan {
			break
		}
		d += o.esp
		currentPos = currentPos.Add(delta.Mul(float32(d)))
		nextPoses = append(nextPoses, o.posToBlockPos(currentPos))
	}
	return nextPoses
}

func (o *WoodAxe) onPosInput() {
	o.currentPlayerKit.Say(o.getPosString())
	// headAtBlock := protocol.BlockPos{int32(math.Floor(float64(o.currentPos[0]))), int32(math.Floor(float64(o.currentPos[1]))), int32(math.Floor(float64(o.currentPos[2])))}
	// fmt.Println(headAtBlock)
	deltaY := math.Sin(float64(-o.currentPitch / 180 * math.Pi))
	deltaXZ := math.Cos(float64(o.currentPitch / 180 * math.Pi))
	deltaX := -math.Sin(float64(o.currentYaw/180*math.Pi)) * deltaXZ
	deltaZ := math.Cos(float64(o.currentYaw/180*math.Pi)) * deltaXZ
	lookAt := mgl32.Vec3{float32(deltaX), float32(deltaY), float32(deltaZ)}
	// o.currentPlayerKit.Say(fmt.Sprintf("LookAtDelta :[%.1f, %.1f, %.1f]", deltaX, deltaY, deltaZ))
	nextTenBlocks := o.computeNextXPos(o.currentPos, lookAt, 30)
	world := o.Frame.GetWorld()
	selected := false
	selectedBlockName := ""
	var selectedBlockPos define.CubePos
	for _, pos := range nextTenBlocks {
		if rtid, found := world.Block(pos); found {
			if rtid == chunk.AirRID {
				continue
			}
			if blockDesc, hasB := chunk.RuntimeIDToBlock(rtid); hasB {
				selected = true
				selectedBlockName = blockDesc.Name
				selectedBlockPos = pos
			}
			break
		}
	}
	if selected {
		o.currentPlayerKit.Say(fmt.Sprintf("§l§b选中 %v @ %v", strings.ReplaceAll(selectedBlockName, "minecraft:", ""), selectedBlockPos))
		o.structureBlock.IndicateCube(selectedBlockPos, selectedBlockPos)
	} else {
		o.currentPlayerKit.Say(fmt.Sprintf("§l§a未选中方块！"))
	}
}

func (o *WoodAxe) getPosString() string {
	return fmt.Sprintf("Pos: [%.1f, %.1f, %.1f] Angle: [%.1f, %.1f]", o.currentPos[0], o.currentPos[1], o.currentPos[2], o.currentYaw, o.currentPitch)
}

func (o *WoodAxe) onPlayerMove(pk *packet.MovePlayer) {
	if pk.EntityRuntimeID == o.currentPlayerPk.EntityRuntimeID {
		o.currentPos = pk.Position
		o.currentYaw = pk.HeadYaw
		o.currentPitch = pk.Pitch
	}
}

func (o *WoodAxe) BlockUpdate(pos define.CubePos, origRTID uint32, currentRTID uint32) {
	orig, _ := chunk.RuntimeIDToBlock(origRTID)
	current, _ := chunk.RuntimeIDToBlock(currentRTID)
	hint := fmt.Sprintf("%v:%v->%v", pos, strings.ReplaceAll(orig.Name, "minecraft:", ""), strings.ReplaceAll(current.Name, "minecraft:", ""))
	fmt.Println(hint)
	// o.currentPlayerKit.ActionBar(hint)
}

func (o *WoodAxe) Inject(frame defines.MainFrame) {
	o.Frame = frame
	// frame.GetGa,e
	o.currentRequestUser = "2401PT"
	frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDAddPlayer, o.onAddPlayer)
	frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDAnimate, o.onAnimate)
	o.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDMobEquipment, func(p packet.Packet) {
		o.onSeeMobItem(p.(*packet.MobEquipment))
	})
	o.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDMovePlayer, func(p packet.Packet) {
		o.onPlayerMove(p.(*packet.MovePlayer))
	})
	o.Frame.GetGameListener().AppendOnBlockUpdateInfoCallBack(o.BlockUpdate)
	// frame.GetGameListener().SetOnAnyPacketCallBack(o.onAnyPacket)
}

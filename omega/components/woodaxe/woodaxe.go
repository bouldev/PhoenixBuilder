package woodaxe

import (
	"encoding/json"
	"fmt"
	"math"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/mirror/items"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"time"

	"github.com/go-gl/mathgl/mgl32"
)

const (
	AreaPosOne = iota
	AreaPosTwo
	BasePointOne
	BasePointTwo
	NotSelect = 255
)

type selectInfo struct {
	selected         map[uint8]bool
	pos              map[uint8]define.CubePos
	nextSelect       uint8
	currentSelectID  uint8
	currentSelectPos define.CubePos
	triggerFN        func()
}

type actionsOccupied struct {
	occupied       bool
	continuousCopy bool
	undo           bool
	largeFill      bool
}

type WoodAxe struct {
	*defines.BasicComponent
	Operators                            []string `json:"授权使用者"`
	Triggers                             []string `json:"触发词"`
	UseLargeFill                         bool     `json:"使用LargeFill功能"`
	releaseChan                          chan struct{}
	chanCreated                          bool
	CurrentRequestUser                   string
	currentPlayerPk                      *packet.AddPlayer
	currentPlayerKit                     defines.PlayerKit
	woodAxeRTID                          int32
	woodAxeOn                            bool
	currentPos                           mgl32.Vec3
	currentYaw, currentPitch             float32
	esp                                  float64
	nan                                  float64
	selectIndicatorPos, areaIndicatorPos define.CubePos
	selectIndicateStructureBlock         *StructureBlock
	areaIndicateStructureBlock           *StructureBlock
	lastSeeTick                          int
	actionManager                        *ActionManager
	actionsChan                          chan func()
	actionsOccupied
	selectInfo
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
			Size:                      protocol.BlockPos{0, 0, 0},
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
	o.basicPacket.ShowBoundingBox = true
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

func (o *WoodAxe) Init(cfg *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
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
	o.actionsChan = make(chan func(), 1024)
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

func (o *WoodAxe) onInitStructureBlock(selectIndicatorPos, areaIndicatorPos define.CubePos) {
	uid := o.Frame.GetUQHolder().BotUniqueID
	o.selectIndicateStructureBlock = NewStructureBlock(
		selectIndicatorPos,
		func(p packet.Packet) {
			o.Frame.GetGameControl().SendMCPacket(p)
		},
		uid,
	)
	o.selectIndicateStructureBlock.OffBound()
	o.areaIndicateStructureBlock = NewStructureBlock(
		areaIndicatorPos,
		func(p packet.Packet) {
			o.Frame.GetGameControl().SendMCPacket(p)
		},
		uid,
	)
	o.areaIndicateStructureBlock.OffBound()
}

func (o *WoodAxe) InitStructureBlock() {
	o.Frame.GetGameControl().SendCmdAndInvokeOnResponse("setblock ~~-2~ air", func(output *packet.CommandOutput) {
		o.Frame.GetGameControl().SendCmd("setblock ~~-1~ structure_block")
		o.Frame.GetGameControl().SendCmdAndInvokeOnResponse("setblock ~~-2~ structure_block", func(output *packet.CommandOutput) {
			if output.SuccessCount > 0 {
				respData := cmdRespDataSet{}
				if err := json.Unmarshal([]byte(output.DataSet), &respData); err == nil {
					cubePos := define.CubePos{respData.Pos.X, respData.Pos.Y, respData.Pos.Z}
					o.onInitStructureBlock(cubePos, cubePos.Add(define.CubePos{0, 1, 0}))
					return
				}
			}
			o.Frame.GetGameControl().SayTo("@a", "小木斧指示器初始化失败")
		})
	})
}

func (o *WoodAxe) InitWorkSpace() {
	o.InitStructureBlock()
	o.selectInfo = selectInfo{selected: make(map[uint8]bool), pos: make(map[uint8]define.CubePos), currentSelectID: NotSelect}
	o.actionManager = NewActionManager("omwa", o.Frame.GetGameControl().SendCmd, o.actionsChan)
	o.currentPlayerKit.Say("小木斧初始化完成")
	o.currentPlayerKit.Say("输入 帮助 获得可用操作")
}

func (o *WoodAxe) onInitWorkSapce(pk *packet.AddPlayer) {
	o.currentPlayerPk = pk
	o.currentPlayerKit = o.Frame.GetGameControl().GetPlayerKit(pk.Username)
	o.currentPlayerKit.Say("建筑师 " + pk.Username + " 已进入小木斧范围内")
	o.onSeeItem(pk.HeldItem.Stack.NetworkID)
	o.InitWorkSpace()
}

func (o *WoodAxe) CleanUpWorkSpace() {
	if o.chanCreated {
		o.chanCreated = false
		close(o.releaseChan)
	}
	pk := o.currentPlayerKit
	o.currentPlayerPk = nil
	o.currentPlayerKit = nil
	o.selectIndicateStructureBlock = nil
	o.areaIndicateStructureBlock = nil
	o.selectInfo = selectInfo{selected: make(map[uint8]bool), pos: make(map[uint8]define.CubePos), currentSelectID: NotSelect}
	o.actionManager = nil
	pk.Say("小木斧工作区已关闭")
}

func (o *WoodAxe) onAddPlayer(pkt packet.Packet) {
	pk := pkt.(*packet.AddPlayer)
	if pk.Username == o.CurrentRequestUser {
		o.onInitWorkSapce(pk)
	}
}

func (o *WoodAxe) onAnimate(pkt packet.Packet) {
	pk := pkt.(*packet.Animate)
	if o.currentPlayerKit != nil && o.currentPlayerKit.GetRelatedUQ().Entity != nil && pk.EntityRuntimeID == o.currentPlayerKit.GetRelatedUQ().Entity.RuntimeID {
		if o.woodAxeOn {
			o.onPosInput()
		}
	}
}

func (o *WoodAxe) onSeeMobItem(pk *packet.MobEquipment) {
	if o.currentPlayerKit == nil || o.currentPlayerKit.GetRelatedUQ().Entity == nil || pk.EntityRuntimeID != o.currentPlayerKit.GetRelatedUQ().Entity.RuntimeID {
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
	nextBlocks := o.computeNextXPos(o.currentPos, lookAt, 30)
	selected := false
	var selectedBlockPos define.CubePos
	t := time.NewTicker(time.Millisecond * 10)
	selectedI := 30
	selectedBlockName := ""
	arrived := 0
	go func() {
		for i, pos := range nextBlocks {
			_pos := define.CubePos{pos[0], pos[1], pos[2]}
			_i := i
			utils.GetBlockAt(o.Frame.GetGameControl(), fmt.Sprintf("%v %v %v", _pos[0], _pos[1], _pos[2]),
				func(outOfWorld bool, isAir bool, name string, realPos define.CubePos) {
					// fmt.Println(_i, outOfWorld, isAir, name, realPos)
					if _i > arrived {
						arrived = _i
					}
					if (!outOfWorld) && (!isAir) {
						if _i < selectedI {
							selectedBlockPos = _pos
							selectedI = _i
							selected = true
							// fmt.Println("selected!")
							selectedBlockName = name
						}
					}
				})
			if !selected {
				<-t.C
			} else {
				break
			}
		}
		if !selected {
			for {
				if arrived >= len(nextBlocks)-2 {
					break
				}
				<-t.C
			}
		}
		if selected {
			o.currentPlayerKit.Say(fmt.Sprintf("§l§b选中 %v @ %v", selectedBlockName, selectedBlockPos))
			o.selectIndicateStructureBlock.IndicateCube(selectedBlockPos, selectedBlockPos)
			o.selectInfo.currentSelectID = o.selectInfo.nextSelect
			o.selectInfo.currentSelectPos = selectedBlockPos
			o.onUpdateSelectPos()
		} else {
			o.currentPlayerKit.Say(fmt.Sprintf("§l§a未选中方块！"))
		}
	}()
}

func (o *WoodAxe) onUpdateSelectPos() {
	o.selectInfo.selected[o.currentSelectID] = true
	o.selectInfo.pos[o.currentSelectID] = o.currentSelectPos
	switch o.selectInfo.currentSelectID {
	case AreaPosOne:
		o.currentPlayerKit.Say("已选中区域起点 1")
		o.areaIndicateStructureBlock.IndicateCube(o.selectInfo.pos[AreaPosOne], o.selectInfo.pos[AreaPosOne])
	case AreaPosTwo:
		o.currentPlayerKit.Say("已选中区域起点 2")
		o.areaIndicateStructureBlock.IndicateCube(o.selectInfo.pos[AreaPosOne], o.selectInfo.pos[AreaPosTwo])
	case BasePointOne:
		o.currentPlayerKit.Say("已选中区域基准点")
	case BasePointTwo:
		o.currentPlayerKit.Say("已选中目标基准点")
	}
	if o.selectInfo.triggerFN == nil {
		o.selectInfo.nextSelect++
		if o.selectInfo.nextSelect == BasePointTwo {
			o.selectInfo.nextSelect = AreaPosOne
			o.currentPlayerKit.Say("输入 帮助 以打开小木斧帮助菜单")
		}
	} else {
		o.selectInfo.triggerFN()
	}
}

func (o *WoodAxe) getPosString() string {
	return fmt.Sprintf("Pos: [%.1f, %.1f, %.1f] Angle: [%.1f, %.1f]", o.currentPos[0], o.currentPos[1], o.currentPos[2], o.currentYaw, o.currentPitch)
}

func (o *WoodAxe) onPlayerMove(pk *packet.MovePlayer) {
	if o.currentPlayerKit != nil && o.currentPlayerKit.GetRelatedUQ().Entity != nil && pk.EntityRuntimeID == o.currentPlayerKit.GetRelatedUQ().Entity.RuntimeID {
		o.currentPos = pk.Position
		o.currentYaw = pk.HeadYaw
		o.currentPitch = pk.Pitch
		o.lastSeeTick = int(pk.Tick)
	}
}

func (o *WoodAxe) BlockUpdate(pos define.CubePos, origRTID uint32, currentRTID uint32) {
	// orig, _ := chunk.RuntimeIDToBlock(origRTID)
	// current, _ := chunk.RuntimeIDToBlock(currentRTID)
	// hint := fmt.Sprintf("%v:%v->%v", pos, strings.ReplaceAll(orig.Name, "minecraft:", ""), strings.ReplaceAll(current.Name, "minecraft:", ""))
	// fmt.Println(hint)
	// o.currentPlayerKit.ActionBar(hint)
}

func (o *WoodAxe) renderMenu() map[string]func(chat *defines.GameChat) {
	actions := map[string]func(chat *defines.GameChat){}
	hints := []string{"输入 初始化 以重置小木斧工作区", "输入 完成 以释放机器人，使之回到原本的任务中"}
	if action, trigger, hint, available := copyEntry(o); available {
		hints = append(hints, hint)
		actions[trigger] = action
	}
	if action, trigger, hint, available := continuousCopyEntry(o); available {
		hints = append(hints, hint)
		actions[trigger] = action
	}
	if action, trigger, hint, available := undoEntry(o); available {
		hints = append(hints, hint)
		actions[trigger] = action
	}
	if action, trigger, hint, available := redoEntry(o); available {
		hints = append(hints, hint)
		actions[trigger] = action
	}
	if action, trigger, hint, available := doneUndoEntry(o); available {
		hints = append(hints, hint)
		actions[trigger] = action
	}
	if o.UseLargeFill {
		if action, trigger, hint, available := largeFillEntry(o); available {
			hints = append(hints, hint)
			actions[trigger] = action
		}
	}
	if action, trigger, hint, available := fillEntry(o); available {
		hints = append(hints, hint)
		actions[trigger] = action
	}
	if action, trigger, hint, available := replaceEntry(o); available {
		hints = append(hints, hint)
		actions[trigger] = action
	}
	if action, trigger, hint, available := moveEntry(o); available {
		hints = append(hints, hint)
		actions[trigger] = action
	}
	if action, trigger, hint, available := flipEntry(o); available {
		hints = append(hints, hint)
		actions[trigger] = action
	}
	actions["完成"] = func(chat *defines.GameChat) {
		o.CleanUpWorkSpace()
	}
	actions["帮助"] = func(chat *defines.GameChat) {
		for _, hint := range hints {
			o.currentPlayerKit.Say(hint)
		}
	}
	actions["初始化"] = func(chat *defines.GameChat) {
		o.InitWorkSpace()
	}
	return actions
}

func (o *WoodAxe) onChat(chat *defines.GameChat) (stop bool) {
	if o.CurrentRequestUser == chat.Name {
		if len(chat.Msg) > 0 {
			menu := o.renderMenu()
			msg := chat.Msg[0]
			if action, hasK := menu[msg]; hasK {
				chat.Msg = chat.Msg[1:]
				action(chat)
				return true
			}
		}
	}
	return false
}

func (o *WoodAxe) onTrigger(chat *defines.GameChat) (stop bool) {
	flag := false
	for _, name := range o.Operators {
		if name == chat.Name {
			flag = true
		}
	}
	if !flag {
		o.Frame.GetGameControl().SayTo(chat.Name, "你没有权限使用这个功能")
		return true
	} else {
		if o.CurrentRequestUser == chat.Name {
			o.Frame.GetGameControl().SayTo(chat.Name, "你已经在使用这个功能了， 或许你想说 初始化 ?")
			return true
		}

		o.Frame.GetBotTaskScheduler().CommitNormalTask(&defines.BasicBotTaskPauseAble{
			BasicBotTask: defines.BasicBotTask{
				Name: fmt.Sprintf("小木斧 " + chat.Name),
				ActivateFn: func() {
					o.CurrentRequestUser = chat.Name
					if !o.chanCreated {
						o.releaseChan = make(chan struct{})
						o.chanCreated = true
					}
					o.currentPlayerKit = o.Frame.GetGameControl().GetPlayerKit(chat.Name)
					o.Frame.GetGameControl().SendCmd(fmt.Sprintf("tp @s \"%v\"", chat.Name))
					time.Sleep(time.Second * 1)
					o.InitWorkSpace()
					o.Frame.GetGameControl().SendCmd(fmt.Sprintf("tp @s ~500 ~ ~"))
					time.Sleep(time.Second * 3)
					o.Frame.GetGameControl().SendCmd(fmt.Sprintf("tp @s ~-500 ~ ~"))
					<-o.releaseChan
				},
			},
		})
	}
	return true
}

func (o *WoodAxe) Inject(frame defines.MainFrame) {
	o.Frame = frame
	frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDAddPlayer, o.onAddPlayer)
	frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDAnimate, o.onAnimate)
	o.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDMobEquipment, func(p packet.Packet) {
		o.onSeeMobItem(p.(*packet.MobEquipment))
	})
	o.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDMovePlayer, func(p packet.Packet) {
		o.onPlayerMove(p.(*packet.MovePlayer))
	})
	o.Frame.GetGameListener().AppendOnBlockUpdateInfoCallBack(o.BlockUpdate)
	o.Frame.GetGameListener().SetGameChatInterceptor(o.onChat)
	// frame.GetGameListener().SetOnAnyPacketCallBack(o.onAnyPacket)
	o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: "",
			FinalTrigger: false,
			Usage:        "启用小木斧帮助编辑建筑和结构",
		},
		OptionalOnTriggerFn: o.onTrigger,
	})
}

func (o *WoodAxe) Activate() {
	go func() {
		for action := range o.actionsChan {
			action()
		}
	}()
	// lastCheckTick := 0
	// go func() {
	// 	tick := time.NewTicker(time.Second * 3)
	// 	for {
	// 		<-tick.C
	// 		if o.currentPlayerPk != nil {
	// 			if lastCheckTick != 0 {
	// 				if o.lastSeeTick-lastCheckTick < 20 {
	// 					o.CleanUpWorkSpace()
	// 				}
	// 			}
	// 			lastCheckTick = o.lastSeeTick
	// 		}
	// 	}
	// }()
}

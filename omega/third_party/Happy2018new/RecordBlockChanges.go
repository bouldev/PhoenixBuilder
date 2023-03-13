package Happy2018new

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"phoenixbuilder/fastbuilder/mcstructure"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/omega/defines"
	Happy2018new_depends "phoenixbuilder/omega/third_party/Happy2018new/depends"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

type RecordBlockChanges struct {
	*defines.BasicComponent
	ListenPacketUpdateBlock     bool     `json:"是否监听 packet.UpdateBlock(21号) 包"`
	ListenPacketBlockActor      bool     `json:"是否监听 packet.BlockActor(56号) 包"`
	DiscardUnknwonOperator      bool     `json:"丢弃未知操作来源的方块"`
	OutputJsonDatasAndCloseThis bool     `json:"下次启动本组件时统计日志为 JSON 形式然后关闭组件"`
	OutputToCMD                 bool     `json:"在控制台实时打印方块变动记录"`
	OnlyRecordList              []string `json:"只记录下列方块(为空时将记录任何方块)"`
	MaxCountToRecord            int      `json:"允许的最大日志数(填 -1 则跳过检查)"`
	MaxPlayerRecord             int      `json:"每次至多追踪的玩家数"`
	TrackingRadius              float64  `json:"追踪半径"`
	FileName                    string   `json:"文件名称"`
	TimeToSaveChanges           int      `json:"文件保存频率(单位为分钟, 需填写整数)"`
	OnlyRecordMap               map[string]bool
	DataReceived                []struct {
		Time               string
		BlockPos           [3]int32
		BlockName_Result   string
		BlockStates_Result string
		BlockNBT           string
		Situation          uint32
		Operator           []string
	}
	StartLengthOfDataReceived int
}

func (o *RecordBlockChanges) InitExcludeMap() {
	o.OnlyRecordMap = map[string]bool{}
	// prepare
	for _, value := range o.OnlyRecordList {
		o.OnlyRecordMap[value] = true
	}
}

func (o *RecordBlockChanges) BeSureThatDiscardOperator(blockName string) bool {
	if len(o.OnlyRecordMap) > 0 {
		_, ok := o.OnlyRecordMap[blockName]
		return ok
	}
	return true
}

func (o *RecordBlockChanges) RequestBlockChangesInfo(BlockInfo packet.UpdateBlock, BlockNBT map[string]interface{}) {
	var blockName_Result string
	var blockStates_Result string
	var resp packet.CommandOutput
	var operator []string = []string{}
	var stringNBT string = "undefined"
	var err error
	// prepare
	if BlockNBT != nil {
		stringNBT, err = mcstructure.ConvertCompoundToString(BlockNBT, false)
		if err != nil {
			stringNBT = "undefined"
		}
	}
	// parse block nbt to string nbt
	if BlockInfo.Flags != 32768 {
		singleBlock, found := chunk.RuntimeIDToBlock(chunk.NEMCRuntimeIDToStandardRuntimeID(BlockInfo.NewBlockRuntimeID))
		if found {
			blockName_Result = singleBlock.Name
		} else {
			blockName_Result = "unknown"
		}
		// get block name
		blockStates_Result, err = mcstructure.ConvertCompoundToString(singleBlock.Properties, true)
		if err != nil {
			blockStates_Result = "undefined"
		}
		// get block states
	} else {
		blockName_Result = "unknown"
		blockStates_Result = "undefined"
		BlockInfo.Flags = 0
		// if the packet is packet.BlockActorData
	}
	// get basic info
	if !o.BeSureThatDiscardOperator(blockName_Result) {
		return
	}
	// if need to discard this operation
	resp = packet.CommandOutput{}
	o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(
		fmt.Sprintf("testfor @a[c=%v,x=%v,y=%v,z=%v,r=%v,name=!\"%v\"]", o.MaxPlayerRecord, BlockInfo.Position.X(), BlockInfo.Position.Y(), BlockInfo.Position.Z(), o.TrackingRadius, o.Frame.GetUQHolder().GetBotName()),
		func(output *packet.CommandOutput) {
			resp = *output
			if resp.SuccessCount > 0 {
				operator = strings.Split(resp.OutputMessages[0].Parameters[0], ", ")
			} else {
				operator = []string{"unknown"}
			}
			if o.DiscardUnknwonOperator && resp.SuccessCount > 0 {
				o.DataReceived = append(o.DataReceived, struct {
					Time               string
					BlockPos           [3]int32
					BlockName_Result   string
					BlockStates_Result string
					BlockNBT           string
					Situation          uint32
					Operator           []string
				}{
					Time:               time.Now().Format("2006-01-02 15:04:05"),
					BlockPos:           BlockInfo.Position,
					BlockName_Result:   blockName_Result,
					BlockStates_Result: blockStates_Result,
					BlockNBT:           stringNBT,
					Situation:          BlockInfo.Flags,
					Operator:           operator,
				})
			}
			if !o.DiscardUnknwonOperator {
				o.DataReceived = append(o.DataReceived, struct {
					Time               string
					BlockPos           [3]int32
					BlockName_Result   string
					BlockStates_Result string
					BlockNBT           string
					Situation          uint32
					Operator           []string
				}{
					Time:               time.Now().Format("2006-01-02 15:04:05"),
					BlockPos:           BlockInfo.Position,
					BlockName_Result:   blockName_Result,
					BlockStates_Result: blockStates_Result,
					BlockNBT:           stringNBT,
					Situation:          BlockInfo.Flags,
					Operator:           operator,
				})
			}
			if o.OutputToCMD && o.DiscardUnknwonOperator && resp.SuccessCount > 0 {
				value := o.DataReceived[len(o.DataReceived)-1]
				pterm.Info.Printf("记录方块改动日志: (%v,%v,%v) 处的方块有更新，内容如下\n", BlockInfo.Position.X(), BlockInfo.Position.Y(), BlockInfo.Position.Z())
				pterm.Info.Printf("操作时间: %v | 关联的方块名: %v | 关联的方块状态: %v | 关联的 NBT 数据: %v | 可能的操作者: %v | 附加数据: %v\n", value.Time, value.BlockName_Result, value.BlockStates_Result, value.BlockNBT, value.Operator, value.Situation)
			}
			if o.OutputToCMD && !o.DiscardUnknwonOperator {
				value := o.DataReceived[len(o.DataReceived)-1]
				pterm.Info.Printf("记录方块改动日志: (%v,%v,%v) 处的方块有更新，内容如下\n", BlockInfo.Position.X(), BlockInfo.Position.Y(), BlockInfo.Position.Z())
				pterm.Info.Printf("操作时间: %v | 关联的方块名: %v | 关联的方块状态: %v | 关联的 NBT 数据: %v | 可能的操作者: %v | 附加数据: %v\n", value.Time, value.BlockName_Result, value.BlockStates_Result, value.BlockNBT, value.Operator, value.Situation)
			}
		},
	)
}

func (o *RecordBlockChanges) OutputDatas() []byte {
	ans := []byte{}
	// prepare
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, uint32(len(o.DataReceived)))
	ans = append(ans, buf.Bytes()...)
	// data length
	for _, value := range o.DataReceived {
		ans = append(ans, uint8(len(value.Time)))
		ans = append(ans, []byte(value.Time)...)
		// time
		for _, val := range value.BlockPos {
			buf := bytes.NewBuffer([]byte{})
			binary.Write(buf, binary.BigEndian, val)
			ans = append(ans, buf.Bytes()...)
		}
		// pos
		ans = append(ans, uint8(len(value.BlockName_Result)))
		ans = append(ans, []byte(value.BlockName_Result)...)
		// blockName_Result
		buf = bytes.NewBuffer([]byte{})
		binary.Write(buf, binary.BigEndian, uint16(len([]byte(value.BlockStates_Result))))
		ans = append(ans, buf.Bytes()...)
		ans = append(ans, []byte(value.BlockStates_Result)...)
		// blockStates_Result
		buf = bytes.NewBuffer([]byte{})
		binary.Write(buf, binary.BigEndian, uint32(len([]byte(value.BlockNBT))))
		ans = append(ans, buf.Bytes()...)
		ans = append(ans, []byte(value.BlockNBT)...)
		// blockNBT
		buf = bytes.NewBuffer([]byte{})
		binary.Write(buf, binary.BigEndian, value.Situation)
		ans = append(ans, buf.Bytes()...)
		// situation
		ans = append(ans, uint8(len(value.Operator)))
		for _, val := range value.Operator {
			ans = append(ans, uint8(len(val)))
			ans = append(ans, []byte(val)...)
		}
		// operator
	}
	return Happy2018new_depends.Compress(ans)
}

func (o *RecordBlockChanges) GetDatas() {
	ans := []struct {
		Time               string
		BlockPos           [3]int32
		BlockName_Result   string
		BlockStates_Result string
		BlockNBT           string
		Situation          uint32
		Operator           []string
	}{}
	got, err := o.Frame.GetFileData(o.FileName)
	if len(got) <= 0 || err != nil {
		o.DataReceived = ans
		return
	}
	// prepare
	current := Happy2018new_depends.Decompress(got)
	// decompress
	reader := bytes.NewReader(current)
	p := make([]byte, 4)
	n, err := reader.Read(p)
	if n < 4 || err != nil {
		panic("无法读取保存的文件，请检查您的文件是否已经损坏！")
	}
	// get length
	buf := bytes.NewBuffer(p)
	var length uint32
	binary.Read(buf, binary.BigEndian, &length)
	// decode length
	if int(length) > o.MaxCountToRecord && o.MaxCountToRecord != -1 {
		panic(fmt.Sprintf("当前日志可能过大，现在已经记录了 %v 条日志，而配置中最多允许出现 %v 条日志", length, o.MaxCountToRecord))
	}
	// 如果超过最大记录数量，就报错处理
	for i := 0; i < int(length); i++ {
		timeLength, err := reader.ReadByte()
		if err != nil {
			panic("无法读取保存的文件，请检查您的文件是否已经损坏！")
		}
		// get length of time
		p = make([]byte, timeLength)
		n, err = reader.Read(p)
		if n < int(timeLength) || err != nil {
			panic("无法读取保存的文件，请检查您的文件是否已经损坏！")
		}
		time := string(p)
		// time
		pos := [3]int32{}
		for j := 0; j < 3; j++ {
			p = make([]byte, 4)
			n, err = reader.Read(p)
			if n < 4 || err != nil {
				panic("无法读取保存的文件，请检查您的文件是否已经损坏！")
			}
			// get pos[j]
			buf = bytes.NewBuffer(p)
			var posSingle int32
			binary.Read(buf, binary.BigEndian, &posSingle)
			// decode pos[j]
			pos[j] = posSingle
		}
		// blockPos
		blockName_Result_length, err := reader.ReadByte()
		if err != nil {
			panic("无法读取保存的文件，请检查您的文件是否已经损坏！")
		}
		// get length of blockName_Result
		p = make([]byte, blockName_Result_length)
		n, err = reader.Read(p)
		if n < int(blockName_Result_length) || err != nil {
			panic("无法读取保存的文件，请检查您的文件是否已经损坏！")
		}
		blockName_Result := string(p)
		// blockName_Result
		p = make([]byte, 2)
		n, err = reader.Read(p)
		if n < 2 || err != nil {
			panic("无法读取保存的文件，请检查您的文件是否已经损坏！")
		}
		// get length of blockStates_Result
		buf = bytes.NewBuffer(p)
		var blockStates_Result_length uint16
		binary.Read(buf, binary.BigEndian, &blockStates_Result_length)
		// decode length of blockStates_Result
		p = make([]byte, blockStates_Result_length)
		n, err = reader.Read(p)
		if n < int(blockStates_Result_length) || err != nil {
			panic("无法读取保存的文件，请检查您的文件是否已经损坏！")
		}
		blockStates_Result := string(p)
		// blockStates_Result
		p = make([]byte, 4)
		n, err = reader.Read(p)
		if n < 4 || err != nil {
			panic("无法读取保存的文件，请检查您的文件是否已经损坏！")
		}
		// get length of blockNBT
		buf = bytes.NewBuffer(p)
		var blockNBT_length uint32
		binary.Read(buf, binary.BigEndian, &blockNBT_length)
		// decode length of blockNBT
		p = make([]byte, blockNBT_length)
		n, err = reader.Read(p)
		if n < int(blockNBT_length) || err != nil {
			panic("无法读取保存的文件，请检查您的文件是否已经损坏！")
		}
		blockNBT := string(p)
		// blockNBT
		p = make([]byte, 4)
		n, err = reader.Read(p)
		if n < 4 || err != nil {
			panic("无法读取保存的文件，请检查您的文件是否已经损坏！")
		}
		// get situation
		buf = bytes.NewBuffer(p)
		var situation uint32
		binary.Read(buf, binary.BigEndian, &situation)
		// decode situation
		operatorLength, err := reader.ReadByte()
		if err != nil {
			panic("无法读取保存的文件，请检查您的文件是否已经损坏！")
		}
		// get length of operator
		operator := []string{}
		for j := 0; j < int(operatorLength); j++ {
			operatorSingleLength, err := reader.ReadByte()
			if err != nil {
				panic("无法读取保存的文件，请检查您的文件是否已经损坏！")
			}
			// get length of operator(single)
			p = make([]byte, operatorSingleLength)
			n, err = reader.Read(p)
			if n < int(operatorSingleLength) || err != nil {
				panic("无法读取保存的文件，请检查您的文件是否已经损坏！")
			}
			operator = append(operator, string(p))
			// operator(single)
		}
		// operator
		ans = append(ans, struct {
			Time               string
			BlockPos           [3]int32
			BlockName_Result   string
			BlockStates_Result string
			BlockNBT           string
			Situation          uint32
			Operator           []string
		}{
			Time:               time,
			BlockPos:           pos,
			BlockName_Result:   blockName_Result,
			BlockStates_Result: blockStates_Result,
			BlockNBT:           blockNBT,
			Situation:          situation,
			Operator:           operator,
		})
	}
	o.DataReceived = ans
}

func (o *RecordBlockChanges) StatisticsDatas() {
	type blockCube struct {
		Posx int32
		Posy int32
		Posz int32
	}
	type single struct {
		Time               string
		BlockName_Result   string
		BlockStates_Result string
		BlockNBT           string
		Situation          uint32
		Operator           []string
	}
	type set []single
	// prepare
	blockCubeMap := map[blockCube]set{}
	for _, value := range o.DataReceived {
		got, ok := blockCubeMap[blockCube{value.BlockPos[0], value.BlockPos[1], value.BlockPos[2]}]
		if !ok {
			blockCubeMap[blockCube{value.BlockPos[0], value.BlockPos[1], value.BlockPos[2]}] = set{
				single{
					Time:               value.Time,
					BlockName_Result:   value.BlockName_Result,
					BlockStates_Result: value.BlockStates_Result,
					BlockNBT:           value.BlockNBT,
					Situation:          value.Situation,
					Operator:           value.Operator,
				},
			}
		} else {
			got = append(got, single{
				Time:               value.Time,
				BlockName_Result:   value.BlockName_Result,
				BlockStates_Result: value.BlockStates_Result,
				BlockNBT:           value.BlockNBT,
				Situation:          value.Situation,
				Operator:           value.Operator,
			})
			blockCubeMap[blockCube{value.BlockPos[0], value.BlockPos[1], value.BlockPos[2]}] = got
		}
	}
	new := map[string]interface{}{}
	for key, value := range blockCubeMap {
		singleNew := []interface{}{}
		for _, val := range value {
			operatorNew := []interface{}{}
			for _, v := range val.Operator {
				operatorNew = append(operatorNew, v)
			}
			singleNew = append(singleNew, map[string]interface{}{
				"操作时间":       val.Time,
				"关联的方块名":     val.BlockName_Result,
				"关联的方块状态":    val.BlockStates_Result,
				"关联的 NBT 数据": val.BlockNBT,
				"附加数据":       float64(val.Situation),
				"可能的操作者":     operatorNew,
			})
		}
		new[fmt.Sprintf("方块 (%v,%v,%v)", key.Posx, key.Posy, key.Posz)] = singleNew
	}
	o.Frame.WriteJsonData(fmt.Sprintf("%v.json", o.FileName), new)
}

func (o *RecordBlockChanges) Init(settings *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
	marshal, _ := json.Marshal(settings.Configs)
	if err := json.Unmarshal(marshal, o); err != nil {
		panic(err)
	}
	if o.MaxPlayerRecord <= 0 {
		o.MaxPlayerRecord = 1
	}
	// init MaxPlayerRecord
	if o.TimeToSaveChanges <= 0 {
		o.TimeToSaveChanges = 1
	}
	// init RecordBlockChangesTime
	o.InitExcludeMap()
	// init ExcludeMap
}

func (o *RecordBlockChanges) Inject(frame defines.MainFrame) {
	o.Frame = frame
	if o.FileName == "" {
		o.FileName = ".Happy2018new"
	}
}

func (o *RecordBlockChanges) Activate() {
	o.GetDatas()
	o.StartLengthOfDataReceived = len(o.DataReceived)
	if o.OutputJsonDatasAndCloseThis {
		o.StatisticsDatas()
		pterm.Success.Printf("记录方块改动日志: 已成功将 %v 统计为 JSON 形式，保存为 %v.json\n", o.FileName, o.FileName)
		pterm.Info.Println("记录方块改动日志: 本组件已关闭")
		return
	}
	// get logs from file
	if o.ListenPacketUpdateBlock {
		o.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDUpdateBlock, func(p packet.Packet) {
			o.RequestBlockChangesInfo(*p.(*packet.UpdateBlock), nil)
		})
	}
	// listen packet.UpdateBlock
	if o.ListenPacketBlockActor {
		o.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDBlockActorData, func(p packet.Packet) {
			got := *p.(*packet.BlockActorData)
			o.RequestBlockChangesInfo(packet.UpdateBlock{
				Position: got.Position,
				Flags:    32768,
			}, got.NBTData)
		})
	}
	// listen packet.BlockActorData
	go func() {
		o.SaveChanges()
	}()
	// 定时保存
}

func (o *RecordBlockChanges) SaveChanges() {
	for {
		length := len(o.DataReceived)
		time.Sleep(time.Duration(o.TimeToSaveChanges) * time.Minute)
		if length != len(o.DataReceived) {
			err := o.Frame.WriteFileData(o.FileName, o.OutputDatas())
			if err != nil {
				panic(err)
			}
		}
	}
}

func (o *RecordBlockChanges) Stop() error {
	if !o.OutputJsonDatasAndCloseThis && o.StartLengthOfDataReceived != len(o.DataReceived) {
		fmt.Println("正在保存 " + o.FileName)
		return o.Frame.WriteFileData(o.FileName, o.OutputDatas())
	}
	return nil
}

//go:build !is_tweak
// +build !is_tweak

package special_tasks

import (
	"bytes"
	"fmt"
	"phoenixbuilder/fastbuilder/bdump"
	"phoenixbuilder/fastbuilder/configuration"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/parsing"
	"phoenixbuilder/fastbuilder/task"
	"phoenixbuilder/fastbuilder/task/fetcher"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/mirror/io/global"
	"phoenixbuilder/mirror/io/lru"
	"phoenixbuilder/mirror/io/world"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pterm/pterm"
)

func CreateExportTask(commandLine string, env *environment.PBEnvironment) *task.Task {
	cmdsender := env.CommandSender
	cfg, err := parsing.Parse(commandLine, configuration.GlobalFullConfig(env).Main())
	if err != nil {
		cmdsender.Output(fmt.Sprintf("Failed to parse command: %v", err))
		return nil
	}
	//cmdsender.Output("Sorry, but compatibility works haven't been done yet, please use lexport.")
	//return nil
	beginPos := cfg.Position
	endPos := cfg.End
	startX, endX, startZ, endZ := 0, 0, 0, 0
	if endPos.X-beginPos.X < 0 {
		temp := endPos.X
		endPos.X = beginPos.X
		beginPos.X = temp
	}
	startX, endX = beginPos.X, endPos.X
	if endPos.Y-beginPos.Y < 0 {
		temp := endPos.Y
		endPos.Y = beginPos.Y
		beginPos.Y = temp
	}
	if endPos.Z-beginPos.Z < 0 {
		temp := endPos.Z
		endPos.Z = beginPos.Z
		beginPos.Z = temp
	}
	startZ, endZ = beginPos.Z, endPos.Z
	hopPath, requiredChunks := fetcher.PlanHopSwapPath(startX, startZ, endX, endZ, 16)
	chunkPool := map[fetcher.ChunkPosDefine]fetcher.ChunkDefine{}
	memoryCacheFetcher := fetcher.CreateCacheHitFetcher(requiredChunks, chunkPool)
	env.LRUMemoryChunkCacher.(*lru.LRUMemoryChunkCacher).Iter(func(pos define.ChunkPos, chunk *mirror.ChunkData) (stop bool) {
		memoryCacheFetcher(fetcher.ChunkPosDefine{int(pos[0]) * 16, int(pos[1]) * 16}, fetcher.ChunkDefine(chunk))
		return false
	})
	hopPath = fetcher.SimplifyHopPos(hopPath)
	fmt.Println("Hop Left: ", len(hopPath))
	teleportFn := func(x, z int) {
		cmd := fmt.Sprintf("tp @s %v 128 %v", x, z)
		uid, _ := uuid.NewUUID()
		cmdsender.SendCommand(cmd, uid)
		cmd = fmt.Sprintf("execute @s ~~~ spreadplayers ~ ~ 3 4 @s")
		uid, _ = uuid.NewUUID()
		cmdsender.SendCommand(cmd, uid)
	}
	feedChan := make(chan *fetcher.ChunkDefineWithPos, 1024)
	deRegFn := env.ChunkFeeder.(*global.ChunkFeeder).RegNewReader(func(chunk *mirror.ChunkData) {
		feedChan <- &fetcher.ChunkDefineWithPos{Chunk: fetcher.ChunkDefine(chunk), Pos: fetcher.ChunkPosDefine{int(chunk.ChunkPos[0]) * 16, int(chunk.ChunkPos[1]) * 16}}
	})
	inHopping := true
	go func() {
		return
		yc := 23
		for {
			if !inHopping {
				break
			}
			uuidval, _ := uuid.NewUUID()
			yv := (yc-4)*16 + 8
			yc--
			if yc < 0 {
				yc = 23
			}
			cmdsender.SendCommand(fmt.Sprintf("tp @s ~ %d ~", yv), uuidval)
			time.Sleep(time.Millisecond * 50)
		}
	}()
	fmt.Println("Begin Fast Hopping")
	fetcher.FastHopper(teleportFn, feedChan, chunkPool, hopPath, requiredChunks, 0.5, 3)
	fmt.Println("Fast Hopping Done")
	deRegFn()
	hopPath = fetcher.SimplifyHopPos(hopPath)
	fmt.Println("Hop Left: ", len(hopPath))
	if len(hopPath) > 0 {
		fetcher.FixMissing(teleportFn, feedChan, chunkPool, hopPath, requiredChunks, 2, 3)
	}
	inHopping = false
	hasMissing := false
	for _, c := range requiredChunks {
		if !c.CachedMark {
			hasMissing = true
			pterm.Error.Printfln("Missing Chunk %v", c.Pos)
		}
	}
	if !hasMissing {
		pterm.Success.Println("all chunks successfully fetched!")
	}
	providerChunksMap := make(map[define.ChunkPos]*mirror.ChunkData)
	for _, chunk := range chunkPool {
		providerChunksMap[chunk.ChunkPos] = (*mirror.ChunkData)(chunk)
	}
	var offlineWorld *world.World
	offlineWorld = world.NewWorld(SimpleChunkProvider{providerChunksMap})

	go func() {
		defer func() {
			r := recover()
			if r != nil {
				debug.PrintStack()
				fmt.Println("go routine @ fastbuilder.task export crashed ", r)
			}
		}()
		cmdsender.Output("EXPORT >> Exporting...")
		V := (endPos.X - beginPos.X + 1) * (endPos.Y - beginPos.Y + 1) * (endPos.Z - beginPos.Z + 1)
		blocks := make([]*types.Module, V)
		counter := 0
		for x := beginPos.X; x <= endPos.X; x++ {
			for z := beginPos.Z; z <= endPos.Z; z++ {
				for y := beginPos.Y; y <= endPos.Y; y++ {
					runtimeId, item, found := offlineWorld.BlockWithNbt(define.CubePos{x, y, z})
					if !found {
						fmt.Printf("WARNING %d %d %d not found\n", x, y, z)
					}
					//block, item:=blk.EncodeBlock()
					block, static_item, _ := chunk.RuntimeIDToState(runtimeId)
					if block == "minecraft:air" {
						continue
					}
					var cbdata *types.CommandBlockData = nil
					var chestData *types.ChestData = nil
					var nbtData []byte = nil
					/*if(block=="chest"||block=="minecraft:chest"||strings.Contains(block,"shulker_box")) {
						content:=item["Items"].([]interface{})
						chest:=make(types.ChestData, len(content))
						for index, iface := range content {
							i:=iface.(map[string]interface{})
							name:=i["Name"].(string)
							count:=i["Count"].(uint8)
							damage:=i["Damage"].(int16)
							slot:=i["Slot"].(uint8)
							name_mcnk:=name[10:]
							chest[index]=types.ChestSlot {
								Name: name_mcnk,
								Count: count,
								Damage: uint16(int(damage)),
								Slot: slot,
							}
						}
						chestData=&chest
					}*/
					// TODO ^ Hope someone could help me to do that, just like what I did below ^
					if strings.Contains(block, "command_block") {
						/*
							=========
							Reference
							=========
							Types for command blocks are checked by their names
							Whether a command block is conditional is checked through its data value.
							SINCE IT IS NOT INCLUDED IN NBT DATA.

							The content of __tag is NBT data w/o keys, flatten placed,
							in such order:

							isMovable:byte
							CustomName:string
							UserCustomData:string
							powered:byte
							auto:byte
							conditionMet:byte
							LPConditionalMode:byte
							LPRedstoneMode:byte
							LPCommandMode:byte
							Command:string
							Version:VarInt32
							SuccessCount:VarInt32
							CustomName:string
							LastOutput:string
							LastOutputParams:list[string]
							TrackOutput:byte
							LastExecution:VarInt64
							TickDelay:VarInt32
							ExecuteOnFirstTick:byte
						*/
						__tag := []byte(item["__tag"].(string))
						//fmt.Printf("CMDBLK %#v\n\n",item["__tag"])
						var mode uint32
						if block == "command_block" || block == "minecraft:command_block" {
							mode = packet.CommandBlockImpulse
						} else if block == "repeating_command_block" || block == "minecraft:repeating_command_block" {
							mode = packet.CommandBlockRepeating
						} else if block == "chain_command_block" || block == "minecraft:chain_command_block" {
							mode = packet.CommandBlockChain
						}
						tagContent := bytes.NewBuffer(__tag)
						tagContent.Next(1)
						// ^ Skip: [isMovable:byte]
						_, err := readNBTString(tagContent)
						if err != nil {
							panic(err)
						}
						// ^ Skip: [CustomName:string]
						_, err = readNBTString(tagContent)
						if err != nil {
							panic(err)
						}
						// ^ Skip: [UserCustomData:string]
						tagContent.Next(1)
						// ^ Skip: [powered:byte]
						aut, err := tagContent.ReadByte()
						if err != nil {
							panic(err)
						}
						// ^ Read: [auto:byte]
						tagContent.Next(4)
						// ^ Skip: [conditionMet:byte]
						//   Skip: [LPConditionMode:byte]
						//   Skip: [LPRedstoneMode:byte]
						//   Skip: [LPCommandMode:byte]
						cmd, err := readNBTString(tagContent)
						if err != nil {
							panic(err)
						}
						// ^ Read: [Command:string]
						_, err = readVarint32(tagContent)
						if err != nil {
							panic(err)
						}
						// ^ Skip: [Version:VarInt32]
						_, err = readVarint32(tagContent)
						if err != nil {
							panic(err)
						}
						// ^ Skip: [SuccessCount:VarInt32]
						cusname, err := readNBTString(tagContent)
						if err != nil {
							panic(err)
						}
						// ^ Read: [CustomName:string]
						lo, err := readNBTString(tagContent)
						if err != nil {
							panic(err)
						}
						// ^ Read: [LastOutput:string]
						lop_in, err := readVarint32(tagContent)
						if err != nil {
							panic(err)
						}
						// ^ PartialRead: **LENGTH OF** [LastOutputParams:list[string]]
						for i := 0; i < int(lop_in); i++ {
							_, err = readNBTString(tagContent)
							if err != nil {
								panic(err)
							}
							// ^ PartialRead: **CONTENT OF** [LastOutputParams:list[string]]
						}
						// ^ Skip: [LastOutputParams:list[string]]
						trackoutput, err := tagContent.ReadByte()
						if err != nil {
							panic(err)
						}
						// ^ Read: [TrackOutput:byte]
						_, err = readVarint64(tagContent)
						if err != nil {
							panic(err)
						}
						// ^ Skip: [LastExecution:VarInt64]
						tickdelay, err := readVarint32(tagContent)
						if err != nil {
							panic(err)
						}
						// ^ Read: [TickDelay:VarInt32]
						exeft, err := tagContent.ReadByte()
						if err != nil {
							panic(err)
						}
						// ^ Read: [ExecuteOnFirstTick:byte]
						if tagContent.Len() != 0 {
							panic("Unterminated command block tag")
						}
						conb_bit := static_item["conditional_bit"].(uint8)
						conb := false
						if conb_bit == 1 {
							conb = true
						}
						var exeftb bool
						if exeft == 0 {
							exeftb = true
						} else {
							exeftb = true
						}
						var tob bool
						if trackoutput == 1 {
							tob = true
						} else {
							tob = false
						}
						var nrb bool
						if aut == 1 {
							nrb = false
							//REVERSED!!
						} else {
							nrb = true
						}
						cbdata = &types.CommandBlockData{
							Mode:               mode,
							Command:            cmd,
							CustomName:         cusname,
							ExecuteOnFirstTick: exeftb,
							LastOutput:         lo,
							TickDelay:          tickdelay,
							TrackOutput:        tob,
							Conditional:        conb,
							NeedsRedstone:      nrb,
						}
						//fmt.Printf("%#v\n",cbdata)
					} else {
						pnd, hasNBT := item["__tag"]
						if hasNBT {
							nbtData = []byte(pnd.(string))
						}
					}
					// it's ok to ignore "found", because it will set lb to air if not found
					lb, _ := chunk.RuntimeIDToLegacyBlock(runtimeId)
					blocks[counter] = &types.Module{
						Block: &types.Block{
							Name: &lb.Name,
							Data: uint16(lb.Val),
						},
						CommandBlockData: cbdata,
						ChestData:        chestData,
						DebugNBTData:     nbtData,
						Point: types.Position{
							X: x,
							Y: y,
							Z: z,
						},
					}
					counter++
				}
			}
		}
		blocks = blocks[:counter]
		runtime.GC()
		out := bdump.BDump{
			Blocks: blocks,
		}
		if strings.LastIndex(cfg.Path, ".bdx") != len(cfg.Path)-4 || len(cfg.Path) < 4 {
			cfg.Path += ".bdx"
		}
		cmdsender.Output("EXPORT >> Writing output file")
		err, signerr := out.WriteToFile(cfg.Path, env.LocalCert, env.LocalKey)
		if err != nil {
			cmdsender.Output(fmt.Sprintf("EXPORT >> ERROR: Failed to export: %v", err))
			return
		} else if signerr != nil {
			cmdsender.Output(fmt.Sprintf("EXPORT >> Note: The file is unsigned since the following error was trapped: %v", signerr))
		} else {
			cmdsender.Output(fmt.Sprintf("EXPORT >> File signed successfully"))
		}
		cmdsender.Output(fmt.Sprintf("EXPORT >> Successfully exported your structure to %v", cfg.Path))
		runtime.GC()
	}()
	return nil
}

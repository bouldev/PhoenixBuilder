package special_tasks

import (
	"fmt"
	"phoenixbuilder/fastbuilder/bdump"
	"phoenixbuilder/fastbuilder/configuration"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/mcstructure"
	"phoenixbuilder/fastbuilder/parsing"
	"phoenixbuilder/fastbuilder/task"
	"phoenixbuilder/fastbuilder/task/fetcher"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/minecraft/nbt"
	"phoenixbuilder/mirror"
	Blocks "phoenixbuilder/mirror/blocks"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/mirror/io/global"
	"phoenixbuilder/mirror/io/lru"
	"phoenixbuilder/mirror/io/world"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

func CreateExportTask(commandLine string, env *environment.PBEnvironment) *task.Task {
	cfg, err := parsing.Parse(commandLine, configuration.GlobalFullConfig(env).Main())
	if err != nil {
		env.GameInterface.Output(fmt.Sprintf("Failed to parse command: %v", err))
		return nil
	}
	//env.GameInterface.Output("Sorry, but compatibility works haven't been done yet, please use lexport.")
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
		env.GameInterface.SendCommand(cmd)
		cmd = fmt.Sprintf("execute as @s at @s run spreadplayers ~ ~ 3 4 @s")
		env.GameInterface.SendCommand(cmd)
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
			yv := (yc-4)*16 + 8
			yc--
			if yc < 0 {
				yc = 23
			}
			env.GameInterface.SendCommand(fmt.Sprintf("tp @s ~ %d ~", yv))
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
		env.GameInterface.Output("EXPORT >> Exporting...")
		V := (endPos.X - beginPos.X + 1) * (endPos.Y - beginPos.Y + 1) * (endPos.Z - beginPos.Z + 1)
		blocks := make([]*types.Module, V)
		counter := 0
		for x := beginPos.X; x <= endPos.X; x++ {
			for z := beginPos.Z; z <= endPos.Z; z++ {
				for y := beginPos.Y; y <= endPos.Y; y++ {
					var tagNBTData []byte
					var blockNBTBytes []byte
					runtimeId, blockNBTMap, found := offlineWorld.BlockWithNbt(define.CubePos{x, y, z})
					if !found {
						fmt.Printf("WARNING: %d %d %d not found\n", x, y, z)
					}
					blockName, blockStates, _ := Blocks.RuntimeIDToState(runtimeId)
					blockStatesString, err := mcstructure.MarshalBlockStates(blockStates)
					if err != nil {
						fmt.Printf("WARNING: Failed to marshal block states %#v; err = %v\n", blockStates, err)
					}
					pnd, hasNBT := blockNBTMap["__tag"].(string)
					if hasNBT {
						tagNBTData = []byte(pnd)
					}
					if len(blockNBTMap) > 0 {
						if strings.Contains(blockName, "command_block") {
							blockNBTMap["conditionalMode"] = blockStates["conditional_bit"].(byte)
						}
						blockNBTBytes, err = nbt.MarshalEncoding(blockNBTMap, nbt.LittleEndian)
						if err != nil {
							fmt.Printf("WARNING: Failed to marshal block NBT map %#v; err = %v\n", blockNBTMap, err)
						}
					}
					// it's ok to ignore "found", because it will set lb to air if not found
					blocks[counter] = &types.Module{
						Block: &types.Block{
							Name:        &blockName,
							BlockStates: blockStatesString,
						},
						DebugNBTData: tagNBTData,
						Point: types.Position{
							X: x,
							Y: y,
							Z: z,
						},
					}
					if len(blockNBTBytes) > 0 {
						blocks[counter].NBTData = blockNBTBytes
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
		env.GameInterface.Output("EXPORT >> Writing output file")
		err = out.WriteToFile(cfg.Path)
		if err != nil {
			env.GameInterface.Output(fmt.Sprintf("EXPORT >> ERROR: Failed to export: %v", err))
			return
		}
		env.GameInterface.Output(fmt.Sprintf("EXPORT >> Successfully exported your structure to %v", cfg.Path))
		runtime.GC()
	}()
	return nil
}

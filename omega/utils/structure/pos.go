package structure

import (
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/mirror/define"
	"sort"

	"github.com/pterm/pterm"
)

// func AlterImportPosStartAndSpeed(inChan chan *IOBlockForDecoder, offset define.CubePos, startFrom int, outChanLen int) (outChan chan *IOBlockForBuilder, stopFn func()) {
// 	outChan = make(chan *IOBlockForBuilder, outChanLen)
// 	stop := false
// 	go func() {
// 		counter := 0
// 		for {
// 			if stop {
// 				return
// 			}
// 			if counter < startFrom {
// 				counter++
// 				<-inChan
// 			} else {
// 				break
// 			}
// 		}
// 		for b := range inChan {
// 			if stop {
// 				return
// 			}
// 			b.Pos = b.Pos.Add(offset)
// 			if b.NBT != nil {
// 				delete(b.NBT, "x")
// 				delete(b.NBT, "y")
// 				delete(b.NBT, "z")
// 			}
// 			outChan <- b
// 		}
// 		close(outChan)
// 	}()
// 	return outChan, func() {
// 		stop = true
// 	}
// }

func AlterImportPosStartAndSpeedWithReArrangeOnce(inChan chan *IOBlockForDecoder, offset define.CubePos, startFrom int, outChanLen int, suggestMinCacheChunks int) (outChan chan *IOBlockForBuilder, stopFn func()) {
	outChan = make(chan *IOBlockForBuilder, outChanLen)
	stop := false
	air := chunk.AirRID
	counter := 0

	reArrangerToDumperChan := make(chan map[define.ChunkPos]*mirror.ChunkData, 2)

	// reArranger go routine
	go func() {
		chunks := make(map[define.ChunkPos]*mirror.ChunkData)
		lastChunkPos := define.ChunkPos{0, 0}
		lastChunk := &mirror.ChunkData{
			Chunk:       chunk.New(chunk.AirRID, define.Range{-64, 319}),
			BlockNbts:   make(map[define.CubePos]map[string]interface{}),
			BlockName:   make(map[define.CubePos]string), // for bdx
			BlockStates: make(map[define.CubePos]string), // for bdx
			BlockData:   make(map[define.CubePos]uint16), // for bdx
			ChunkPos:    lastChunkPos,
		}
		chunks[lastChunkPos] = lastChunk

		// define set block function
		setBlock := func(b *IOBlockForDecoder) {
			pos := b.Pos
			if pos.OutOfYBounds() {
				pterm.Warning.Printfln("位于 %v 的方块超出高度上限", pos)
				// Fast way out.
				return
			}
			chunkPos := define.ChunkPos{int32(pos[0] >> 4), int32(pos[2] >> 4)}
			if chunkPos != lastChunkPos {
				c, hasK := chunks[chunkPos]
				if !hasK {
					// chunk=&mirror.ChunkData{}
					c = &mirror.ChunkData{
						Chunk:       chunk.New(chunk.AirRID, define.Range{-64, 319}),
						BlockNbts:   make(map[define.CubePos]map[string]interface{}),
						BlockName:   make(map[define.CubePos]string), // for bdx
						BlockStates: make(map[define.CubePos]string), // for bdx
						BlockData:   make(map[define.CubePos]uint16), // for bdx
						ChunkPos:    chunkPos,
					}
					chunks[chunkPos] = c
				}
				lastChunk = c
				lastChunkPos = chunkPos
			}
			lastChunk.Chunk.SetBlock(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 0, b.RTID)
			if b.NBT != nil {
				lastChunk.BlockNbts[b.Pos] = b.NBT
			}
			if b.BlockName != "" && b.BlockStates == "" {
				lastChunk.BlockName[b.Pos] = b.BlockName
				lastChunk.BlockData[b.Pos] = b.BlockData
			} // for bdx
			if b.BlockName != "" && b.BlockStates != "" {
				lastChunk.BlockName[b.Pos] = b.BlockName
				lastChunk.BlockStates[b.Pos] = b.BlockStates
			} // for bdx
		}

		// do rearrange
		for b := range inChan {
			if stop {
				close(reArrangerToDumperChan)
				return
			}
			b.Pos = b.Pos.Add(offset)
			setBlock(b)
			if len(chunks) > suggestMinCacheChunks {
				reArrangerToDumperChan <- chunks
				chunks = make(map[define.ChunkPos]*mirror.ChunkData)
				lastChunkPos = define.ChunkPos{0, 0}
				lastChunk = &mirror.ChunkData{
					Chunk:       chunk.New(chunk.AirRID, define.Range{-64, 319}),
					BlockNbts:   make(map[define.CubePos]map[string]interface{}),
					BlockName:   make(map[define.CubePos]string), // for bdx
					BlockStates: make(map[define.CubePos]string), // for bdx
					BlockData:   make(map[define.CubePos]uint16), // for bdx
					ChunkPos:    lastChunkPos,
				}
			}
		}
		reArrangerToDumperChan <- chunks
		close(reArrangerToDumperChan)
	}()

	// dumper routine
	go func() {
		for chunks := range reArrangerToDumperChan {
			chunkXs := make([]int, 0)
			chunkZs := make([]int, 0)
			for chunkPos, _ := range chunks {
				chunkXs = append(chunkXs, int(chunkPos.X()))
				chunkZs = append(chunkZs, int(chunkPos.Z()))
			}
			cleanArr := func(in []int) []int {
				out := make([]int, 0)
				m := map[int]bool{}
				for _, i := range in {
					if _, hasK := m[i]; !hasK {
						m[i] = true
						out = append(out, i)
					}
				}
				return out
			}
			sort.Ints(chunkXs)
			sort.Ints(chunkZs)
			chunkXs = cleanArr(chunkXs)
			chunkZs = cleanArr(chunkZs)
			reOrderedChunks := make([]define.ChunkPos, 0)
			for i, chunkX := range chunkXs {
				if i%2 == 0 {
					for _, chunkZ := range chunkZs {
						p := define.ChunkPos{int32(chunkX), int32(chunkZ)}
						if _, hasK := chunks[p]; hasK {
							reOrderedChunks = append(reOrderedChunks, p)
						}
					}
				} else {
					for zi, _ := range chunkZs {
						chunkZ := chunkZs[len(chunkZs)-1-zi]
						p := define.ChunkPos{int32(chunkX), int32(chunkZ)}
						if _, hasK := chunks[p]; hasK {
							reOrderedChunks = append(reOrderedChunks, p)
						}
					}
				}
			}

			for _, chunkPos := range reOrderedChunks {
				chunk := chunks[chunkPos]
				// fmt.Println(chunkPos)
				nbts := chunk.BlockNbts
				blockName := chunk.BlockName     // for bdx
				blockStates := chunk.BlockStates // for bdx
				blockData := chunk.BlockData     // for bdx
				for subChunkI := int16(0); subChunkI < 24; subChunkI++ {
					subChunk := chunk.Chunk.Sub()[subChunkI]
					if subChunk.Empty() {
						continue
					}
					subChunkY := subChunkI*16 + int16(define.WorldRange[0])
					subChunkX := int(chunkPos[0]) * 16
					subChunkZ := int(chunkPos[1]) * 16
					subChunkStorage := subChunk.Layer(0)
					blk := subChunkStorage.At(0, 0, 0)
					// fmt.Println(subChunkStorage.Palette().Len(), subChunkStorage.IsPerIndexWithBitSizeUnder32Same(), blk)
					if subChunkStorage.Palette().Len() == 2 && subChunkStorage.IsPerIndexWithBitSizeUnder32Same() && blk != air {
						// fmt.Println("fast dump")
						p := define.CubePos{int(subChunkX), int(subChunkY), int(subChunkZ)}
						if counter < startFrom {
							counter += 16 * 16 * 16
						} else {
							if blockName[p] != "" && blockStates[p] != "" {
								outChan <- &IOBlockForBuilder{
									Pos:         p,
									RTID:        blk,
									Expand16:    true,
									BlockName:   blockName[p],   // for operation 13 which named `PlaceBlockWithBlockStates`
									BlockStates: blockStates[p], // for operation 13 which named `PlaceBlockWithBlockStates`
								}
							} else if blockName[p] != "" && blockStates[p] == "" {
								outChan <- &IOBlockForBuilder{
									Pos:       p,
									RTID:      blk,
									Expand16:  true,
									BlockName: blockName[p], // for operation 7 which named `PlaceBlock`
									BlockData: blockData[p], // for operation 7 which named `PlaceBlock`
								}
							} else {
								outChan <- &IOBlockForBuilder{
									Pos:      p,
									RTID:     blk,
									Expand16: true,
								}
							}
						}
						continue
					}
					for x := uint8(0); x < 16; x++ {
						for z := uint8(0); z < 16; z++ {
							for sy := uint8(0); sy < 16; sy++ {
								if stop {
									close(outChan)
									return
								}
								blk = subChunkStorage.At(x, sy, z)
								if blk == air {
									continue
								}
								p := define.CubePos{int(x) + subChunkX, int(subChunkY + int16(sy)), int(z) + subChunkZ}
								if counter < startFrom {
									counter++
									continue
								}
								if nbt, hasK := nbts[p]; hasK {
									if blockName[p] != "" && blockStates[p] != "" {
										outChan <- &IOBlockForBuilder{
											Pos:         p,
											RTID:        blk,
											NBT:         nbt,
											BlockName:   blockName[p],   // for operation 13 which named `PlaceBlockWithBlockStates`
											BlockStates: blockStates[p], // for operation 13 which named `PlaceBlockWithBlockStates`
										}
									} else if blockName[p] != "" && blockStates[p] == "" {
										outChan <- &IOBlockForBuilder{
											Pos:       p,
											RTID:      blk,
											NBT:       nbt,
											BlockName: blockName[p], // for operation 7 which named `PlaceBlock`
											BlockData: blockData[p], // for operation 7 which named `PlaceBlock`
										}
									} else {
										outChan <- &IOBlockForBuilder{
											Pos:  p,
											RTID: blk,
											NBT:  nbt,
										}
									}
								} else {
									if blockName[p] != "" && blockStates[p] != "" {
										outChan <- &IOBlockForBuilder{
											Pos:         p,
											RTID:        blk,
											BlockName:   blockName[p],   // for operation 13 which named `PlaceBlockWithBlockStates`
											BlockStates: blockStates[p], // for operation 13 which named `PlaceBlockWithBlockStates`
										}
									} else if blockName[p] != "" && blockStates[p] == "" {
										outChan <- &IOBlockForBuilder{
											Pos:       p,
											RTID:      blk,
											BlockName: blockName[p], // for operation 7 which named `PlaceBlock`
											BlockData: blockData[p], // for operation 7 which named `PlaceBlock`
										}
									} else {
										outChan <- &IOBlockForBuilder{
											Pos:  p,
											RTID: blk,
										}
									}
								}
							}
						}
					}
				}
			}
		}
		close(outChan)
	}()
	return outChan, func() {
		stop = true
	}
}

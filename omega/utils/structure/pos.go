package structure

import (
	"fmt"
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/mirror/define"
	"sort"
)

func AlterImportPosStartAndSpeed(inChan chan *IOBlock, offset define.CubePos, startFrom int, outChanLen int) (outChan chan *IOBlock, stopFn func()) {
	outChan = make(chan *IOBlock, outChanLen)
	stop := false
	go func() {
		counter := 0
		for {
			if stop {
				return
			}
			if counter < startFrom {
				counter++
				<-inChan
			} else {
				break
			}
		}
		for b := range inChan {
			if stop {
				return
			}
			b.Pos = b.Pos.Add(offset)
			if b.NBT != nil {
				delete(b.NBT, "x")
				delete(b.NBT, "y")
				delete(b.NBT, "z")
			}
			outChan <- b
		}
		close(outChan)
	}()
	return outChan, func() {
		stop = true
	}
}

func AlterImportPosStartAndSpeedWithReArrangeOnce(inChan chan *IOBlock, offset define.CubePos, startFrom int, outChanLen int, suggestMinCacheChunks int) (outChan chan *IOBlock, stopFn func()) {
	outChan = make(chan *IOBlock, outChanLen)
	stop := false
	air := chunk.AirRID
	counter := 0

	reArrangerToDumperChan := make(chan map[define.ChunkPos]*mirror.ChunkData, 2)

	// reArranger go routine
	go func() {
		chunks := make(map[define.ChunkPos]*mirror.ChunkData)
		lastChunkPos := define.ChunkPos{0, 0}
		lastChunk := &mirror.ChunkData{
			Chunk:     chunk.New(chunk.AirRID, define.Range{-64, 319}),
			BlockNbts: make(map[define.CubePos]map[string]interface{}),
			ChunkPos:  lastChunkPos,
		}
		chunks[lastChunkPos] = lastChunk

		// define set block function
		setBlock := func(b *IOBlock) {
			pos := b.Pos
			if pos.OutOfYBounds() {
				fmt.Println(pos)
				// Fast way out.
				return
			}
			chunkPos := define.ChunkPos{int32(pos[0] >> 4), int32(pos[2] >> 4)}
			if chunkPos != lastChunkPos {
				c, hasK := chunks[chunkPos]
				if !hasK {
					// chunk=&mirror.ChunkData{}
					c = &mirror.ChunkData{
						Chunk:     chunk.New(chunk.AirRID, define.Range{-64, 319}),
						BlockNbts: make(map[define.CubePos]map[string]interface{}),
						ChunkPos:  chunkPos,
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
					Chunk:     chunk.New(chunk.AirRID, define.Range{-64, 319}),
					BlockNbts: make(map[define.CubePos]map[string]interface{}),
					ChunkPos:  lastChunkPos,
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
				for subChunkI := int16(0); subChunkI < 24; subChunkI++ {
					subChunk := chunk.Chunk.Sub()[subChunkI]
					if subChunk.Empty() {
						continue
					}
					for x := uint8(0); x < 16; x++ {
						for z := uint8(0); z < 16; z++ {
							for sy := uint8(0); sy < 16; sy++ {
								if stop {
									close(outChan)
									return
								}
								y := subChunkI*16 + int16(sy) + int16(define.WorldRange[0])
								blk := subChunk.Block(x, sy, z, 0)
								if blk == air {
									continue
								}
								p := define.CubePos{int(x) + int(chunkPos[0])*16, int(y), int(z) + int(chunkPos[1])*16}
								if counter < startFrom {
									counter++
									continue
								}
								if nbt, hasK := nbts[p]; hasK {
									outChan <- &IOBlock{
										Pos:  p,
										RTID: blk,
										NBT:  nbt,
									}
								} else {
									outChan <- &IOBlock{
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
		close(outChan)
	}()
	return outChan, func() {
		stop = true
	}
}

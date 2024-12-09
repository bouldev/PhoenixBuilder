package fetcher

/*
 * This file is part of PhoenixBuilder.

 * PhoenixBuilder is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License.

 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.

 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.

 * Copyright (C) 2021-2025 Bouldev
 */

import (
	"fmt"
	"time"

	"github.com/pterm/pterm"
)

func doHop(
	teleportFn TeleportFn,chunkFeedChan ChunkFeedChan,
	chunkPool map[ChunkPosDefine]ChunkDefine,hopPoint *ExportHopPos,requiredChunks ExportedChunksMap,
	minWaitTime,maxWaitTime float32,
){
	teleportFn(hopPoint.Pos[0],hopPoint.Pos[1])
	maxTimer:=time.NewTimer(time.Duration(int(float32(time.Second)*maxWaitTime)))
	minTimer:=time.NewTimer(time.Duration(int(float32(time.Second)*minWaitTime)))
	allChunksHit:=false
	for{
		select{
		case <-minTimer.C:
			if allChunksHit{
				fmt.Println("no new chunk arrived in min hop time after last chunk arrived, quit hop point")
				return
			}
		case <-maxTimer.C:
			pterm.Warning.Println("Max hop time exceed, quit hop point")
			return
		case chunkWithPos:=<-chunkFeedChan:
			minTimer=time.NewTimer(time.Duration(int(float32(time.Second)*minWaitTime)))
			maxTimer=time.NewTimer(time.Duration(int(float32(time.Second)*maxWaitTime)))
			if _,hasK:=requiredChunks[chunkWithPos.Pos];hasK{
				chunkPool[chunkWithPos.Pos]=chunkWithPos.Chunk
				requiredChunks[chunkWithPos.Pos].CachedMark=true
				if !allChunksHit{
					_allHit:=true
					for _,c:=range hopPoint.LinkedChunk{
						if !c.CachedMark{
							_allHit=false
							break
						}
					}
					if _allHit{
						allChunksHit=true
					}
				}
			}
		}
	}
}

func FastHopper(
	teleportFn TeleportFn,chunkFeedChan ChunkFeedChan,
	chunkPool map[ChunkPosDefine]ChunkDefine,hopPath []*ExportHopPos,requiredChunks ExportedChunksMap,
	minWaitTime,maxWaitTime float32,
	){
	for _,hp:=range hopPath{
		fmt.Println("now hop to: ",hp.Pos)
		doHop(teleportFn,chunkFeedChan,chunkPool,hp,requiredChunks,minWaitTime,maxWaitTime)
	}
}

func FixMissing(
	teleportFn TeleportFn,chunkFeedChan ChunkFeedChan,
	chunkPool map[ChunkPosDefine]ChunkDefine,hopPath []*ExportHopPos,requiredChunks ExportedChunksMap,
	minWaitTime,maxWaitTime float32,
	){
	for round:=0;round<2;round++{
		if len(hopPath)==0{
			return
		}
		if round%2==0{
			teleportFn(12401,-12401)
		}else{
			teleportFn(-12401,12401)
		}
		time.Sleep(time.Second*2)
		for _,hp:=range hopPath{
			fmt.Println("now hop to ",hp.Pos," for missing fixing")
			doHop(teleportFn,chunkFeedChan,chunkPool,hp,requiredChunks,minWaitTime,maxWaitTime)
		}
		hopPath=SimplifyHopPos(hopPath)
	}
	for _,c:=range requiredChunks{
		if c.CachedMark{
			continue
		}
		tmpHop:=&ExportHopPos{Pos:c.Pos,LinkedChunk: []*ExportedChunkPos{c}}
		teleportFn(10000,-10000)
		time.Sleep(time.Second*2)
		fmt.Println("now tp to ",c.Pos," for chunk missing fixing")
		doHop(teleportFn,chunkFeedChan,chunkPool,tmpHop,requiredChunks,minWaitTime,maxWaitTime)
		if !c.CachedMark{
			teleportFn(-10000,10000)
			time.Sleep(time.Second*2)
			fmt.Println("now tp to ",c.Pos," for chunk missing fixing")
			doHop(teleportFn,chunkFeedChan,chunkPool,tmpHop,requiredChunks,minWaitTime,maxWaitTime)
		}
	}
}
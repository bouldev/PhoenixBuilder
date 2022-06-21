package woodaxe

import (
	"fmt"
	"math"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/omega/defines"
	"strconv"
	"time"
)

func copyEntry(o *WoodAxe) (action func(chat *defines.GameChat), actionName string, hint string, available bool) {
	if o.actionsOccupied.occupied {
		return nil, "", "", false
	}
	finalFunc := func() {
		start, end := sortPos(o.selectInfo.pos[AreaPosOne], o.selectInfo.pos[AreaPosTwo])
		offset := o.selectInfo.pos[BasePointTwo].Sub(o.selectInfo.pos[BasePointOne])
		size := end.Sub(start)
		target := start.Add(offset)
		o.actionManager.Commit(&Action{
			Do: func() {
				o.Frame.GetGameControl().SendCmd(fmt.Sprintf("clone %v %v %v %v %v %v %v %v %v",
					start[0], start[1], start[2], end[0], end[1], end[2], target[0], target[1], target[2]))
			},
			AffectAreas: [][2]define.CubePos{{target, target.Add(size)}},
		})

		o.selectInfo.nextSelect = AreaPosOne
		o.selectInfo.currentSelectID = NotSelect
		o.selectInfo.triggerFN = nil
	}
	activateFinalHint := func() {
		o.currentPlayerKit.Say("§6选择目标基准点以决定复制的位置")
		o.selectInfo.nextSelect = BasePointTwo
		o.selectInfo.triggerFN = finalFunc
	}
	activateBasePointOneHint := func() {
		o.currentPlayerKit.Say("§6选择当前区域基准点(当然你也可以再点击一下起点)")
		o.selectInfo.nextSelect = BasePointOne
		o.selectInfo.triggerFN = func() {
			activateFinalHint()
		}
	}
	activateAreaTwoHint := func() {
		o.currentPlayerKit.Say("§6选择当前区域结束点")
		o.selectInfo.nextSelect = AreaPosTwo
		o.selectInfo.triggerFN = func() {
			activateBasePointOneHint()
		}
	}
	activateAreaOneHint := func() {
		o.currentPlayerKit.Say("§6选择当前区域起始点")
		o.selectInfo.nextSelect = AreaPosOne
		o.selectInfo.triggerFN = func() {
			activateAreaTwoHint()
		}
	}
	if o.selectInfo.currentSelectID == BasePointOne {
		return func(chat *defines.GameChat) {
			activateFinalHint()
		}, "复制", "输入 复制 以复制选中的区域，复制位置由两个基准点决定", true
	} else if o.selectInfo.currentSelectID == AreaPosTwo {
		return func(chat *defines.GameChat) {
			activateBasePointOneHint()
		}, "复制", "输入 复制 以复制选中的区域，复制位置由两个基准点决定", true
	} else if o.selectInfo.currentSelectID == AreaPosOne {
		return func(chat *defines.GameChat) {
			activateAreaTwoHint()
		}, "复制", "输入 复制 以复制选中的区域，区域由两个基准点决定", true
	} else if o.selectInfo.currentSelectID == NotSelect {
		return func(chat *defines.GameChat) {
			activateAreaOneHint()
		}, "复制", "输入 复制 以复制选中的区域，区域由两个基准点决定", true
	}
	return nil, "", "", false
}

func moveEntry(o *WoodAxe) (action func(chat *defines.GameChat), actionName string, hint string, available bool) {
	if o.actionsOccupied.occupied {
		return nil, "", "", false
	}
	finalFunc := func() {
		start, end := sortPos(o.selectInfo.pos[AreaPosOne], o.selectInfo.pos[AreaPosTwo])
		offset := o.selectInfo.pos[BasePointTwo].Sub(o.selectInfo.pos[BasePointOne])
		size := end.Sub(start)
		target := start.Add(offset)
		o.actionManager.Commit(&Action{
			Do: func() {
				o.Frame.GetGameControl().SendCmd(fmt.Sprintf("clone %v %v %v %v %v %v %v %v %v",
					start[0], start[1], start[2], end[0], end[1], end[2], target[0], target[1], target[2]))
				o.Frame.GetGameControl().SendCmd(fmt.Sprintf("fill %v %v %v %v %v %v air 0",
					start[0], start[1], start[2], end[0], end[1], end[2]))
			},
			AffectAreas: [][2]define.CubePos{{start, end}, {target, target.Add(size)}},
		})

		o.selectInfo.nextSelect = AreaPosOne
		o.selectInfo.currentSelectID = NotSelect
		o.selectInfo.triggerFN = nil
	}
	activateFinalHint := func() {
		o.currentPlayerKit.Say("§6选择目标基准点以决定移动的位置")
		o.selectInfo.nextSelect = BasePointTwo
		o.selectInfo.triggerFN = finalFunc
	}
	activateBasePointOneHint := func() {
		o.currentPlayerKit.Say("§6选择当前区域基准点(当然你也可以再点击一下起点)")
		o.selectInfo.nextSelect = BasePointOne
		o.selectInfo.triggerFN = func() {
			activateFinalHint()
		}
	}
	activateAreaTwoHint := func() {
		o.currentPlayerKit.Say("§6选择当前区域结束点")
		o.selectInfo.nextSelect = AreaPosTwo
		o.selectInfo.triggerFN = func() {
			activateBasePointOneHint()
		}
	}
	activateAreaOneHint := func() {
		o.currentPlayerKit.Say("§6选择当前区域起始点")
		o.selectInfo.nextSelect = AreaPosOne
		o.selectInfo.triggerFN = func() {
			activateAreaTwoHint()
		}
	}
	if o.selectInfo.currentSelectID == BasePointOne {
		return func(chat *defines.GameChat) {
			activateFinalHint()
		}, "移动", "输入 移动 以复制选中的区域，移动位置由两个基准点决定", true
	} else if o.selectInfo.currentSelectID == AreaPosTwo {
		return func(chat *defines.GameChat) {
			activateBasePointOneHint()
		}, "移动", "输入 移动 以复制选中的区域，移动位置由两个基准点决定", true
	} else if o.selectInfo.currentSelectID == AreaPosOne {
		return func(chat *defines.GameChat) {
			activateAreaTwoHint()
		}, "移动", "输入 移动 以移动选中的区域，区域由两个基准点决定", true
	} else if o.selectInfo.currentSelectID == NotSelect {
		return func(chat *defines.GameChat) {
			activateAreaOneHint()
		}, "移动", "输入 移动 以移动选中的区域，区域由两个基准点决定", true
	}
	return nil, "", "", false
}

func continuousCopyEntry(o *WoodAxe) (action func(chat *defines.GameChat), actionName string, hint string, available bool) {
	if o.actionsOccupied.occupied && !o.actionsOccupied.continuousCopy {
		return nil, "", "", false
	} else if o.actionsOccupied.occupied && o.actionsOccupied.continuousCopy {
		return func(chat *defines.GameChat) {
			o.selectInfo.nextSelect = AreaPosOne
			o.selectInfo.currentSelectID = NotSelect
			o.selectInfo.triggerFN = nil
			o.actionsOccupied.occupied = false
			o.actionsOccupied.continuousCopy = false
			o.currentPlayerKit.Say("已停止")
		}, "停止", "§6输入 停止 以停止连续复制", true
	}
	finalFunc := func() {
		start, end := sortPos(o.selectInfo.pos[AreaPosOne], o.selectInfo.pos[AreaPosTwo])
		offset := o.selectInfo.pos[BasePointTwo].Sub(o.selectInfo.pos[BasePointOne])
		size := end.Sub(start)
		target := start.Add(offset)
		o.actionManager.Commit(&Action{
			Do: func() {
				o.Frame.GetGameControl().SendCmd(fmt.Sprintf("clone %v %v %v %v %v %v %v %v %v",
					start[0], start[1], start[2], end[0], end[1], end[2], target[0], target[1], target[2]))
			},
			AffectAreas: [][2]define.CubePos{{target, target.Add(size)}},
		})
		o.selectInfo.nextSelect = BasePointTwo
		o.currentPlayerKit.Say("§6选择复制基准点")
		o.currentPlayerKit.Say("§6要停止复制请输入 停止")
	}
	activateFinalHint := func() {
		o.currentPlayerKit.Say("§6选择目标基准点以决定复制的位置")
		o.selectInfo.nextSelect = BasePointTwo
		o.selectInfo.triggerFN = finalFunc
		o.actionsOccupied.occupied = true
		o.actionsOccupied.continuousCopy = true
	}
	activateBasePointOneHint := func() {
		o.currentPlayerKit.Say("§6选择当前区域基准点(当然你也可以再点击一下起点)")
		o.selectInfo.nextSelect = BasePointOne
		o.selectInfo.triggerFN = func() {
			activateFinalHint()
		}
	}
	activateAreaTwoHint := func() {
		o.currentPlayerKit.Say("§6选择当前区域结束点")
		o.selectInfo.nextSelect = AreaPosTwo
		o.selectInfo.triggerFN = func() {
			activateBasePointOneHint()
		}
	}
	activateAreaOneHint := func() {
		o.currentPlayerKit.Say("§6选择当前区域起始点")
		o.selectInfo.nextSelect = AreaPosOne
		o.selectInfo.triggerFN = func() {
			activateAreaTwoHint()
		}
	}
	if o.selectInfo.currentSelectID == BasePointOne {
		return func(chat *defines.GameChat) {
			activateFinalHint()
		}, "连续复制", "输入 连续复制 以复制选中的区域，区域由第一个基准点和后续每个复制基准点决定", true
	} else if o.selectInfo.currentSelectID == AreaPosTwo {
		return func(chat *defines.GameChat) {
			activateBasePointOneHint()
		}, "连续复制", "输入 连续复制 以复制选中的区域，区域由第一个基准点和后续每个复制基准点决定", true
	} else if o.selectInfo.currentSelectID == AreaPosOne {
		return func(chat *defines.GameChat) {
			activateAreaTwoHint()
		}, "连续复制", "输入 连续复制 以复制选中的区域，区域由两个基准点决定", true
	} else if o.selectInfo.currentSelectID == NotSelect {
		return func(chat *defines.GameChat) {
			activateAreaOneHint()
		}, "连续复制", "输入 连续复制 以复制选中的区域，区域由两个基准点决定", true
	}
	return nil, "", "", false
}

func undoEntry(o *WoodAxe) (action func(chat *defines.GameChat), actionName string, hint string, available bool) {
	if o.actionsOccupied.occupied && !o.actionsOccupied.undo {
		return nil, "", "", false
	}
	return func(chat *defines.GameChat) {
		o.selectInfo.triggerFN = func() {
			o.currentPlayerKit.Say("§6输入 完成 以完成接受当前状态")
		}
		step := 1
		if len(chat.Msg) > 0 {
			if _steps, err := strconv.Atoi(chat.Msg[0]); err == nil {
				if _steps < 1 {
					o.currentPlayerKit.Say("§6撤销的步数无效")
					return
				} else {
					step = _steps
				}
			} else {
				o.currentPlayerKit.Say("§6撤销的步数无效")
				return
			}
		}
		o.actionManager.Freeze()
		o.actionsOccupied.occupied = true
		o.actionsOccupied.undo = true
		for i := 0; i < step; i++ {
			if err := o.actionManager.Undo(); err != nil {
				o.currentPlayerKit.Say("§6无法继续撤销了")
				break
			}
		}
		o.currentPlayerKit.Say("你可以继续输入 撤销/重做/完成")
	}, "撤销", "输入 撤销 [数量] 以撤销指定数量的操作，不指定数量时默认撤销一步", true
}

func redoEntry(o *WoodAxe) (action func(chat *defines.GameChat), actionName string, hint string, available bool) {
	if !(o.actionsOccupied.occupied && o.actionsOccupied.undo) {
		return nil, "", "", false
	}
	return func(chat *defines.GameChat) {
		o.selectInfo.triggerFN = func() {
			o.currentPlayerKit.Say("输入 完成 以完成接受当前状态")
		}
		step := 1
		if len(chat.Msg) > 0 {
			if _steps, err := strconv.Atoi(chat.Msg[0]); err == nil {
				if _steps < 1 {
					o.currentPlayerKit.Say("§6重做的步数无效")
					return
				} else {
					step = _steps
				}
			} else {
				o.currentPlayerKit.Say("§6重做的步数无效")
				return
			}
		}
		o.actionsOccupied.occupied = true
		o.actionsOccupied.undo = true
		for i := 0; i < step; i++ {
			if err := o.actionManager.Redo(); err != nil {
				o.currentPlayerKit.Say("§6无法继续重做了")
				break
			}
		}
		o.currentPlayerKit.Say("§6你可以继续输入 撤销/重做/完成")
	}, "重做", "输入 重做 [数量] 以重做指定数量的操作，不指定数量时默认重做一步", true
}

func doneUndoEntry(o *WoodAxe) (action func(chat *defines.GameChat), actionName string, hint string, available bool) {
	if !(o.actionsOccupied.occupied && o.actionsOccupied.undo) {
		return nil, "", "", false
	}
	return func(chat *defines.GameChat) {
		o.actionsOccupied.occupied = false
		o.actionsOccupied.undo = false
		o.selectInfo.triggerFN = nil
		o.actionManager.Trim()
		o.actionManager.DeFreeze()
	}, "完成", "§6输入 完成 以接受当前的撤销动作", true
}

func (o *WoodAxe) splitLargeArea(start, end define.CubePos) [][2]define.CubePos {
	splitPosPair := func(start int, end int) [][2]int {
		if start > end {
			t := start
			start = end
			end = t
		}
		chunkBegin := int(math.Floor(float64(start)/16) * 16)
		pairs := [][2]int{}
		for {
			p := [2]int{chunkBegin, chunkBegin + 16}
			if chunkBegin < start {
				p[0] = start
			}
			if p[1] >= end {
				p[1] = end
				pairs = append(pairs, p)
				break
			} else {
				chunkBegin += 16
				pairs = append(pairs, p)
			}
		}
		return pairs
	}
	xg := splitPosPair(start[0], end[0])
	yg := splitPosPair(start[1], end[1])
	zg := splitPosPair(start[2], end[2])
	cubes := make([][2]define.CubePos, 0, len(xg)*len(yg)*len(zg))
	for xi := 0; xi < len(xg); xi++ {
		for zi := 0; zi < len(zg); zi++ {
			for yi := 0; yi < len(yg); yi++ {
				xr, yr, zr := xg[xi], yg[yi], zg[zi]
				cubes = append(cubes, [2]define.CubePos{{xr[0], yr[0], zr[0]}, {xr[1], yr[1], zr[1]}})
			}
		}
	}
	return cubes
}

func fillEntry(o *WoodAxe) (action func(chat *defines.GameChat), actionName string, hint string, available bool) {
	if o.actionsOccupied.occupied {
		return nil, "", "", false
	}

	block := "air"
	data := "0"

	doFill := func() {
		o.actionsOccupied.occupied = true
		o.actionsOccupied.largeFill = true
		cubes := o.splitLargeArea(o.selectInfo.pos[AreaPosOne], o.selectInfo.pos[AreaPosTwo])
		o.actionManager.Commit(&Action{
			Do: func() {
				for i, cube := range cubes {
					o.currentPlayerKit.ActionBar(fmt.Sprintf("fill sub chunk %v/%v", i, len(cubes)))
					o.Frame.GetGameControl().SendCmd(fmt.Sprintf("fill %v %v %v %v %v %v %v %v",
						cube[0][0], cube[0][1], cube[0][2], cube[1][0], cube[1][1], cube[1][2], block, data),
					)
					time.Sleep(time.Millisecond * 100)
				}
			},
			AffectAreas: cubes,
		})
		o.currentPlayerKit.Say("§6填充完成")
		o.actionsOccupied.occupied = false
		o.actionsOccupied.largeFill = false
		o.nextSelect = AreaPosOne
		o.selectInfo.triggerFN = nil
	}

	onAreaPosOneSelected := func() {
		o.currentPlayerKit.Say("§6请选择区域终点")
		o.selectInfo.triggerFN = doFill
	}

	onNotSelected := func() {
		o.currentPlayerKit.Say("§6请选择区域起点")
		o.selectInfo.triggerFN = onAreaPosOneSelected
	}

	return func(chat *defines.GameChat) {
		if len(chat.Msg) > 0 {
			block = chat.Msg[0]
		}
		if len(chat.Msg) > 1 {
			data = chat.Msg[1]
		}
		if o.selectInfo.currentSelectID > AreaPosOne {
			doFill()
		} else if o.selectInfo.currentSelectID == AreaPosOne {
			onAreaPosOneSelected()
		} else if o.selectInfo.currentSelectID == NotSelect {
			onNotSelected()
		}
	}, "填充", "输入 填充 [方块名] 以进行填充", true
}

func flipEntry(o *WoodAxe) (action func(chat *defines.GameChat), actionName string, hint string, available bool) {
	if o.actionsOccupied.occupied {
		return nil, "", "", false
	}

	dir := "x"

	doFill := func() {
		o.actionsOccupied.occupied = true
		o.actionsOccupied.largeFill = true
		area := [2]define.CubePos{o.selectInfo.pos[AreaPosOne], o.selectInfo.pos[AreaPosTwo]}
		startPos, endPos := sortPos(area[0], area[1])
		o.actionManager.Commit(&Action{
			Do: func() {
				cmd := fmt.Sprintf("structure save %v %v %v %v %v %v %v true memory ", "omwat", startPos[0], startPos[1], startPos[2], endPos[0], endPos[1], endPos[2])
				o.Frame.GetGameControl().SendCmd(cmd)
				time.Sleep(time.Millisecond * 50)
				cmd = fmt.Sprintf("structure load %v %v %v %v 0_degrees %v", "omwat", startPos[0], startPos[1], startPos[2], dir)
				o.Frame.GetGameControl().SendCmd(cmd)
			},
			AffectAreas: [][2]define.CubePos{area},
		})
		o.currentPlayerKit.Say("§6翻转完成")
		o.actionsOccupied.occupied = false
		o.actionsOccupied.largeFill = false
		o.nextSelect = AreaPosOne
		o.selectInfo.triggerFN = nil
	}

	onAreaPosOneSelected := func() {
		o.currentPlayerKit.Say("§6请选择区域终点")
		o.selectInfo.triggerFN = doFill
	}

	onNotSelected := func() {
		o.currentPlayerKit.Say("§6请选择区域起点")
		o.selectInfo.triggerFN = onAreaPosOneSelected
	}

	return func(chat *defines.GameChat) {
		if len(chat.Msg) > 0 {
			dir = chat.Msg[0]
		}
		if o.selectInfo.currentSelectID > AreaPosOne {
			doFill()
		} else if o.selectInfo.currentSelectID == AreaPosOne {
			onAreaPosOneSelected()
		} else if o.selectInfo.currentSelectID == NotSelect {
			onNotSelected()
		}
	}, "翻转", "输入 翻转 [x/z/xz] 以进行翻转", true
}

func replaceEntry(o *WoodAxe) (action func(chat *defines.GameChat), actionName string, hint string, available bool) {
	if o.actionsOccupied.occupied {
		return nil, "", "", false
	}

	block1 := ""
	data1 := ""
	block2 := ""
	data2 := ""

	doFill := func() {
		o.actionsOccupied.occupied = true
		o.actionsOccupied.largeFill = true
		cubes := o.splitLargeArea(o.selectInfo.pos[AreaPosOne], o.selectInfo.pos[AreaPosTwo])
		o.actionManager.Commit(&Action{
			Do: func() {
				for i, cube := range cubes {
					o.currentPlayerKit.ActionBar(fmt.Sprintf("replace sub chunk %v/%v", i, len(cubes)))
					o.Frame.GetGameControl().SendCmd(fmt.Sprintf("fill %v %v %v %v %v %v %v %v replace %v %v",
						cube[0][0], cube[0][1], cube[0][2], cube[1][0], cube[1][1], cube[1][2], block2, data2, block1, data1),
					)
					time.Sleep(time.Millisecond * 100)
				}
			},
			AffectAreas: cubes,
		})
		o.currentPlayerKit.Say("§6替换完成")
		o.actionsOccupied.occupied = false
		o.actionsOccupied.largeFill = false
		o.nextSelect = AreaPosOne
		o.selectInfo.triggerFN = nil
	}

	onAreaPosOneSelected := func() {
		o.currentPlayerKit.Say("§6请选择区域终点")
		o.selectInfo.triggerFN = doFill
	}

	onNotSelected := func() {
		o.currentPlayerKit.Say("§6请选择区域起点")
		o.selectInfo.triggerFN = onAreaPosOneSelected
	}

	return func(chat *defines.GameChat) {
		if len(chat.Msg) < 4 {
			o.currentPlayerKit.Say("§6参数不正确")
		}
		block1 = chat.Msg[0]
		data1 = chat.Msg[1]
		block2 = chat.Msg[2]
		data2 = chat.Msg[3]
		if o.selectInfo.currentSelectID > AreaPosOne {
			doFill()
		} else if o.selectInfo.currentSelectID == AreaPosOne {
			onAreaPosOneSelected()
		} else if o.selectInfo.currentSelectID == NotSelect {
			onNotSelected()
		}
	}, "替换", "输入 替换 [原方块名] [原方块值] [新方块名] [新方块值] 以进行替换", true
}

func largeFillEntry(o *WoodAxe) (action func(chat *defines.GameChat), actionName string, hint string, available bool) {
	if o.actionsOccupied.occupied {
		return nil, "", "", false
	}

	return func(chat *defines.GameChat) {
		block := "air"
		data := "0"
		start := define.CubePos{0, 0, 0}
		end := define.CubePos{0, 0, 0}
		if len(chat.Msg) < 8 {
			o.currentPlayerKit.Say("参数不正确")
		}
		if i, err := strconv.Atoi(chat.Msg[0]); err != nil {
			o.currentPlayerKit.Say("参数不正确")
		} else {
			start[0] = i
		}
		if i, err := strconv.Atoi(chat.Msg[1]); err != nil {
			o.currentPlayerKit.Say("参数不正确")
		} else {
			start[1] = i
		}
		if i, err := strconv.Atoi(chat.Msg[2]); err != nil {
			o.currentPlayerKit.Say("参数不正确")
		} else {
			start[2] = i
		}
		if i, err := strconv.Atoi(chat.Msg[3]); err != nil {
			o.currentPlayerKit.Say("参数不正确")
		} else {
			end[0] = i
		}
		if i, err := strconv.Atoi(chat.Msg[4]); err != nil {
			o.currentPlayerKit.Say("参数不正确")
		} else {
			end[1] = i
		}
		if i, err := strconv.Atoi(chat.Msg[5]); err != nil {
			o.currentPlayerKit.Say("参数不正确")
		} else {
			end[2] = i
		}
		block = chat.Msg[6]
		data = chat.Msg[7]
		go func() {
			o.actionsOccupied.occupied = true
			o.actionsOccupied.largeFill = true
			cubes := o.splitLargeArea(start, end)
			for i, cube := range cubes {
				o.currentPlayerKit.ActionBar(fmt.Sprintf("fill sub chunk %v/%v", i, len(cubes)))
				o.Frame.GetGameControl().SendCmd(fmt.Sprintf("tp @s %v %v %v ",
					cube[0][0], cube[0][1], cube[0][2]),
				)
				time.Sleep(time.Millisecond * 100)
				o.Frame.GetGameControl().SendCmd(fmt.Sprintf("fill %v %v %v %v %v %v %v %v",
					cube[0][0], cube[0][1], cube[0][2], cube[1][0], cube[1][1], cube[1][2], block, data),
				)
				time.Sleep(time.Millisecond * 100)
			}
			fmt.Println("填充完成")
			o.currentPlayerKit.Say("填充完成")
			o.actionsOccupied.occupied = false
			o.actionsOccupied.largeFill = false
		}()
	}, "largefill", "输入 largefill [x0] [y0] [z0] [x1] [y1] [z1] [方块名] [方块参数] 以进行大范围 fill (无法撤销)", true
}

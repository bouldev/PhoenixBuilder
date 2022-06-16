package woodaxe

import (
	"fmt"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/omega/defines"
	"strconv"
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
		o.currentPlayerKit.Say("选择目标基准点以决定复制的位置")
		o.selectInfo.nextSelect = BasePointTwo
		o.selectInfo.triggerFN = finalFunc
	}
	activateBasePointOneHint := func() {
		o.currentPlayerKit.Say("选择当前区域基准点(当然你也可以再点击一下起点)")
		o.selectInfo.nextSelect = BasePointOne
		o.selectInfo.triggerFN = func() {
			activateFinalHint()
		}
	}
	activateAreaTwoHint := func() {
		o.currentPlayerKit.Say("选择当前区域结束点")
		o.selectInfo.nextSelect = AreaPosTwo
		o.selectInfo.triggerFN = func() {
			activateBasePointOneHint()
		}
	}
	activateAreaOneHint := func() {
		o.currentPlayerKit.Say("选择当前区域起始点")
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
		}, "停止", "输入 停止 以停止连续复制", true
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
		o.currentPlayerKit.Say("选择复制基准点")
	}
	activateFinalHint := func() {
		o.currentPlayerKit.Say("选择目标基准点以决定复制的位置")
		o.selectInfo.nextSelect = BasePointTwo
		o.selectInfo.triggerFN = finalFunc
		o.actionsOccupied.occupied = true
		o.actionsOccupied.continuousCopy = true
	}
	activateBasePointOneHint := func() {
		o.currentPlayerKit.Say("选择当前区域基准点(当然你也可以再点击一下起点)")
		o.selectInfo.nextSelect = BasePointOne
		o.selectInfo.triggerFN = func() {
			activateFinalHint()
		}
	}
	activateAreaTwoHint := func() {
		o.currentPlayerKit.Say("选择当前区域结束点")
		o.selectInfo.nextSelect = AreaPosTwo
		o.selectInfo.triggerFN = func() {
			activateBasePointOneHint()
		}
	}
	activateAreaOneHint := func() {
		o.currentPlayerKit.Say("选择当前区域起始点")
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
		step := 1
		if len(chat.Msg) > 0 {
			if _steps, err := strconv.Atoi(chat.Msg[0]); err == nil {
				if _steps < 1 {
					o.currentPlayerKit.Say("撤销的步数无效")
					return
				} else {
					step = _steps
				}
			} else {
				o.currentPlayerKit.Say("撤销的步数无效")
				return
			}
		}
		o.actionManager.Freeze()
		o.actionsOccupied.occupied = true
		o.actionsOccupied.undo = true
		for i := 0; i < step; i++ {
			if err := o.actionManager.Undo(); err != nil {
				o.currentPlayerKit.Say("无法继续撤销了")
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
		step := 1
		if len(chat.Msg) > 0 {
			if _steps, err := strconv.Atoi(chat.Msg[0]); err == nil {
				if _steps < 1 {
					o.currentPlayerKit.Say("重做的步数无效")
					return
				} else {
					step = _steps
				}
			} else {
				o.currentPlayerKit.Say("重做的步数无效")
				return
			}
		}
		o.actionsOccupied.occupied = true
		o.actionsOccupied.undo = true
		for i := 0; i < step; i++ {
			if err := o.actionManager.Redo(); err != nil {
				o.currentPlayerKit.Say("无法继续重做了")
				break
			}
		}
		o.currentPlayerKit.Say("你可以继续输入 撤销/重做/完成")
	}, "重做", "输入 重做 [数量] 以重做指定数量的操作，不指定数量时默认重做一步", true
}

func doneUndoEntry(o *WoodAxe) (action func(chat *defines.GameChat), actionName string, hint string, available bool) {
	if !(o.actionsOccupied.occupied && o.actionsOccupied.undo) {
		return nil, "", "", false
	}
	return func(chat *defines.GameChat) {
		o.actionsOccupied.occupied = false
		o.actionsOccupied.undo = false
		o.actionManager.Trim()
		o.actionManager.DeFreeze()
	}, "完成", "输入 完成 以接受当前的撤销动作", true
}

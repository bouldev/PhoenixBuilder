package woodaxe

import (
	"fmt"
	"phoenixbuilder/omega/defines"
)

func copyEntry(o *WoodAxe) (action func(chat *defines.GameChat), actionName string, hint string, available bool) {
	if o.actionsOccupied.occupied {
		return nil, "", "", false
	}
	finalFunc := func() {
		start, end := sortPos(o.selectInfo.pos[AreaPosOne], o.selectInfo.pos[AreaPosTwo])
		offset := o.selectInfo.pos[BasePointTwo].Sub(o.selectInfo.pos[BasePointOne])
		target := start.Add(offset)
		o.Frame.GetGameControl().SendCmd(fmt.Sprintf("clone %v %v %v %v %v %v %v %v %v",
			start[0], start[1], start[2], end[0], end[1], end[2], target[0], target[1], target[2]))
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
		target := start.Add(offset)
		o.Frame.GetGameControl().SendCmd(fmt.Sprintf("clone %v %v %v %v %v %v %v %v %v",
			start[0], start[1], start[2], end[0], end[1], end[2], target[0], target[1], target[2]))
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

package task_config

import (
	"fmt"
	"phoenixbuilder/fastbuilder/configuration"
	"phoenixbuilder/fastbuilder/task"
	"phoenixbuilder/fastbuilder/types"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type GUI struct {
	setContent   func(v fyne.CanvasObject)
	getContent   func() fyne.CanvasObject
	origContent  fyne.CanvasObject
	masterWindow fyne.Window

	content      fyne.CanvasObject
	majorContent fyne.CanvasObject
}

func New() *GUI {
	gui := &GUI{}
	return gui
}

type DelaySetter interface {
	DelayConfigGetter() *types.DelayConfig
	DelayConfigSetter(*types.DelayConfig)
	CreationTypeGetter() byte
	CreationTypeSetter(byte)
	Submit() bool
}
type GlobalDelaySetter struct {
	mirrorDelayConfig  *types.DelayConfig
	mirrorCreationType byte
}

func (gds *GlobalDelaySetter) DelayConfigGetter() *types.DelayConfig {
	return gds.mirrorDelayConfig
}
func (gds *GlobalDelaySetter) DelayConfigSetter(dc *types.DelayConfig) {
	gds.mirrorDelayConfig = dc
}
func (gds *GlobalDelaySetter) CreationTypeGetter() byte {
	return gds.mirrorCreationType
}
func (gds *GlobalDelaySetter) CreationTypeSetter(b byte) {
	gds.mirrorCreationType = b
}

func (gds *GlobalDelaySetter) Submit() bool {
	configuration.GlobalFullConfig().Delay().Delay = gds.mirrorDelayConfig.Delay
	configuration.GlobalFullConfig().Delay().DelayMode = gds.mirrorDelayConfig.DelayMode
	configuration.GlobalFullConfig().Delay().DelayThreshold = gds.mirrorDelayConfig.DelayThreshold
	configuration.GlobalFullConfig().Global().TaskCreationType = gds.mirrorCreationType
	return true
}

func makeGlobalDelaySetter() *GlobalDelaySetter {
	_mirrorDelayConfig := *configuration.GlobalFullConfig().Delay()
	return &GlobalDelaySetter{
		mirrorDelayConfig:  &_mirrorDelayConfig,
		mirrorCreationType: configuration.GlobalFullConfig().Global().TaskCreationType,
	}
}

type TaskDelaySetter struct {
	id                 interface{}
	mirrorDelayConfig  *types.DelayConfig
	mirrorCreationType byte
	task               *task.Task
}

func (tds *TaskDelaySetter) DelayConfigGetter() *types.DelayConfig {
	return tds.mirrorDelayConfig
}
func (tds *TaskDelaySetter) DelayConfigSetter(dc *types.DelayConfig) {
	tds.mirrorDelayConfig = dc
}
func (tds *TaskDelaySetter) CreationTypeGetter() byte {
	return tds.mirrorCreationType
}
func (tds *TaskDelaySetter) CreationTypeSetter(b byte) {
	tds.mirrorCreationType = b
}

func (tds *TaskDelaySetter) Submit() bool {
	v, ok := task.TaskMap.Load(tds.id)
	if !ok {
		return false
	}
	av := v.(*task.Task)
	av.Config.Delay().Delay = tds.mirrorDelayConfig.Delay
	av.Config.Delay().DelayMode = tds.mirrorDelayConfig.DelayMode
	av.Config.Delay().DelayThreshold = tds.mirrorDelayConfig.DelayThreshold
	return true
}

func (tds *TaskDelaySetter) Pause() {
	tds.task.Pause()
}

func (tds *TaskDelaySetter) Resume() {
	tds.task.Resume()
}

func (tds *TaskDelaySetter) Break() {
	tds.task.Break()
}

func makeAllTasksSetter() []fyne.CanvasObject {
	taskDelaySetters := make([]fyne.CanvasObject, 0)
	task.TaskMap.Range(func(k, v interface{}) bool {
		t := v.(*task.Task)
		_mirrorDelayConfig := *(t.Config.Delay())
		var content fyne.CanvasObject

		taskDelaySetter := &TaskDelaySetter{
			id:                 k,
			mirrorDelayConfig:  &_mirrorDelayConfig,
			mirrorCreationType: t.Type,
			task:               t,
		}
		delaySettingerGUI := MakeDelaySetterGUI(taskDelaySetter, false)
		taskHandler := container.NewHBox(
			widget.NewLabel(fmt.Sprintf("Task-%v", k)),
			container.NewGridWithColumns(3,
				widget.NewButtonWithIcon("暂停", theme.MediaPauseIcon(), func() {
					t.Pause()
				}),
				widget.NewButtonWithIcon("恢复", theme.MediaPlayIcon(), func() {
					t.Resume()
				}),
				widget.NewButtonWithIcon("取消", theme.MediaStopIcon(), func() {
					t.Break()
					content.Hide()
				}),
			),
		)
		content = container.NewVBox(
			taskHandler,
			delaySettingerGUI.content,
		)
		taskDelaySetters = append(taskDelaySetters, content)
		return true
	})
	return taskDelaySetters
}

type DelaySetterGUI struct {
	ds                    DelaySetter
	taskTypeName          *widget.Label
	taskTypeRG            *widget.RadioGroup
	delayTypeRG           *widget.RadioGroup
	typeContinuousContent fyne.CanvasObject
	typeDiscreteContent   fyne.CanvasObject
	typeNoneContent       fyne.CanvasObject
	bindDelay             binding.ExternalInt
	bindDelayThres        binding.ExternalInt
	submit                *widget.Button
	content               fyne.CanvasObject
	isGlobal              bool
}

const DescriptionSync = "边算边建(无进度显示)"
const DescriptionAsync = "先算后建(推荐)"
const DescriptionContinuous = "连续(每放置一个方块等待一会儿/推荐)"
const DescriptionDiscrete = "离散(每放置几个方块等待一会儿)"
const DescriptionNone = "极限速度"

func (dsg *DelaySetterGUI) UpdateUI(firstOpen bool) {
	taskType := dsg.ds.CreationTypeGetter()
	if taskType == types.TaskTypeSync {
		if !dsg.isGlobal {
			dsg.taskTypeName.Text += ": " + DescriptionSync
		}
		dsg.taskTypeRG.SetSelected(DescriptionSync)
	} else if taskType == types.TaskTypeAsync {
		if !dsg.isGlobal {
			dsg.taskTypeName.Text += ": " + DescriptionAsync
		}
		dsg.taskTypeRG.SetSelected(DescriptionAsync)
	}
	DelayMode := dsg.ds.DelayConfigGetter().DelayMode
	if DelayMode == types.DelayModeContinuous {
		dsg.delayTypeRG.SetSelected(DescriptionContinuous)
	} else if DelayMode == types.DelayModeDiscrete {
		dsg.delayTypeRG.SetSelected(DescriptionDiscrete)
	} else if DelayMode == types.DelayModeNone {
		dsg.delayTypeRG.SetSelected(DescriptionNone)
	}
	dsg.updateDelayContent(DelayMode, firstOpen)
}

func (dsg *DelaySetterGUI) updateDelayContent(delayMode byte, firstOpen bool) {
	switch delayMode {
	case types.DelayModeContinuous:
		if !firstOpen {
			dsg.bindDelay.Set(1000)
		}
		dsg.typeContinuousContent.Show()
		dsg.typeDiscreteContent.Hide()
		dsg.typeNoneContent.Hide()
	case types.DelayModeDiscrete:
		if !firstOpen {
			dsg.bindDelay.Set(15)
			dsg.bindDelayThres.Set(20000)
		}
		dsg.typeContinuousContent.Hide()
		dsg.typeDiscreteContent.Show()
		dsg.typeNoneContent.Hide()
	case types.DelayModeNone:
		if firstOpen {
			dsg.bindDelay.Set(0)
		}
		dsg.typeContinuousContent.Hide()
		dsg.typeDiscreteContent.Hide()
		dsg.typeNoneContent.Show()
	}
}

func MakeDelaySetterGUI(ds DelaySetter, isGlobal bool) *DelaySetterGUI {
	dsg := &DelaySetterGUI{
		ds:       ds,
		isGlobal: isGlobal,
	}
	dsg.taskTypeRG = &widget.RadioGroup{
		Horizontal: false,
		Options:    []string{DescriptionSync, DescriptionAsync},
		Required:   true,
	}
	dsg.delayTypeRG = &widget.RadioGroup{
		Horizontal: false,
		Options:    []string{DescriptionContinuous, DescriptionDiscrete, DescriptionNone},
		Required:   true,
	}
	dsg.typeNoneContent = container.NewHBox(
		widget.NewLabelWithStyle("极限速度/不稳定/极其快", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	)
	Delay := int(dsg.ds.DelayConfigGetter().Delay)
	bDelay := binding.BindInt(&Delay)
	dsg.bindDelay = bDelay
	// delayEntry := widget.NewEntryWithData(binding.IntToString(bDelay))
	DelayThres := int(dsg.ds.DelayConfigGetter().DelayThreshold)
	bDelayThres := binding.BindInt(&DelayThres)
	dsg.bindDelayThres = bDelayThres
	delayDelayThresEntry := widget.NewEntryWithData(binding.IntToString(bDelayThres))
	dsg.typeContinuousContent = container.NewVBox(
		widget.NewLabelWithStyle("连续发包/稳定", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewBorder(nil, nil, widget.NewLabel("放置每个方块的延迟(毫秒):"), nil, widget.NewEntryWithData(binding.IntToString(bDelay))),
	)
	dsg.typeDiscreteContent = container.NewVBox(
		widget.NewLabelWithStyle("离散发包/较慢", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewBorder(nil, nil, widget.NewLabel("每放置"), widget.NewLabel("个方块"), delayDelayThresEntry),
		container.NewBorder(nil, nil, widget.NewLabel("就等待"), widget.NewLabel("秒"), widget.NewEntryWithData(binding.IntToString(bDelay))),
	)
	dsg.submit = &widget.Button{
		Text: "设置",
		OnTapped: func() {
			origConfig := dsg.ds.DelayConfigGetter()
			nv, err := bDelayThres.Get()
			if err == nil {
				origConfig.DelayThreshold = int(nv)
			}
			nv, err = bDelay.Get()
			if err == nil {
				origConfig.Delay = int64(nv)
			}
			dsg.ds.DelayConfigSetter(origConfig)

			success := dsg.ds.Submit()
			if !success {
				dsg.content.Hide()
			}
		},
		Icon:          theme.ConfirmIcon(),
		IconPlacement: widget.ButtonIconTrailingText,
		Importance:    widget.HighImportance,
	}
	dsg.taskTypeName = widget.NewLabel("建造模式")
	dsg.UpdateUI(true)
	if isGlobal {
		dsg.taskTypeRG.OnChanged = func(s string) {
			if s == DescriptionAsync {
				dsg.ds.CreationTypeSetter(types.TaskTypeAsync)
			} else if s == DescriptionSync {
				dsg.ds.CreationTypeSetter(types.TaskTypeSync)
			}
		}
		dsg.taskTypeName.Hide()

	} else {
		dsg.taskTypeRG.Hide()
		dsg.taskTypeRG.Disable()
	}
	dsg.delayTypeRG.OnChanged = func(s string) {
		delayMode := byte(types.DelayModeInvalid)
		if s == DescriptionContinuous {
			delayMode = types.DelayModeContinuous
		} else if s == DescriptionDiscrete {
			delayMode = types.DelayModeDiscrete
		} else if s == DescriptionNone {
			delayMode = types.DelayModeNone
		}
		dsg.updateDelayContent(byte(delayMode), false)
		currentDelay := dsg.ds.DelayConfigGetter()
		currentDelay.DelayMode = delayMode
		dsg.ds.DelayConfigSetter(currentDelay)
	}
	dsg.content = container.NewVBox(
		dsg.taskTypeName,
		dsg.taskTypeRG,
		widget.NewSeparator(),
		widget.NewLabel("延迟模式"),
		dsg.delayTypeRG,
		dsg.typeNoneContent,
		dsg.typeContinuousContent,
		dsg.typeDiscreteContent,
		widget.NewSeparator(),
		dsg.submit,
	)
	return dsg
}

func (g *GUI) makeMajorContent() fyne.CanvasObject {
	globalSetter := makeGlobalDelaySetter()
	globalSetterWidget := MakeDelaySetterGUI(globalSetter, true)
	taskSetters := makeAllTasksSetter()
	var taskContent fyne.CanvasObject
	if len(taskSetters) == 0 {
		taskContent = widget.NewLabel("还没有运行中的任务")
	} else {
		taskContent = container.NewVBox(taskSetters...)
	}

	return container.NewVBox(
		widget.NewCard("全局配置", "对新任务生效", globalSetterWidget.content),
		// widget.NewLabelWithStyle("全局配置(对新任务生效)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		// globalSetterWidget.content,
		// widget.NewSeparator(),
		widget.NewLabelWithStyle("现有任务", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewCard("现有任务", "调整正在运行的任务", taskContent),
	)
}

func (g *GUI) GetContent(setContent func(v fyne.CanvasObject), getContent func() fyne.CanvasObject, masterWindow fyne.Window) fyne.CanvasObject {
	g.origContent = getContent()
	g.setContent = setContent
	g.getContent = getContent
	g.masterWindow = masterWindow
	g.majorContent = g.makeMajorContent()
	g.content = container.NewBorder(nil, &widget.Button{
		Text: "关闭",
		OnTapped: func() {
			g.setContent(g.origContent)
		},
		Icon:          theme.CancelIcon(),
		IconPlacement: widget.ButtonIconLeadingText,
	}, nil, nil, container.NewVScroll(g.majorContent))

	return g.content
}

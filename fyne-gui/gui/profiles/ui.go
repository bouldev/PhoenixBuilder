package profiles

import (
	"fmt"
	"io/ioutil"
	"sort"

	"gopkg.in/yaml.v3"

	"phoenixbuilder/dedicate/fyne/session"
	"phoenixbuilder_fyne_gui/gui/profiles/config"
	ui_session "phoenixbuilder_fyne_gui/gui/profiles/session"

	"fyne.io/fyne/v2/dialog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type GUI struct {
	setContent   func(v fyne.CanvasObject)
	getContent   func() fyne.CanvasObject
	masterWindow fyne.Window
	app          fyne.App
	onPanic      func(error)
	content      fyne.CanvasObject
	entryIndex   []int
	configs      map[int]*config.SessionConfigWithName
	counter      int
	storage      fyne.Storage
}

func New(storage fyne.Storage) *GUI {
	//configPath := path.Join(dataFolder, "config.yaml")
	gui := &GUI{
		entryIndex: make([]int, 0),
		configs:    make(map[int]*config.SessionConfigWithName, 0),
		counter:    0,
		storage:    storage,
	}
	return gui
}

func makeProfileEntry() fyne.CanvasObject {
	canvasObject := container.NewBorder(
		nil, nil,
		widget.NewLabel("err,name not set"),
		container.NewHBox(
			&widget.Button{
				Text:          "",
				Icon:          theme.DeleteIcon(),
				IconPlacement: widget.ButtonIconLeadingText,
				Importance:    widget.LowImportance,
			},
			&widget.Button{
				Text:          "",
				Icon:          theme.DocumentCreateIcon(),
				IconPlacement: widget.ButtonIconLeadingText,
				Importance:    widget.LowImportance,
			},
			&widget.Button{
				Text:          "登录",
				Icon:          theme.MailSendIcon(),
				IconPlacement: widget.ButtonIconLeadingText,
				Importance:    widget.HighImportance,
			},
		),
	)
	return canvasObject
}

func updateProfileEntry(entry fyne.CanvasObject, name string, deleteFn func(), editFn func(), loginFn func()) {
	c := entry.(*fyne.Container)
	c.Objects[0].(*widget.Label).SetText(name)
	bs := c.Objects[1].(*fyne.Container).Objects
	bs[0].(*widget.Button).OnTapped = deleteFn
	bs[1].(*widget.Button).OnTapped = editFn
	bs[2].(*widget.Button).OnTapped = loginFn
}

func (g *GUI) updateEntryIndex() {
	g.entryIndex = make([]int, 0)
	for k, _ := range g.configs {
		g.entryIndex = append(g.entryIndex, k)
	}
	sort.Ints(g.entryIndex)
}

func (g *GUI) newConfig(config *config.SessionConfigWithName) int {
	g.counter++
	id := g.counter
	g.configs[id] = config
	return id
}

func (g *GUI) makeProfilesList() fyne.CanvasObject {
	profilesList := widget.NewList(
		func() int {
			return len(g.entryIndex)
		},
		func() fyne.CanvasObject {
			return makeProfileEntry()
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			updateProfileEntry(
				o,
				g.configs[g.entryIndex[i]].Name,
				func() { g.onDelete(g.entryIndex[i]) },
				func() { g.onEdit(g.entryIndex[i]) },
				func() { g.onLogin(g.entryIndex[i]) },
			)
		},
	)
	return profilesList
}

func (g *GUI) onEdit(i int) {
	configForm := config.New(g.configs[i], func(filledConfig *config.SessionConfigWithName) {
		g.configs[i] = filledConfig
		g.updateEntryIndex()
		g.content.Refresh()
		g.WriteBackConfigFile()
	})
	g.setContent(configForm.GetContent(g.masterWindow, g.setContent, g.getContent))
}

func (g *GUI) onDelete(i int) {
	delete(g.configs, i)
	g.updateEntryIndex()
	g.content.Refresh()
	g.WriteBackConfigFile()
}

func (g *GUI) onLogin(i int) {
	s := ui_session.New(g.configs[i], g.WriteBackConfigFile)
	g.setContent(s.GetContent(g.setContent, g.getContent, g.masterWindow, g.app))
	s.AfterMount()
	fmt.Println("login", i)
}

func (g *GUI) onNewProfile() {
	blankConfig := &config.SessionConfigWithName{Name: "未命名配置" + fmt.Sprintf("%d", g.counter), Config: session.NewConfig()}

	configForm := config.New(blankConfig, func(filledConfig *config.SessionConfigWithName) {
		g.newConfig(filledConfig)
		g.updateEntryIndex()
		g.content.Refresh()
		g.WriteBackConfigFile()
	})
	g.setContent(configForm.GetContent(g.masterWindow, g.setContent, g.getContent))
}

func (g *GUI) ReadConfigFile() []*config.SessionConfigWithName {
	plainConfigs := make([]*config.SessionConfigWithName, 0)
	open, err := g.storage.Open("config.yaml")
	if err == nil {
		defer open.Close()
		byteConfig, err := ioutil.ReadAll(open)
		if err != nil {
			g.onPanic(fmt.Errorf("无法读取配置文件:\n %v\n请检查权限或尝试手动删除该文件", open.URI().String()))
			return plainConfigs
		}
		err = yaml.Unmarshal(byteConfig, &plainConfigs)
		if err != nil {
			g.onPanic(fmt.Errorf("无法解析配置文件\n %v\n文件已损坏,请尝试手动删除该文件", open.URI().String()))
			return plainConfigs
		}
		return plainConfigs
	}
	fp, err := g.storage.Create("config.yaml")
	if err != nil {
		g.onPanic(fmt.Errorf("无法创建配置文件"))
		return plainConfigs
	} else {
		fp.Close()
	}

	//fp, err := os.OpenFile(g.configPath, os.O_RDONLY|os.O_CREATE, 0644)
	//if err != nil {
	//	g.onPanic(fmt.Errorf("无法读取配置文件: %v,请检查权限或尝试手动删除该文件", g.configPath))
	//}
	//byteConfig, err := ioutil.ReadAll(fp)
	//if err != nil {
	//	g.onPanic(fmt.Errorf("无法读取配置文件: %v,请检查权限或尝试手动删除该文件", g.configPath))
	//}
	//fp.Close()
	//plainConfigs := make([]*config.SessionConfigWithName, 0)
	//err = yaml.Unmarshal(byteConfig, &plainConfigs)
	//if err != nil {
	//	g.onPanic(fmt.Errorf("无法解析配置文件: %v,文件已损坏,请尝试手动删除该文件", g.configPath))
	//}
	return plainConfigs
}

func (g *GUI) WriteBackConfigFile() {
	plainConfigs := make([]*config.SessionConfigWithName, 0)
	for _, c := range g.configs {
		plainConfigs = append(plainConfigs, c)
	}
	hasFileFlag := false
	for _, fn := range g.storage.List() {
		if fn == "config.yaml" {
			hasFileFlag = true
			break
		}
	}
	var fp fyne.URIWriteCloser
	var err error
	if hasFileFlag {
		fp, err = g.storage.Save("config.yaml")
	} else {
		fp, err = g.storage.Create("config.yaml")
	}

	//if err != nil {
	//	return
	//}
	//g.storage.Remove("config.yaml")
	//fp, err := g.storage.Create("config.yaml")
	if err != nil {
		g.onPanic(fmt.Errorf("无法打开配置文件\n %v\n请检查权限或尝试手动删除该文件", fp.URI().String()))
		return
	}
	//os.Rename(g.configPath, g.configPath+".bak")
	//fp, err := os.OpenFile(g.configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	//defer fp.Close()
	//if err != nil {
	//	g.onPanic(fmt.Errorf("无法打开配置文件: %v,请检查权限或尝试手动删除该文件", g.configPath))
	//}
	outBytes, err := yaml.Marshal(plainConfigs)
	if err != nil {
		g.onPanic(fmt.Errorf("无法序列化配置信息: %v,请联系开发者", plainConfigs))
		return
	}
	_, err = fp.Write(outBytes)
	if err != nil {
		g.onPanic(fmt.Errorf("无法写入配置文件\n %v\n请检查权限或尝试手动删除该文件", fp.URI().String()))
		return
	}
}

func (g *GUI) GetContent(setContent func(v fyne.CanvasObject), getContent func() fyne.CanvasObject, masterWindow fyne.Window, app fyne.App) fyne.CanvasObject {
	g.onPanic = func(err error) {
		dialog.ShowError(fmt.Errorf("发生了严重错误，程序即将退出：\n\n%v", err), masterWindow)
		// os.Exit(-1)
	}
	g.masterWindow = masterWindow
	g.setContent = setContent
	g.getContent = getContent
	g.app = app

	for _, c := range g.ReadConfigFile() {
		g.newConfig(c)
	}
	g.updateEntryIndex()

	newProfileBtn := &widget.Button{
		Text:          "添加新登录配置",
		OnTapped:      g.onNewProfile,
		Icon:          theme.ContentAddIcon(),
		IconPlacement: widget.ButtonIconLeadingText,
		Importance:    widget.HighImportance,
	}
	g.content = container.NewBorder(
		nil,
		newProfileBtn,
		nil,
		nil,
		g.makeProfilesList(),
	)
	return g.content
}

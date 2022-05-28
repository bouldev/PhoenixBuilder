package mainframe

import (
	"fmt"
	"os"
	"phoenixbuilder/omega/components"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/mainframe/upgrade"
	"phoenixbuilder/omega/utils"
	"strings"

	"github.com/pterm/pterm"
)

func (o *Omega) bootstrapDirs() {
	o.storageRoot = "omega_storage"
	// android
	if utils.IsDir("/sdcard/Download/omega_storage") {
		o.storageRoot = "/sdcard/Download/omega_storage"
	} else {
		if utils.IsDir("/sdcard") {
			if err := utils.MakeDirP("/sdcard/Download/omega_storage"); err == nil {
				o.storageRoot = "/sdcard/Download/omega_storage"
			}
		}
	}
	if o.storageRoot == "/sdcard/Download/omega_storage" {
		fmt.Println("您似乎在使用安卓手机，Omega的配置和数据将被保存到 /sdcard/Download/omega_storage")
	}
	if !utils.IsDir(o.storageRoot) {
		fmt.Println("创建数据文件夹 " + o.storageRoot)
		if err := utils.MakeDirP(o.storageRoot); err != nil {
			panic(err)
		}
	}
	o.readConfig()
	dataDir := o.GetPath("data")
	if !utils.IsDir(dataDir) {
		fmt.Println("创建数据文件夹: " + dataDir)
		if err := utils.MakeDirP(dataDir); err != nil {
			panic(err)
		}
	}
	logDir := o.GetPath("logs")
	if !utils.IsDir(logDir) {
		fmt.Println("创建日志文件夹: " + logDir)
		if err := utils.MakeDirP(logDir); err != nil {
			panic(err)
		}
	}
	noSqlDir := o.GetPath("noSQL")
	if !utils.IsDir(noSqlDir) {
		fmt.Println("创建非关系型数据库文件夹: " + noSqlDir)
		if err := utils.MakeDirP(noSqlDir); err != nil {
			panic(err)
		}
	}
	worldsDir := o.GetPath("worlds")
	if !utils.IsDir(worldsDir) {
		fmt.Println("创建镜像存档文件夹: " + worldsDir)
		if err := utils.MakeDirP(worldsDir); err != nil {
			panic(err)
		}
	}
}

func (o *Omega) bootstrapComponents() (success bool) {
	success = false
	defer func() {
		r := recover()
		if r != nil {
			success = false
			pterm.Error.Printf("组件配置文件不正确，因此 Omega 系统拒绝启动，具体错误如下:\n%v\n", r)
		}
	}()
	total := len(o.ComponentConfigs)
	// coreComponentsLoaded := map[string]bool{}
	corePool := getCoreComponentsPool()
	builtInPool := components.GetComponentsPool()
	// for n, _ := range corePool {
	// 	coreComponentsLoaded[n] = false
	// }
	for i, cfg := range o.ComponentConfigs {
		I := i + 1
		Name := cfg.Name
		Version := cfg.Version
		Source := cfg.Source
		if cfg.Disabled {
			o.backendLogger.Write(pterm.Warning.Sprintf("\t跳过加载组件 %3d/%3d [%v] %v@%v", I, total, Source, Name, Version))
			continue
		}
		o.backendLogger.Write(pterm.Success.Sprintf("\t正在加载组件 %3d/%3d [%v] %v@%v", I, total, Source, Name, Version))
		var component defines.Component
		if Source == "Core" {
			if componentFn, hasK := corePool[Name]; !hasK {
				o.backendLogger.Write("没有找到核心组件: " + Name)
				panic("没有找到核心组件: " + Name)
			} else {
				// coreComponentsLoaded[Name] = true
				_component := componentFn()
				_component.SetSystem(o)
				component = _component
			}
		} else if Source == "Built-In" {
			if componentFn, hasK := builtInPool[Name]; !hasK {
				o.backendLogger.Write("没有找到内置组件: " + Name)
				panic("没有找到内置组件: " + Name)
			} else {
				component = componentFn()
			}
		}
		component.Init(cfg)
		component.Inject(NewBox(o, Name))
		o.Components = append(o.Components, component)
	}
	// for n, l := range coreComponentsLoaded {
	// 	if !l {
	// 		panic(fmt.Errorf("核心组件 (Core) 必须被加载, 但是 %v 被配置为不加载", n))
	// 	}
	// }
	return true
}

func (o *Omega) Bootstrap(adaptor defines.ConnectionAdaptor) {
	fmt.Println("开始配置升级检测")
	upgrade.Upgrade()
	fmt.Println("开始预处理任务")
	o.bootstrapDirs()
	o.adaptor = adaptor
	o.uqHolder = adaptor.GetInitUQHolderCopy()
	o.backendLogger = &BackEndLogger{
		loggers: []defines.LineDst{
			o.GetLogger("后台信息.log"),
			utils.NewIONormalLogger(os.Stdout),
		},
	}
	o.redAlertLogger = &BackEndLogger{
		loggers: []defines.LineDst{
			o.backendLogger,
			o.GetLogger("security_event.log"),
			&FuncsToLogger{GetFns: func() []func(info string) {
				return o.redAlertHandlers
			}},
		},
	}
	o.backendLogger.Write("日志系统已可用,正在激活主框架...")
	o.backendLogger.Write("加载组件中...")
	o.Reactor.onBootstrap()
	if o.bootstrapComponents() == false {
		o.Stop()
		return
	}
	//o.backendLogger.Write("组件全部加载&配置完成, 正在将更新后的配置写回配置文件...")
	//o.writeBackConfig()
	o.configStageComplete()
	o.backendLogger.Write("启用 Game Ctrl 模块")
	o.GameCtrl = newGameCtrl(o)

	o.backendLogger.Write("开始激活组件并挂载后执行任务...")
	for _, component := range o.Components {
		c := component
		o.CloseFns = append(o.CloseFns, func() error {
			return c.Stop()
		})
		go component.Activate()
	}
	//fmt.Println(o.CloseFns)
	o.backendLogger.Write("全部完成，系统启动")
	for _, p := range o.uqHolder.PlayersByEntityID {
		for _, cb := range o.Reactor.OnFirstSeePlayerCallback {
			cb(p.Username)
		}
	}
	{
		logo := GetLogo(LOGO_BOTH)
		//banner := []string{
		//	"┌───────────────────────────────────────────────────────────────────────┐",
		//	"|   ██████  ███    ███ ███████  ██████   █████      ███    ██  ██████   |",
		//	"|  ██    ██ ████  ████ ██      ██       ██   ██     ████   ██ ██        |",
		//	"|  ██    ██ ██ ████ ██ █████   ██   ███ ███████     ██ ██  ██ ██   ███  |",
		//	"|  ██    ██ ██  ██  ██ ██      ██    ██ ██   ██     ██  ██ ██ ██    ██  |",
		//	"|   ██████  ██      ██ ███████  ██████  ██   ██     ██   ████  ██████   |",
		//	"└───────────────────────────────────────────────────────────────────────┘",
		//}
		fmt.Println(strings.Join(logo, "\n"))
	}
	pterm.Success.Println("OMEGA_ng 等待指令")
	pterm.Success.Println("输入 ? 以获得帮助")
}

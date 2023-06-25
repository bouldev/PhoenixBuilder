package mainframe

import (
	"archive/zip"
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"phoenixbuilder/omega/collaborate"
	"phoenixbuilder/omega/components"
	"phoenixbuilder/omega/defines"
	third_party_omega_components "phoenixbuilder/omega/third_party"
	"phoenixbuilder/omega/utils"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

func CompressLog(srcFile string, dstFile string, zipWriter *zip.Writer, startThres time.Time, stopThres time.Time) (lineCompressed, lineRetained int, err error) {
	const TIME_LAYOUT = "2006/01/02"
	// first, determine whether the file need to be compressed
	var info os.FileInfo
	var reader *bufio.Reader
	var currentLine []byte
	var zipFileWriter io.Writer
	var fp *os.File
	defer func() {
		if fp != nil {
			fp.Close()
		}
	}()
	// get file info
	{
		info, err = os.Stat(srcFile)
		if err != nil {
			return
		}
		fp, err = os.OpenFile(srcFile, os.O_RDONLY, 0755)
		if err != nil {
			return
		}
		reader = bufio.NewReader(fp)
	}
	// get log start time
	{
		firstLine, err := reader.ReadBytes('\n')
		if err != nil {
			return 0, 0, nil
		}
		currentLine = firstLine

		if len(firstLine) < 10 || firstLine[4] != '/' || firstLine[7] != '/' {
			return 0, 0, nil
		}
		possibleDataInfo := firstLine[:10]
		startTime, err := time.Parse(TIME_LAYOUT, string(possibleDataInfo))
		if err != nil {
			return 0, 0, nil
		}
		if startTime.After(startThres) {
			return 0, 0, nil
		}
	}
	// create zip file entry in zip
	{
		var header *zip.FileHeader
		header, err = zip.FileInfoHeader(info)
		if err != nil {
			return
		}
		header.Name = srcFile
		header.Name = strings.ReplaceAll(header.Name, "\\", "/")
		header.Method = zip.Deflate
		zipFileWriter, err = zipWriter.CreateHeader(header)
	}
	// zip file
	{
		fmt.Printf("正在压缩日志文件: %v\n", srcFile)
		fastDeterminCache := []byte{}
		lineCompressed++
		zipFileWriter.Write(currentLine)
		for {
			currentLine, err = reader.ReadBytes('\n')
			if err != nil {
				return lineCompressed, lineRetained, nil
			}
			if len(currentLine) >= 10 && currentLine[4] == '/' && currentLine[7] == '/' {
				possibleDataInfo := currentLine[:10]
				if bytes.Equal(fastDeterminCache, possibleDataInfo) {
					lineCompressed++
					zipFileWriter.Write(currentLine)
				} else {
					startTime, err := time.Parse(TIME_LAYOUT, string(possibleDataInfo))
					if err != nil {
						lineCompressed++
						zipFileWriter.Write(currentLine)
					}
					if startTime.After(stopThres) {
						break
					} else {
						fastDeterminCache = possibleDataInfo
						lineCompressed++
						zipFileWriter.Write(currentLine)
					}
				}
			} else {
				lineCompressed++
				zipFileWriter.Write(currentLine)
			}
		}
		{
			var dstFp *os.File
			if dstFp, err = os.OpenFile(dstFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755); err != nil {
				return
			}
			lineRetained++
			dstFp.Write(currentLine)
			for {
				currentLine, err = reader.ReadBytes('\n')
				if err != nil {
					dstFp.Close()
					return lineCompressed, lineRetained, nil
				}
				lineRetained++
				dstFp.Write(currentLine)
			}

		}
	}
}

func CompressLogs(root string, startThres, endThres int) error {
	totalCompressed := 0
	totalRetained := 0
	var zipWriter *zip.Writer
	workDir := path.Join(root, "日志压缩临时目录")
	logRoot := path.Join(root, "logs")
	zipDir := path.Join(root, "日志压缩")
	if _, err := os.Stat(logRoot); err != nil {
		return nil
	}
	os.RemoveAll(workDir)
	os.MkdirAll(workDir, 0755)
	defer func() {
		os.RemoveAll(workDir)
	}()

	today := time.Now().Truncate(24 * time.Hour)
	fp, err := os.OpenFile(path.Join(workDir, "压缩中.zip.tmp"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	zipWriter = zip.NewWriter(fp)

	fileInfos, err := ioutil.ReadDir(logRoot)
	if err != nil {
		return err
	}
	filesToMove := []string{}
	srcFileToRemove := []string{}
	var origSize int64
	for _, info := range fileInfos {
		if info.IsDir() {
			continue
		} else {
			fileName := info.Name()
			if strings.HasSuffix(fileName, ".log") {
				if lineCompressed, lineRetained, err := CompressLog(
					path.Join(logRoot, fileName), path.Join(workDir, fileName),
					zipWriter, today.AddDate(0, 0, -startThres), today.AddDate(0, 0, -endThres)); err != nil {
					fmt.Println(err)
				} else {
					if lineCompressed != 0 || lineRetained != 0 {
						totalCompressed += lineCompressed
						totalRetained += lineRetained
						if lineRetained > 0 {
							filesToMove = append(filesToMove, fileName)
						} else {
							srcFileToRemove = append(srcFileToRemove, fileName)
						}
						origSize += info.Size()
					}
				}
			}
		}
	}
	zipWriter.Close()

	if totalCompressed == 0 {
		return nil
	}
	fileInfos, err = ioutil.ReadDir(workDir)
	if err != nil {
		return err
	}
	compressedSize := int64(0)
	for _, info := range fileInfos {
		compressedSize += info.Size()
	}
	origSizef := float32(origSize) / 1024 / 1024
	compressedSizef := float32(compressedSize) / 1024 / 1024
	os.MkdirAll(zipDir, 0755)
	if err := os.Rename(path.Join(workDir, "压缩中.zip.tmp"), path.Join(zipDir, "截止到"+today.AddDate(0, 0, -endThres).Format("2006_01_02")+"的日志.zip")); err != nil {
		return err
	}
	for _, fileName := range filesToMove {
		if err := os.Rename(path.Join(workDir, fileName), path.Join(logRoot, fileName)); err != nil {
			return err
		}
	}
	for _, fileName := range srcFileToRemove {
		os.Remove(path.Join(logRoot, fileName))
	}
	if totalCompressed > 0 {
		pterm.Success.Printf("共计压缩 %v 行日志(超出%v天的日志), 保留 %v 行日志(%v天内的日志) \n原始文件大小 %.1f MB 压缩后日志总大小 %.1f MB, 节约空间 %.1f MB, 比率 %.1f%%\n",
			totalCompressed, endThres, totalRetained, endThres, origSizef, compressedSizef, origSizef-compressedSizef, compressedSizef*100/origSizef)
	}

	return nil
}

func (o *Omega) bootstrapRootDir() string {
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
	return o.storageRoot
}

func (o *Omega) bootstrapDirs() {
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
	omegaSideDir := o.GetPath("side")
	if !utils.IsDir(omegaSideDir) {
		fmt.Println("创建 Omega Side 文件夹: " + omegaSideDir)
		if err := utils.MakeDirP(omegaSideDir); err != nil {
			panic(err)
		}
	}
	omegaCacheDir := o.GetOmegaCacheDir()
	if utils.IsDir("omega_cache") {
		os.RemoveAll("omega_cache")
	}
	if !utils.IsDir(omegaCacheDir) {
		fmt.Println("创建 Omega Cache 文件夹: " + omegaSideDir)
		if err := utils.MakeDirP(omegaCacheDir); err != nil {
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
			pterm.Error.Printf("正在加载的组件配置文件不正确，因此 Omega 系统拒绝启动，具体错误如下:\n%v\n建议根据说明修改对应的配置文件，如果你修不好了，删除对应配置文件即可\n", r)
		}
	}()
	total := len(o.ComponentConfigs)
	corePool := getCoreComponentsPool()
	coreComponentsLoaded := map[string]bool{}
	for n, _ := range corePool {
		coreComponentsLoaded[n] = false
	}
	builtInPool := components.GetComponentsPool()
	thirdPartPool := make(map[string]defines.Component)
	thirdPartNameSpace := make(map[string]string, 0)
	for _, g := range third_party_omega_components.GetAllThirdPartComponents() {
		for n, c := range g.Components {
			thirdPartPool[n] = c
			thirdPartNameSpace[n] = string(g.NameSpace)
		}
	}
	beforeActivateFNs := map[string]func() error{}
	for i, cfg := range o.ComponentConfigs {
		nameSpace := ""
		I := i + 1
		Name := cfg.Name
		Version := cfg.Version
		Source := cfg.Source
		NameShort := cfg.Name
		SourceShort := Source
		if strings.HasPrefix(SourceShort, "第三方::") {
			SourceShort = "第三方"
			NameShort = strings.Replace(NameShort, "第三方::", "", 1)
		}
		if cfg.Disabled {
			o.backendLogger.Write(pterm.Warning.Sprintf("\t跳过加载组件 %3d/%3d [%v] %v@%v", I, total, SourceShort, NameShort, Version))
			continue
		}
		o.backendLogger.Write(pterm.Success.Sprintf("\t正在加载组件 %3d/%3d [%v] %v@%v", I, total, SourceShort, NameShort, Version))
		var component defines.Component
		if Source == "Core" {
			if componentFn, hasK := corePool[Name]; !hasK {
				o.backendLogger.Write("没有找到核心组件: " + Name)
				panic("没有找到核心组件: " + Name)
			} else {
				coreComponentsLoaded[Name] = true
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
		} else if Source == "Lua-Component" {
			//跳过交给lua组件处理
		} else if strings.HasPrefix(Source, "第三方::") {
			if _component, hasK := thirdPartPool[Name]; !hasK {
				pterm.Error.Println("没有找到第三方组件: " + Name)
				continue
			} else {
				component = _component
				nameSpace = thirdPartNameSpace[Name]
			}
		}
		box := NewBox(o, Name, nameSpace)
		component.Init(cfg, box)
		component.Inject(box)
		o.Components = append(o.Components, component)
		beforeActivateFNs[cfg.Name] = component.BeforeActivate
	}
	for name, beforeActivate := range beforeActivateFNs {
		err := beforeActivate()
		if err != nil {
			panic(fmt.Errorf("组件: %v 激活前任务出现错误: %v", name, err))
		}
	}
	for n, l := range coreComponentsLoaded {
		if !l {
			panic(fmt.Errorf("核心组件 (Core) 必须被加载, 但是 %v 被配置为不加载", n))
		}
	}
	return true
}

func (o *Omega) initContext() {
	var func1 collaborate.GEN_STRING_LIST_HINT_RESOLVER = func(available []string) (string, func(params []string) (selection int, cancel bool, err error)) {
		return collaborate.GenStringListHintResolver(available)
	}
	var func2 collaborate.GEN_STRING_LIST_HINT_RESOLVER_WITH_INDEX = func(_available []string) (string, func(params []string) (selection int, cancel bool, err error)) {
		return collaborate.GenStringListHintResolverWithIndex(_available)
	}
	var func3 collaborate.GEN_INT_RANGE_RESOLVER = func(min, max int) (string, func(params []string) (selection int, cancel bool, err error)) {
		return collaborate.GenIntRangeResolver(min, max)
	}
	var func4 collaborate.GEN_YES_NO_RESOLVER = func() (string, func(params []string) (bool, error)) {
		return collaborate.GenYesNoResolver()
	}
	var func5 collaborate.QUERY_FOR_PLAYER_NAME = func(src, dst string, searchFn collaborate.FUNCTYPE_GET_POSSIBLE_NAME) (name string, cancel bool) {
		return collaborate.QueryForPlayerName(o.GameCtrl, src, dst, searchFn)
	}
	o.SetGlobalContext(collaborate.INTERFACE_GEN_STRING_LIST_HINT_RESOLVER, func1)
	o.SetGlobalContext(collaborate.INTERFACE_GEN_STRING_LIST_HINT_RESOLVER_WITH_INDEX, func2)
	o.SetGlobalContext(collaborate.INTERFACE_GEN_INT_RANGE_RESOLVER, func3)
	o.SetGlobalContext(collaborate.INTERFACE_GEN_YES_NO_RESOLVER, func4)
	o.SetGlobalContext(collaborate.INTERFACE_QUERY_FOR_PLAYER_NAME, func5)
}

//go:embed update.log
var updateInfo []byte

func (o *Omega) Bootstrap(adaptor defines.ConnectionAdaptor) {
	if len(updateInfo) > 10 {
		pterm.DefaultBox.WithTitle(pterm.FgBlue.Sprintf("更新日志(最近 20 次更新)")).WithTitleBottomRight().WithRightPadding(0).WithBottomPadding(0).Println(string(updateInfo))
	}
	rootDir := o.bootstrapRootDir()
	fmt.Printf("根目录为: %v， 开始分配存储目录\n", rootDir)
	o.bootstrapDirs()
	o.QuerySensitiveInfoFN = adaptor.QuerySensitiveInfo
	o.adaptor = adaptor
	o.uqHolder = adaptor.GetInitUQHolderCopy()
	o.Reactor.scoreboardHolder = defines.NewScoreBoardHolder(o.uqHolder)
	fmt.Println("开始空间回收任务: 日志压缩")
	CompressLogs(o.storageRoot, 7, 3)

	o.backendLogger = &BackEndLogger{
		loggers: []defines.LineDst{
			o.GetLogger("后台信息.log"),
			utils.NewIOColorTranslateLogger(os.Stdout),
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
	o.backendLogger.Write("日志系统已可用,开始激活主框架")

	timeLocal := time.FixedZone("CST", 3600*8)
	time.Local = timeLocal
	o.backendLogger.Write("固定时区为 CST +8:00")

	o.backendLogger.Write("启用 Omega Reactor 反应核心")
	o.Reactor.onBootstrap()

	o.backendLogger.Write("读取配置文件中...")
	o.checkAndLoadConfig()

	o.backendLogger.Write("启用 Game Ctrl 模块")
	o.GameCtrl = newGameCtrl(o)

	o.backendLogger.Write("初始化 Context ...")
	o.initContext()

	o.backendLogger.Write("初始化组件 ...")

	if !o.bootstrapComponents() {
		o.Stop()
		return
	}

	o.backendLogger.Write("激活组件并挂载组件 Post Process Task ...")
	for _, component := range o.Components {
		c := component
		o.CloseFns = append(o.CloseFns, func() error {
			return c.Stop()
		})
		go component.Activate()
	}
	o.backendLogger.Write("全部完成，系统启动")
	for _, p := range o.uqHolder.PlayersByEntityID {
		for _, cb := range o.Reactor.OnKnownPlayerExistCallback {
			cb(p.Username)
		}
	}
	fmt.Println(strings.Join(GetLogo(LOGO_BOTH), "\n"))
	pterm.Success.Println("OMEGA_ng 等待指令")
	pterm.Success.Println("输入 ? 以获得帮助")
	if len(updateInfo) > 10 {
		pterm.Success.Println("您可在最上方看到更新信息")
	}
}

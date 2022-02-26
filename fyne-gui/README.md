# ~~PhoenixBuilder.3rd.GUI~~
# PhoenixBuilder.fyne.GUI

## 说明
~~PhoenixBuilder.3rd.GUI 是第三方开发者开发的套壳 FastBuilder~~
PhoenixBuilder.fyne.GUI 是使用 fyne 框架的套壳 Fastbuilder
提供了一个带有界面的，完全跨平台的图形化Fastbuilder  
其核心来自 Fastbuilder https://github.com/LNSSPsd/PhoenixBuilder  
图形界面/跨平台编译技术来自 Fyne: https://github.com/fyne-io/fyne    

~~PhoenixBuilderHeadless 是原FB项目的无头版本，尽量减少对原项目的修改：  
https://github.com/LNSSPsd/PhoenixBuilder~~

~~没办法，为了能让项目在安卓上编译，不得不修改fastbuilder->现在的fb文件夹~~
使用 go build tag 提供的选择性编译实现 cli/gui 版本的切换
PhoenixBuilder GPLv3 协议，项目的核心:  
https://github.com/LNSSPsd/PhoenixBuilder  
根据协议要求，本项目同为 GPL v3 协议    
除了该项目外，本项目：
- 使用了 go-raknet 源代码MIT协议  
https://github.com/Sandertv/go-raknet 
- 使用了gophertunnel 源代码完成数据包解析等操作，MIT协议  
https://github.com/Sandertv/gophertunnel
- 使用了 dragonfly MC 服务器框架源代码完成对MC服务器的模拟  
https://github.com/df-mc/dragonfly  
- 字体来自 Consolas-with-Yahei （因为需要跨平台，所以内嵌了20+MB的字体文件）  
https://github.com/crvdgc/Consolas-with-Yahei 

## 运行
__切换到该目录下__  
你可以很简单的使用
```
go build -tags fyne_gui main.go
```
编译出对应平台的程序

## 匹配版本号和哈希值
不解释，具体来说，需要修改：
go.mod （这个文件夹下） 第5行
fustbuilder 主目录下 dedicate/fyne/session/version 5～6行

## 编译发行版
首先你需要安装必须的工具  
```
go get fyne.io/fyne/v2/cmd/fyne
go install fyne.io/fyne/v2/cmd/fyne
```

### 对于Windows/Linux/Mac

```
fyne package -os linux -tags fyne_gui
fyne package -os windows -tags fyne_gui
fyne package -os darwin -tags fyne_gui
```

### 重要！！！：
遗憾的是 fyne，fyne-cross 对 tags 的支持有问题，无法编译安卓和ios，因此我们不得不简单模拟 go build tag 的效果
运行 python3 prepare_build.py 将会创建一个临时目录，后续操作都在该目录中进行

### 对于 android：
1. 准备环境，ndk，adb，并设置环境变量 ANDROID_NDK_HOME
2. 编译（windows上似乎无法正常工作）
```
fyne package -os android/arm64 -appID fastbuilder.fyne.gui
```
3. 安装测试
```
fyne install -os android
```

## 另一种编译方式(fyne-cross)
安装环境和工具
```
go get github.com/fyne-io/fyne-cross
go install github.com/fyne-io/fyne-cross
```
安装 docker，并想办法确保网络连接
编译 (输出在 fyne-cross/dist 目录下)
运行
```
bash fyne_cross_compile.sh
```
编译全部
```
Linux:
fyne-cross linux -arch=amd64 -app-build 169 -app-id "fastbuilder.fyne.gui" -app-version 0.0.4 -icon unbundled_assets/Icon.png  -name "fastbuilder.fyne.gui"

MacOS:
fyne-cross darwin -arch=amd64 -app-build 169 -app-id "fastbuilder.fyne.gui" -app-version 0.0.4 -icon unbundled_assets/Icon.png  -name "fastbuilder.fyne.gui"

Windows:
fyne-cross windows -arch=amd64 -app-build 169 -app-id "fastbuilder.fyne.gui" -app-version 0.0.4 -icon unbundled_assets/Icon.png  -name "fastbuilder_gui.exe"

Android:
fyne-cross android -arch=arm64 -app-build 169 -app-id "fastbuilder.fyne.gui" -app-version 0.0.4 -icon unbundled_assets/Icon.png  -name "fastbuilder.fyne.gui"

IOS:
你需要创建一个开发者账号，并建立一个同名 Xcode项目 "fastbuilder_gui" 且保证 bundle identifier 为 "fastbuilder.fyne.gui" 接着
fyne-cross ios -app-build 169 -app-id "fastbuilder.fyne.gui" -app-version 0.0.4 -icon unbundled_assets/Icon.png  -name "fastbuilder-gui"
```

### 2022.2.25补充
现在配置好环境后，输入
```
bash fyne_cross_compile.sh
```
即可自动打包全平台的分发了

### 更多
参考  
https://developer.fyne.io/started/cross-compiling   
https://developer.fyne.io/started/packaging   
的编译说明

## 致谢
感谢 Ruphane 在该程序开发和测试中的帮助  
感谢 CodePwn 帮忙测试和反馈问题  
以及 fyne 库的开发者
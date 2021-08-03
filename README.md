# Fast Builder Phoenix 

## 简介

> Fast Builder 是一个为Minecraft bedrock设计的建筑辅助程序，可以在多个平台使用。Phoenix版本专门为中国版租赁服设计。目前支持的功能有欧几里得几何和Schematic（方块NBT数据丢弃）建筑生成，以及图像绘制。

### 原理

​	本次Fast Builder 采用了全新的技术，不再受限于 Websocket，因此在速度，性能和稳定性等方面有了巨大提升，并且拥有极大的可扩展性。当前Fast Builder的核心基于Sandertv的[Gophertunnel](https://github.com/Sandertv/gophertunnel/) （Gophertunnel是开源的，MIT协议）。

### 源代码开放

​	FastBuilder Phoenix 部分代码已开源（不包括验证系统和部分核心算法），源代码点击 [此处](https://fastbuilder.pro/source.tar.gz) 可见（GPL v3协议）。

### 注意事项（购买须知）

- 购买即代表您同意并且会遵守Fast Builder [**用户协议**](LICENSE.html)
- 任何技术都是有**时效性**的，Fast Builder可能随时会失效。
- 您购买的Fast Builder（Phoenix，发布于**2021 - 07 - 21**； 当前能够正常使用）功能仅仅包括欧几里得几何和Schematic/ACME(mcacblock)建筑生成，以及图片绘制，未来可能会更新其他功能，但我们无法保证。
- Fast Builder 部分安装，使用步骤需要一定的**计算机和数学知识**，并且安装教程默认您已经**具备**了这些知识。
- Fast Builder 个人账号只能绑定到2个租赁服，一个月可以进行一次更改，无法在**联机房间**或者**国际版**使用。
- 请勿使用Fast Builder Pro来导入**未经授权的著作**，每一个建筑作者都靠自己的体力及脑力劳动在社区生存，如果您使用其他人的建筑作品并加以盈利，会对整个游戏环境带来破坏。同时，如果建筑的著作权人追究责任，您应当且必须履行，并且我们不应为此承担任何责任。
- 开发者不是客服，他们没有义务**解决**你在使用软件过程中遇到的各种小问题，更不会**亲自**指导你安装。如果你在使用软件的过程中遇到了Bug并且**你非常确定**，可以向代理**提交Bug报告**。
- 当前版本(Phoenix Alpha)为测试版，可能出现各种bug，购买则表明**自愿承担风险**!
- **尽管程序设计成了全英文提示语，但不代表我们不是中国人**。
- 每个功能都尽量设计**完美**，但依然是存在**缺陷**的。

### 基本概念

​	Fast Builder与其他程序的不同点在于，存在“客户端”与“服务端”概念。玩家运行游戏的手机，或者说玩家正在运行的游戏，就是客户端；Fast Builder就是服务端。

​	因此，服务端可以与客户端不在同一台设备上，甚至没有客户端都能使用Fast Builder。

​	由于使用命令行操作，需要用户具备如下基础：

- 文件操作能力，能够理解路径和文件层级
- 英语能力，能够认出‘Error’、‘Permission denied’、‘not found’等字眼并明白其含义
- 能够区分全角符号与半角符号
- (最好会)在命令行界面输入并执行命令的能力

请确保您具备以上能力，如果您不具备这样的条件还执意要购买Fast Builder导致无法使用，开发组不承担任何责任，更不会提供帮助。

### 安装指导

#### 必要条件

- 拥有Fast Builder账号并且已经购买了Fast Builder

- 因设备而异：

  - 电脑端（Windows/Linux/macOS）：电脑需具备正常的网卡

  - 安卓Android:  拥有Termux,安装步骤见下文

- 勤劳的双手和善于思考的**大脑**

1. 登录Fast Builder[用户中心](uc.fastbuilder.pro) 
2. 进入Profile，设置游戏名（中国版用户名）
3. 设定您要使用Fast Builder的**租赁服**的号码（注意：该租赁服必须允许**任意等级**的玩家加入）
4. 设置辅助用户名称，并点击【生成】来创建**辅助用户**

以上是必要信息填写，接下来进入安装步骤，不同平台方案不同，请找到自己的平台：

#### 安装步骤

- windows: 直接在用户中心的【Download】下载

- linux(推荐): 

  ```shell
  wget -O fastbuilder https://fastbuilder.pro/downloads/phoenix/phoenixbuilder
  chmod +x fastbuilder
  ```

- android: 

  - a.点击[链接](https://f-droid.org/repo/com.termux_117.apk)安装Termux （**0.117**）; 

  - b.安装完成后，前往系统设置，给予**存储空间权限**且允许**无限制后台运行**。

  - c.打开Termux，复制并执行如下命令，第一步，先使用如下命令替换将官方源替换为TUNA镜像源（可选，此步骤可以让后续步骤更快完成）

    > **警告：三个命令要按序依次执行，不可以一次性直接复制粘贴； 该镜像仅适用于 Android 7.0 (API 24) 及以上版本，旧版本系统使用本镜像可能导致程序错误。如果您是旧版，请不要使用该镜像**

    ```shell
    sed -i 's@^\(deb.*stable main\)$@#\1\ndeb https://mirrors.tuna.tsinghua.edu.cn/termux/termux-packages-24 stable main@' $PREFIX/etc/apt/sources.list
    sed -i 's@^\(deb.*games stable\)$@#\1\ndeb https://mirrors.tuna.tsinghua.edu.cn/termux/game-packages-24 games stable@' $PREFIX/etc/apt/sources.list.d/game.list
    sed -i 's@^\(deb.*science stable\)$@#\1\ndeb https://mirrors.tuna.tsinghua.edu.cn/termux/science-packages-24 science stable@' $PREFIX/etc/apt/sources.list.d/science.list
    ```

    接下来执行更新源和软件包的命令（**必须**）：

    ```shell
    apt update -y
    apt upgrade -y
    apt install wget -y
    ```

   - d.待上述命令执行完成后，再下载Fast Builder (如果您是armv7架构，则将命令中的arm64改成armv7)

     > **注意：此步骤(d)也为更新步骤，之后更新Fast Builder可以直接通过执行本步骤完成。**

     ```shell
     wget -O fastbuilder https://fastbuilder.pro/downloads/phoenix/phoenixbuilder-android-executable-arm64
     chmod +x fastbuilder
     ```
### 使用指导

​	Fast Builder Phoenix是纯命令行程序，没有复杂的GUI，这使得程序本身非常易于使用。

#### 启动步骤

- windows: 双击Fast Builder可执行文件

- linux: 无需解释

- android: 打开termux， 执行以下命令:

  ```shell
  ./fastbuilder
  ```

#### 初始化程序

​	如果不出意外，在您完成了上述步骤，并且看到了Fast Builder成功启动的字样。第一次启动会要求用户输入Fast Builder账号密码，照做即可。完成后，Fast Builder会要求用户输入租赁服号码(Rental Server Number)和租赁服密码（没有可忽略）。如果没有报错退出，那么Fast Builder就成功启动了（第二次启动不会要求用户输入账号密码）。接下来，将FB挂在后台，进入网易租赁服，若看到辅助用户在线，那么Fast Builder就成功运行了。这个时候，请**给予辅助用户OP权限**。辅助用户只听从**（您）操作员**的命令，**操作员**的名称必须和用户中心所设置的**中国版用户名**一致，请勿使用带有**称号**的皮肤，否则辅助用户将**不会**听从您的命令！

#### Fast Builder命令解析

​	Fast Builder采用类似Linux Shell的命令系统（并非原版Minecraft的命令），无需使用【/】，直接在聊天窗口输入即可。如果您是Linux爱好者,那您将会很快掌握FB的命令。我们这次有机会做GUI，但是这显然没有CUI好用。

```shell
# 例子：生成半径为5,方向朝y轴的圆,以下两条命令是等价的
round -r 5 -f y -h 1
round --radius 5 --facing y --height 1
```

##### 生成器设置

​	成功初始化FB之后，我们将辅助用户（以下简称Bot）所在的维度称为一个【空间】，所有的操作都是在该空间中进行的。因此，如果您想在不同的【空间】中使用FB，您必须让Bot和您同处一个空间。

​	想要使用FB，必须设置【**空间**】的【**原点**】（**结构会在原点构造**），默认的原点在Bot进入游戏时所处的位置。

​	使用 `get` 命令可以修改原点设置为用户当前所处的位置，下面是相关功能命令的详解。

​	Fast Builder Phoenix使用了**协程**，这意味这您可以一次执行多个任务，您可以使用task命令管理任务。

##### Task命令

​	task命令用于管理当前的【**任务**】（每个**建筑进程**都被抽象为一个任务）。每个task都有自己的runtime，你可以用一些功能命令设定任务的内部参数。task命令具有如下几个基本子命令（#后面为注释，解释了子命令的作用）：

````shell
tasktype <type:async|sync> # Task类型，async不支持进度显示，一边计算一边发送数据，sync则是完成计算后发送，支持速度显示
task list # 列出所有任务的ID，内容以及状态
task pause <taskID> # 暂停某个任务
task resume <taskID> # 恢复某个被暂停的任务
task break <taskID> # 销毁某个任务（不可恢复，它将在列表中消失）
````

​	task命令还可以设定任务执行延迟(Delay)以及发包方案(Delay Mode):

```shell
task setdelay <taskID> <delay> # 设定某个任务的发包延迟。在continuous模式下单位为微秒； 在discrete模式下单位为秒
task setdelaymode <taskID> <delayMode:continuous|discrete|none> # 设定某个任务的发包方案，有3个可选参数
task setdelaythreshold <taskID> <threshold:int> # 设定某个任务的阈值（最大方块集合），仅在 discrete 方案下有效
```

- continuous（连续）：以一定速度发包，不会突变，比较稳定。
- discrete（离散）：每经过一段固定时间发送大量数据包（最大数量不大于阈值），突变，不稳定，有概率造成服务器崩溃。
- none（无延迟）：持续发送大量数据包，不会突变，非常不稳定，很大可能造成服务器崩溃。

三种方法各有优缺点（请斟酌使用）：
- 速度: continuous <= discrete < none (部分配置也可能导致continuous > discrete)
- 稳定性: continuous > discrete > none
- 平滑：continuous > none >= discrete 

##### 功能命令

- 设置空间原点：

  ```shell
  get
  ```

- 设置全局指令执行延迟和执行方案：

  ```shell
  delay mode <delayMode:continuous|discrete|none> # 设定默认的发包方案
  delay threshold <threshold:int> # 设定默认阈值（最大方块集合），仅在 discrete 方案下有效
  delay set <Delay> # 设定默认的发包延迟。在continuous模式下单位为微秒； 在discrete模式下单位为秒
  ```

- 设置是否显示进度条（显示建筑的进度百分比，方块总数，瞬时速度等信息。默认为true）

  ```shell
  progress <bool:true|false>
  ```

##### 几何命令

​	Fast Builder具备在空间中构造简单几何体的功能（如圆,圈,球,线,椭圆等）
- 圆面/圈 : 
  ```shell
  round/circle -r <radius> -f <facing> -h <height> -b <tileName> -d <data>
  -r --radius 圆或圈的半径
  -f --facing 朝向,可选值有x,y,z
  -b --block 方块名称
  -d --data 方块特殊值
  ```

- 球:
  ```shell
  sphere -r <radius> -s <shape>
  -s --shape hollow(空心) |solid(实心)
  ```

- 椭圆: 

  ```shell
  ellipse -l <length> -w <width> -f <facing>
  -l --length 长度
  -w --width 宽度
  ```

- 椭球: 

  ```shell
  ellipsoid -l <length> -w <width> -h <height>
  ```

##### 建筑生成命令

​	从schematic或mcacblock（ACME导出文件）文件加载建筑：

- -p --path 文件路径

  ```shell
  schem -p <filePath>
  acme -p <filePath>
  ```

##### 像素画生成

- 将图片加载到空间：

  ```shell
  plot -p <imageFilePath> -f <facing>
  ```
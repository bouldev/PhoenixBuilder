# FastBuilder.JS 编写指南

## 为什么需要 FastBuilder.JS
实际上，FastBuilder 不仅仅能实现导入导出等功能，对其稍加扩展便可实现包括但不限于以下的功能：
- 租赁服菜单
- 登录登出日志
- 聊天信息记录
- 群服互通
- 云黑

然而，对 FastBuilder 的修改和扩展要求用户至少需要一台电脑，一定程度的 golang 知识和 makefile 知识。不仅如此若希望为手机开发扩展，或是为其它平台开发扩展，您还需要 Linux 的电脑、知识以及交叉编译链的知识。您不得不分发编译出来的文件，并试图教会使用者如何部署。

我们希望 FastBuilder.JS 可以改善这一现状，让有兴趣开发 FastBuilder 潜能的人不再被设备限制，让使用者不再受部署问题的折磨。

简单来说，开发者现在只需要一部手机或者电脑，无需安装任何特殊软件即可使用 JavaScript/NodeJS 语言开发新的 Fastbuilder 功能，例如上述所有功能。

## 概述
我们在 Fastbuilder 中内置了著名的，最强大的 JavaScript V8 引擎，并提供了一系列接口以驱动 Fastbuilder 这意味这两件事:
- 你不需要学习新的语言，如果你有前端或是 NodeJS 开发经验，相同的脚本在 FastBuilder.JS 中有完全一致的效果
- 你无需研究该如何修改，编译，发布，部署你的脚本。你甚至只需打开手机上的文本编辑器，然后在 FastBuilder 中输入 script 你的脚本名 即可执行脚本

## 如何启动?
FastBuilder.JS 脚本既可以是一个本地文件，也可以是一个网址，FastBuilder 会自行尝试。

方法1: 在启动时运行
```
在命令行中输入:

./fastbuilder -S 你的脚本.js    
或者
./fastbuilder -S 脚本网址
```

方法2: 在启动后运行
```
在fb中输入:

script 你的脚本.js   
或者
script 脚本网址 
```

## 提供了哪些接口？
以下函数赋予了 js 脚本控制 fb 和 mc 的能力  
每个接口都有函数式和对象式两种方法，这两种方法是完全等价的(见 全部api和等价实现.js) 可根据自己喜好使用

- 设置脚本名（非必需）
```
// name 为脚本名
engine.setName(string name)
```

- 等待直到FB连接到服务器
```
// 注意，当以 ./fastbuilder -S 你的脚本.js 加载时，在这行代码之后，FB才会开始连接MC服务器
// 当以 script 你的脚本.js 加载时，脚本会卡在这一行，直到成功连接MC服务器
engine.waitConnectionSync()
```

- 当连接到服务器时，执行 onConnect 函数
```
// 注意，当以 ./fastbuilder -S 你的脚本.js 加载时，在这行代码之后，FB才会开始连接MC服务器
engine.waitConnection(function onConnect)
// function onConnect() {}
```

- 向用户显示一条信息
```
engine.message(string msg)
function engine.message(string msg)
```

- 请求用户输入信息
```
// 脚本将停止，直到用户输入完成
engine.questionSync(string hint)-> string userInput
```

- 请求用户输入信息，输入完成后调用 onResult 函数
```
// 脚本不会停止
engine.question(string hint,function onResult)
// function onResult(string result)
```

- 查询 FB 提供的信息  (警告，出于安全考虑，一些信息可能被移除)
```
// 请参考 example03，以下内容可以被查询
consts.script_sha256 脚本哈希值
consts.script_path 脚本所在路径 (可能被移除)
consts.engine_version JS解释器实现版本号 jsEngine.hostBridge.api
// 第一位表示 JS 引擎实现，第二位表示桥接器，第三位表示 api 版本，一般，只有api需要关心
consts.user_name 用户名 (可能被移除)
consts.sha_token 用户FB Token的哈希值
consts.server_code 服务器代码 (可能被移除)
consts.fb_version FB 版本信息
consts.fb_dir 工作路径(一般情况下就是fb所在路径) (可能被移除)
```

- 崩溃脚本
```
// 可以用该函数崩溃脚本
engine.crash(string reason)
```

- 启用自动重启
请自己实现自动重启功能，由脚本控制host是不好的～

- 等效于用户在fb中输入一条指令,例如 bdump, plot, task 等
```
game.eval(string fbCmd)
```

- 发送一条MC指令，不等待其返回值
```
game.oneShotCommand(string mcCmd)
```

- 发送一条MC指令，并等待其返回结果
```
// 注意，结果先在go中使用json序列化，然后在js中反序列化，所以并不高效
// 警告，一些指令没有返回值(如 say)，这种指令会导致程序卡死
game.sendCommandSync(string mcCmd) -> CommandOutput result
// 你可以 engine.message(JSON.stringify(result)) 来观察 CommandOutput 的结构
```

- 发送一条MC指令，当返回指令结果时，onResult 函数将被调用
```
game.sendCommand(string mcCmd,function onResult)
// function onResult(CommandOutput result)
```

- 获得机器人的位置
```
game.botPos() -> Pos p
// p.x, p.y, p.z 为坐标
```

- 订阅一种特定类型的数据包，当收到该种数据包时，函数 onPacket 将会被调用
```
game.subscribePacket(string packetType,function onPacket) -> function deRegFn
// 数据包先在go中使用json序列化，然后在js中反序列化，所以并不高效
// 警告，不合理的利用该函数可能导致性能低下
// deRegFn 的类型是 function，即在js中， deRegFn()会取消订阅这种数据包
// packetType 所有可用值附在文末
// function onPacket (Packet p){}
```

- 订阅聊天信息，当收到新聊天信息时，函数 onMsg 将会被调用
```
// 实际上可以通过 game.subscribePacket 实现，但是毕竟这种信息相当常用
game.listenChat(function onMsg)
// function onMsg(string name, string msg){}
```

- 请求一个文件夹的储存权限
```
// 向用户索要权限（需要玩家确认）
// 如果用户给了权限，第二次索要时不需要玩家确认，直接就能获得
fs.requestFilePermission(string hint, string path) -> bool isSuccess
```

- 加载文件现有内容, 内容总是以 string 形式返回
```
// 警告！如果尝试访问未授权文件夹，脚本会被强制停止
// 即使获取了文件夹权限，fbtoken等敏感文件也是禁止访问的（脚本会被强制停止）
fs.readFile(string path) -> string data
```

- 保存内容到文件，内容应该以 string 形式提供
```
fs.writeFile(string path, string data)
```

- websocket client
```javascript
// 请参考 example06_websocket.js
// address 为ws地址，当连接成功的时候 onNewMessage 函数会被调用
// class ws
let client=new ws(string address) -> Object ws
// client.send
// message: 信息
// msgType (Optional): 消息类型
client.send(String message, [Number msgType])
// client.close
client.close()
// 指示是否已连接到服务器
client.isConnecting (bool)
// 指示连接是否已经关闭
client.closed (bool)
// 用户指定的回调
client.onconnection (Object ws)
client.onopen (Object ws)
// 当client.onconnection已经指定时，client.onopen不再会被调用
client.onclose (String error)
client.onmessage (String content, int msgType)
client.onerror (String error)
```

- websocket server
```

```

## 一些内置的 JavaScript库
一些库已经内置在 FastBuilder.js, 主要涉及网络，加密和计时器，请参考 example07_网络和计时器及base64.js 、 example11_密码学.js   
其中，网络包括 fetch (fetch 可以很方便的实现 GET 和 POST 操作)，及 URL, base64(atob, btoa)
计时器包括 setTimeout, clearTimeout, setInterval and clearInterval （是的，这4个函数不属于 JS 的标准，是以内置库的方式出现的）
加密库我们直接内置了 Crypto.JS 的主要库，包括 aes,md5,rc4,sha256,tripledes,hmac-md5,hmac-256


## 例子？
参考 example01~13
## 其他
```
PacketType 可用值:

"IDLogin"
"IDPlayStatus"
"IDServerToClientHandshake"
"IDClientToServerHandshake"
"IDDisconnect"
"IDResourcePacksInfo"
"IDResourcePackStack"
"IDResourcePackClientResponse"
"IDText"
"IDSetTime"
"IDStartGame"
"IDAddPlayer"
"IDAddActor"
"IDRemoveActor"
"IDAddItemActor"
"IDTakeItemActor"
"IDMoveActorAbsolute"
"IDMovePlayer"
"IDRiderJump"
"IDUpdateBlock"
"IDAddPainting"
"IDTickSync"
"IDLevelEvent"
"IDBlockEvent"
"IDActorEvent"
"IDMobEffect"
"IDUpdateAttributes"
"IDInventoryTransaction"
"IDMobEquipment"
"IDMobArmourEquipment"
"IDInteract"
"IDBlockPickRequest"
"IDActorPickRequest"
"IDPlayerAction"
"IDHurtArmour"
"IDSetActorData"
"IDSetActorMotion"
"IDSetActorLink"
"IDSetHealth"
"IDSetSpawnPosition"
"IDAnimate"
"IDRespawn"
"IDContainerOpen"
"IDContainerClose"
"IDPlayerHotBar"
"IDInventoryContent"
"IDInventorySlot"
"IDContainerSetData"
"IDCraftingData"
"IDCraftingEvent"
"IDGUIDataPickItem"
"IDAdventureSettings"
"IDBlockActorData"
"IDPlayerInput"
"IDLevelChunk"
"IDSetCommandsEnabled"
"IDSetDifficulty"
"IDChangeDimension"
"IDSetPlayerGameType"
"IDPlayerList"
"IDSimpleEvent"
"IDEvent"
"IDSpawnExperienceOrb"
"IDClientBoundMapItemData"
"IDMapInfoRequest"
"IDRequestChunkRadius"
"IDChunkRadiusUpdated"
"IDItemFrameDropItem"
"IDGameRulesChanged"
"IDCamera"
"IDBossEvent"
"IDShowCredits"
"IDAvailableCommands"
"IDCommandRequest"
"IDCommandBlockUpdate"
"IDCommandOutput"
"IDUpdateTrade"
"IDUpdateEquip"
"IDResourcePackDataInfo"
"IDResourcePackChunkData"
"IDResourcePackChunkRequest"
"IDTransfer"
"IDPlaySound"
"IDStopSound"
"IDSetTitle"
"IDAddBehaviourTree"
"IDStructureBlockUpdate"
"IDShowStoreOffer"
"IDPurchaseReceipt"
"IDPlayerSkin"
"IDSubClientLogin"
"IDAutomationClientConnect"
"IDSetLastHurtBy"
"IDBookEdit"
"IDNPCRequest"
"IDPhotoTransfer"
"IDModalFormRequest"
"IDModalFormResponse"
"IDServerSettingsRequest"
"IDServerSettingsResponse"
"IDShowProfile"
"IDSetDefaultGameType"
"IDRemoveObjective"
"IDSetDisplayObjective"
"IDSetScore"
"IDLabTable"
"IDUpdateBlockSynced"
"IDMoveActorDelta"
"IDSetScoreboardIdentity"
"IDSetLocalPlayerAsInitialised"
"IDUpdateSoftEnum"
"IDNetworkStackLatency"
"IDScriptCustomEvent"
"IDSpawnParticleEffect"
"IDAvailableActorIdentifiers"
"IDNetworkChunkPublisherUpdate"
"IDBiomeDefinitionList"
"IDLevelSoundEvent"
"IDLevelEventGeneric"
"IDLecternUpdate"
"IDAddEntity"
"IDRemoveEntity"
"IDClientCacheStatus"
"IDOnScreenTextureAnimation"
"IDMapCreateLockedCopy"
"IDStructureTemplateDataRequest"
"IDStructureTemplateDataResponse"
"IDClientCacheBlobStatus"
"IDClientCacheMissResponse"
"IDEducationSettings"
"IDEmote"
"IDMultiPlayerSettings"
"IDSettingsCommand"
"IDAnvilDamage"
"IDCompletedUsingItem"
"IDNetworkSettings"
"IDPlayerAuthInput"
"IDCreativeContent"
"IDPlayerEnchantOptions"
"IDItemStackRequest"
"IDItemStackResponse"
"IDPlayerArmourDamage"
"IDCodeBuilder"
"IDUpdatePlayerGameType"
"IDEmoteList"
"IDPositionTrackingDBServerBroadcast"
"IDPositionTrackingDBClientRequest"
"IDDebugInfo"
"IDPacketViolationWarning"
"IDMotionPredictionHints"
"IDAnimateEntity"
"IDCameraShake"
"IDPlayerFog"
"IDCorrectPlayerMovePrediction"
"IDItemComponent"
"IDFilterText"
"IDClientBoundDebugRenderer"
"IDSyncActorProperty"
"IDAddVolumeEntity"
"IDRemoveVolumeEntity"
"IDNeteaseJson"
"IDPyRpc"
```
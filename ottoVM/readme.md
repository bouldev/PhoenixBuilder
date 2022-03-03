# FastBuilder.JS

## 说明
我们在 fast builder 中添加了 js 脚本支持(基于 otto)，扩展了 FB 的功能  
诸如：
  - 租赁服菜单
  - 登录登出
  - 聊天信息记录 
  - 群服互通

的实现不需要再修改fb的程序了，而可以简单的通过加载js脚本实现

## 如何启动?
方法1: 在启动时运行
```
./fastbuilder --script=你的脚本.js    
```
   
方法2: 在启动后运行
```
在fb中输入:

script 你的脚本.js   
```

## 提供了哪些函数？
以下函数赋予了 js 脚本控制 fb 和 mc 的能力
```
// 等待直到FB连接到服务器
function FB_WaitConnect() None

// 等效于用户在fb中输入一条指令
function FB_GeneralCmd(fbCmd string) None

// 发送一条MC指令，不等待其返回值
function FB_SendMCCmd(mcCmd string) None

// 发送一条MC指令，并等待其返回结果, 
// 注意，结果先在go中使用json序列化，然后在js中反序列化，所以并不高效
// 警告，一些指令没有返回值，这种指令会导致程序卡死
function FB_SendMCCmdAndGetResult(mcCmd string) object

// 订阅一种特定类型的数据包，当收到该种数据包时，指定的函数将会被调用，
// 数据包先在go中使用json序列化，然后在js中反序列化，所以并不高效
// 警告，不合理的利用该函数可能导致性能低下
// callbackFn 中this将被设为 undefined
// deRegFn 的类型是 function，即在js中， deRegFn()会取消订阅这种数据包
// packetType 所有可用值附在文末
function FB_RegPackCallBack(packetType string,callbackFn func(object)) deRegFn

// 订阅聊天信息
// 实际上可以通过 FB_regPackCallBack 实现，但是毕竟这种信息相当常用
function FB_RegChat(callBackFunc func(name string, msg string)) deRegFn

// 请求用户输入信息
function FB_RequireUserInput(hint string) string

// 向用户显示一条信息
function FB_Println(msg string) None

// 获得获取fb的某些信息，例如，用户的游戏名，无论是何种值，结果都以string形式返回
// "user_name" 玩家名
// "sha_token" 玩家 token sha256后的base64编码值
// 其他的暂时没想好
function FB_Query(info string) string

// 保存数据，出于安全考虑，fileName 不能包括 / 或 \\, 且 data 应该为string形式
function FB_SaveFile(fileName string, data string) None

// 读取数据，出于安全考虑，fileName 不能包括 / 或 \\, 且 data 为string形式
function FB_ReadFile(fileName string) string

// websocket 连接
// sendMessage 实现为 func(string) 可以用来主动发送数据
// onMessage在收到数据时会被调用
// 注意，无论是收还是发，都只接受text(string)类型的数据
function FB_websocketConnectV1(serverAddress string,onMessage func(string)) sendMessage

// 很遗憾，otto 只是js解释器，而没有 eventloop，我们必须自己实现一些常用功能
FB_setTimeout(function,delayInMillionSecond int)
// 类似的，我们还有以下三个功能
FB_setInterval
FB_clearTimeout
FB_clearInterval
```

## 示例脚本？
- example00.js  
本脚演示了时间记分板的校正   
演示了 FB_WaitConnect，FB_RequireUserInput，FB_Println，FB_SendMCCmd 的功能   
假设用户有一个记分板，记分板里有 year, month, day, hour, minute 四个项目   
需要与现实时间同步
```
// 等待连接到 MC
FB_WaitConnect()

// 请求用户输入信息 (时间相关记分板的名字)
scoreBoardName=FB_RequireUserInput("时间记分板的名字是?")

// js: 计算时间
nowTime=new Date()
nowYear=nowTime.getFullYear()
nowMonth=nowTime.getMonth()
nowDay=nowTime.getDay()
nowHour=nowTime.getHours()
nowMinute=nowTime.getMinutes()

// 发送指令
FB_SendMCCmd("scoreboard players set "+scoreBoardName+" year "+nowYear)
FB_SendMCCmd("scoreboard players set "+scoreBoardName+" month "+nowMonth)
FB_SendMCCmd("scoreboard players set "+scoreBoardName+" day "+nowDay)
FB_SendMCCmd("scoreboard players set "+scoreBoardName+" hour "+nowHour)
FB_SendMCCmd("scoreboard players set "+scoreBoardName+" minute "+nowMinute)

// 向用户发送提示信息
FB_Println("时间记分板校准完成！")
```

- example01.js  
本脚本演示了自动将机器人移动到玩家身边，并设置全局延迟为 100   
演示了 FB_Query，FB_SendMCCmdAndGetResult，FB_GeneralCmd 的功   
```
// 等待连接到 MC
FB_WaitConnect()

// 通用fb功能，相当于用户在fb中输入了这条指令
FB_GeneralCmd("delay set 100")

userName=FB_Query("user_name")

// 查看当前玩家有哪些，只是为了演示功能才那么做，其实没必要
listResult=FB_SendMCCmdAndGetResult("list")
currentPlayers=listResult["OutputMessages"][1]["Parameters"] // currentUsers [name1,name2,...]

displayStr="当前的玩家有: "
currentPlayers.forEach(function (playerName) {
    displayStr+=" "+playerName
})

FB_Println(displayStr)

if(userName in currentPlayers){
    FB_SendMCCmdAndGetResult("tp @s "+userName)
}else {
    FB_Println("看起来用户 "+userName+" 不在线耶")
}
```

- example02.js   
本脚本演示了一个菜单，并会在用户登录时主动发送提示信息  
演示了 FB_setTimeout，FB_RegPackCallBack，FB_GeneralCmd 的功能

```
// 当有新玩家时，一定会收到 IDPlayerList 数据包，现在我们从这个数据包中判断玩家是谁
function onPlayerListUpdate(pk){
    if (pk.ActionType!==0){
        // Action Type 为 0 时为玩家登录，否则为玩家退出
        return
    }
    pk.Entries.forEach(function (playerInfo){
        // player Info 包括了相当多的信息，我们只需要其中的名字即可
        playerName=playerInfo.Username
        // 值得注意的是，玩家刚上线时并不能看到消息，所以我们延迟 8 秒（8000ms）再显示
        FB_setTimeout(function () {
            FB_SendMCCmd("tellraw @a {\"rawtext\":[{\"text\":\"欢迎回来！ @"+ playerName +"\"}]}")
            FB_SendMCCmd("tellraw "+playerName+" {\"rawtext\":[{\"text\":\"试试在聊天栏输入 '菜单' ! \"}]}")
        },8000)
    })
}

// 告诉 FB，当有这个数据包时就执行上面的函数
FB_RegPackCallBack("IDPlayerList",onPlayerListUpdate)

// 实际上聊天功能基本就是一个问答机器人，接收聊天信息，并做出反应
FB_RegChat(function (name,msg) {
    if(name===""){
        // 不是人发出的聊天消息没有名字，比如命令块
        return
    }
    if (msg==="回城"){
        // 假设目的地是 0 100 0，这只是演示一下
        FB_SendMCCmd("tp "+name+" 0 100 0")
    }
    if (msg==="冒险"){
        // 假设目的地是 0 100 0，这只是演示一下
        FB_SendMCCmd("gamemode a "+name)
    }
    if (msg==="菜单"){
        // 假设目的地是 0 100 0，这只是演示一下
        FB_SendMCCmd("tellraw "+name+" {\"rawtext\":[{\"text\":\"输入 回城 以回到 0 100 0  \"}]}")
        FB_SendMCCmd("tellraw "+name+" {\"rawtext\":[{\"text\":\"输入 冒险 以切换为 冒险模式  \"}]}")
    }
})


// 等待连接到 MC
FB_WaitConnect()
```

- example03.js   
本脚本演示了一个日志功能，主要用来展示文件读写  
演示了 FB_setInterval，FB_ReadFile，FB_SaveFile 的功能


```
// 一般情况下，应该使用 Append，但是考虑到跨平台，有的系统无法提供append，故只提供
// Save/Read 功能
logData=FB_ReadFile("日志.txt")

FB_setInterval(function () {
    // 每隔十秒保存一次
    console.log("Save Log")
    FB_SaveFile("日志.txt",logData)
},10000)

// 添加一行记录
function LogString(info) {
    newDate = new Date();
    logData=logData+newDate.toLocaleString()+": "+info+"\n"
}

LogString("脚本启动")

// 记录聊天信息
FB_RegChat(function (name,msg) {
    LogString("chat: "+name+" :"+msg)
})

// 等待连接到 MC
FB_WaitConnect()
LogString("成功连接到 MC")
```

- example04.js
  本脚本演示了websocket功能   
  假设一台webscoket 服务器运行在地址 ws://localhost:8888/ws_test 上   
  我们现在要与其通信

```
// 当接收到新消息时，这个函数会被调用
function onNewMessage(newMessage) {
    FB_Println(newMessage)
}

// 连接到 ws://localhost:8888/ws_test 上
sendFn=FB_websocketConnectV1("ws://localhost:8888/ws_test",onNewMessage)

// 使用返回的发送函数向服务器发送消息
sendFn("hello ws!")
```

- example05.js  
本脚本演示了fetch功能

```
var x = fetch('https://storage.fastbuilder.pro').then(function(r) {
    r.text().then(function(d) {
        FB_Println(r.statusText)
        for (var k in r.headers._headers) {
            FB_Println(k + ':', r.headers.get(k))
        }
        FB_Println(d)
    });
});

FB_Println("Awaiting...")
```

## 其他
以下内容会被自动插入到用户脚本的开头
```
function FB_GeneralCmd(fbCmd){
    r=_FB_GeneralCmd(fbCmd)
    if(r instanceof Error){
        throw r
    }
    return r
}

function FB_SendMCCmd(mcCmd){
    r=_FB_SendMCCmd(mcCmd)
    if(r instanceof Error){
        throw r
    }
    return r
}

function FB_SendMCCmdAndGetResult(mcCmd){
    r=_FB_SendMCCmdAndGetResult(mcCmd)
    if(r instanceof Error){
        throw r
    }
    return JSON.parse(r)
}

function FB_RequireUserInput(hint){
    r=_FB_RequireUserInput(hint)
    if(r instanceof Error){
        throw r
    }
    return r
}

function FB_Println(msg){
    r=_FB_Println(msg)
    if(r instanceof Error){
        throw r
    }
    return r
}

function FB_RegPackCallBack(packetType,callBackFn){
    r=_FB_RegPackCallBack(packetType,function (jsonPacket) {
        // console.log(jsonPacket)
        callBackFn(JSON.parse(jsonPacket))
    })
    if (r instanceof Error){
        throw r
    }
    return r
}

// 订阅聊天信息
// 实际上只是对 golang 函数 _FB_RegPackCallBack 的重新利用
function FB_RegChat(callBackFn){
    r=_FB_RegPackCallBack("IDText",function (jsonPacket) {
        chatMsg=JSON.parse(jsonPacket)
        SourceName=chatMsg["SourceName"]
        Message=chatMsg["Message"]
        callBackFn(SourceName,Message)
    })
    if (r instanceof Error){
        throw r
    }
    return r
}

function FB_Query(info){
    r=_FB_Query(info)
    if (r instanceof Error){
        throw r
    }
    return r
}

function FB_SaveFile(fileName,data){
    if (_FB_SaveFile(fileName,data) instanceof Error){
        throw r
    }
}

function FB_ReadFile(fileName){
    r=_FB_ReadFile(fileName)
    if (r instanceof Error){
        throw r
    }
    return r
}

function FB_websocketConnectV1(serverAddress,onMessage) {
    r=_websocketConnectV1(serverAddress,function (newMessage) {
        if(newMessage instanceof Error){
            throw newMessage
        }
        onMessage(newMessage)
    })
    if(r instanceof Error){
        throw r
    }
    return r
}
```
PacketType 可用值:
```

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
// 本脚本演示了一个菜单，并会在用户登录时主动发送提示信息
// 演示了 game.subscribePacket，game.eval 的功能
// 核心是消息及数据包订阅

engine.setName("租赁服菜单")

// 打开自动重启功能，在连接断开时会自动重连
// engine.autoRestart()
// 请自己实现自动重启功能，由脚本控制host是不好的～

// 当有新玩家时，一定会收到 IDPlayerList 数据包，现在我们从这个数据包中判断玩家是谁
function onPlayerListUpdate(pk) {
    if (pk.ActionType !== 0) {
        // Action Type 为 0 时为玩家登录，否则为玩家退出
        return
    }
    pk.Entries.forEach(function (playerInfo) {
        // player Info 包括了相当多的信息，我们只需要其中的名字即可
        playerName = playerInfo.Username
        engine.message("新玩家:" + playerName)
        // 值得注意的是，玩家刚上线时并不能看到消息，所以我们延迟 8 秒（8000ms）再显示
        setTimeout(function () {
            game.oneShotCommand("tellraw @a {\"rawtext\":[{\"text\":\"欢迎回来！ @" + playerName + "\"}]}")
            game.oneShotCommand("tellraw " + playerName + " {\"rawtext\":[{\"text\":\"试试在聊天栏输入 '菜单' ! \"}]}")
        }, 8000)
    })
}

// 告诉 FB，当有这个数据包时就执行上面的函数

game.subscribePacket("IDPlayerList", onPlayerListUpdate)



// 实际上聊天功能基本就是一个问答机器人，接收聊天信息，并做出反应
game.listenChat(function (name, msg) {
    engine.message("Msg: " + name + ": " + msg)
    if (name === "") {
        // 不是人发出的聊天消息没有名字，比如命令块
        return
    }
    if (msg === "回城") {
        // 假设目的地是 0 100 0，这只是演示一下
        game.oneShotCommand("tp " + name + " 0 100 0")
    }
    if (msg === "冒险") {
        game.oneShotCommand("gamemode a " + name)
    }
    if (msg === "菜单") {
        game.oneShotCommand("tellraw " + name + " {\"rawtext\":[{\"text\":\"输入 回城 以回到 0 100 0  \"}]}")
        game.oneShotCommand("tellraw " + name + " {\"rawtext\":[{\"text\":\"输入 冒险 以切换为 冒险模式  \"}]}")
    }
})

// 当任何消息到来时，该函数都会被调用
function onAnyPacket(pk){
    engine.message(JSON.stringify(pk))
}

// 监听任何消息 (会有非常多的数据包)
// game.subscribePacket("*", onAnyPacket)

// 等待连接到 MC
engine.waitConnectionSync()

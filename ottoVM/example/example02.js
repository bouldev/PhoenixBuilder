// example02.js
// 本脚本演示了一个菜单，并会在用户登录时主动发送提示信息
// 演示了 FB_setTimeout，FB_RegPackCallBack，FB_GeneralCmd 的功能

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

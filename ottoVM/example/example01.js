// example01.js
// 本脚本演示了自动将机器人移动到玩家身边，并设置全局延迟为 100
// 演示了 FB_Query，FB_SendMCCmdAndGetResult，FB_GeneralCmd 的功能


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
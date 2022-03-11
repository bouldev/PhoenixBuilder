// 本脚本演示了自动将机器人移动到玩家身边，并设置全局延迟为 100
// 演示了 game.sendCommandSync，game.eval 的功能

engine.setName("FB信息和通用指令")

// 等待连接到 MC
engine.waitConnectionSync()

// 通用fb功能，相当于用户在fb中输入了这条指令
game.eval("delay set 100")

// FB相关信息保存在 consts 中
userName = consts.user_name

// 查看当前玩家有哪些，只是为了演示功能才那么做，其实没必要
listResult = game.sendCommandSync("list")
currentPlayers = listResult["OutputMessages"][1]["Parameters"] // "玩家1, 玩家2"

currentPlayersList = String(currentPlayers).split(", ")

engine.message("当前的玩家有:")
currentPlayersList.forEach(function (playerName) {
    engine.message(playerName)
    if (playerName === userName) {
        result = game.sendCommandSync("tp @s " + userName)
        engine.message("成功移动! " + JSON.stringify(result))
    }
})


// consts里所包含的信息
// 具体哪些信息可以被获得还在讨论中
// 所以这不是最终版本
// 脚本内容的哈希值
engine.message(consts.script_sha256)
// 脚本所在路径
engine.message(consts.script_path)
//用户名
engine.message(consts.user_name)
//用户FB token 的哈希值
engine.message(consts.sha_token)
//服务器号
engine.message(consts.server_code)
//FB 版本
engine.message(consts.fb_version)
//工作目录，一般为 FB 目录
engine.message(consts.fb_dir)
//js解释器相关各组件版本 引擎版本.桥接器版本.接口版本
engine.message(consts.engine_version)
// 本脚本演示了自动将机器人移动到玩家身边，并设置全局延迟为 100
// 演示了 consts，game.sendCommandSync.eval 的功能

engine.setName("FB信息和通用指令");

// 等待连接到 MC
// engine.waitConnectionSync();

// // 通用fb功能，相当于用户在fb中输入了这条指令
// game.eval("delay set 100");

// // 通过consts 查询信息
// let userName = consts.user_name
//
// // 查看当前玩家有哪些，只是为了演示功能才那么做，其实没必要
// let listResult = game.sendCommandSync("list")
// let currentPlayers = listResult["OutputMessages"][1]["Parameters"] // "玩家1, 玩家2"
//
// let currentPlayersList = String(currentPlayers).split(", ")
//
// engine.message("当前的玩家有:")
// currentPlayersList.forEach(function (playerName) {
// 	engine.message(playerName)
// 	if (playerName == userName) {
// 		let result = game.oneShotCommandAndGetResult("tp @s " + userName)
// 		engine.message("成功移动! " + JSON.stringify(result))
// 	}
// });

engine.message(JSON.stringify(consts));

// consts 能查询的所有信息
// 脚本内容的哈希值
engine.message(consts.script_sha256);
// 脚本所在路径
engine.message(consts.script_path);
// JS解释器实现
engine.message(consts.engine_version);
//用户名
// engine.message(consts.user_name);
//服务器代码
engine.message(consts.server_code);
//工作路径(一般情况下就是fb所在路径)
engine.message(consts.fb_dir);
//FB 版本信息
engine.message(consts.fb_version);
// uqHolder
let currentStatus = game.uqHolder()
// CommandRelatedEnums, InventoryContent 和 AvailableCommands 的信息过于丰富，屏幕显示不下，有兴趣的话自己研究
currentStatus["CommandRelatedEnums"] = {}
currentStatus["AvailableCommands"] = {}
currentStatus["InventoryContent"] = {}
engine.message(JSON.stringify(currentStatus));

players=currentStatus["PlayersByEntityID"]
for (const playerUID in players) {
    
    player=players[playerUID]

    // 在视野内的玩家有 Entity 属性,但是机器人本身反而没有这个属性
    if (player.Entity!==null){
        // 删除 Attributes 信息，因为内容太多了
        player.Entity.Attributes={}
    }
    
    engine.message(JSON.stringify(player));
}

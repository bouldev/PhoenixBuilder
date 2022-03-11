engine.waitConnectionSync()

// 获得机器人位置 pos.x pos.y pos.z
pos=game.botPos()
engine.message(JSON.stringify(pos))

// 移动
game.sendCommandSync("tp @s 100 200 300")
pos=game.botPos()
engine.message(JSON.stringify(pos))
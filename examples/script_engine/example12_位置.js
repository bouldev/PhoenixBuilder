engine.waitConnectionSync();

// 获得机器人位置 pos.x pos.y pos.z
let pos = game.botPos();
engine.message(JSON.stringify(pos));

// 移动
game.oneShotCommandAndGetResult("tp @s 100 200 300");
pos = game.botPos();
engine.message(JSON.stringify(pos));

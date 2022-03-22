// example03.js
// 本脚本演示了一个日志功能，主要用来展示文件读写
// 演示了 FB_setInterval，fs.readFile，fs.writeFile 的功能

engine.setName("日志");

// 向用户索要权限（需要玩家确认）
// 如果用户给了权限，第二次索要时不需要玩家确认，直接就能获得
let perm_success = fs.requestFilePermission("日志记录", "需要访问这个目录来保存数据");
if (!perm_success) {
	// 如果你的脚本必须要文件权限才能正常工作
	// 你可以使用该函数主动崩溃脚本
	engine.message("没有获得权限!");
	engine.crash("本脚本需要获得文件写入权限以运行。");
} else {
	engine.message("成功获得了权限!");
	engine.message("绝对路径为" + fs.getAbsPath("日志记录"));
}


// 一般情况下，应该使用 Append，但是考虑到跨平台，有的系统无法提供append，故只提供
// Save/Read 功能
// 警告！如果尝试访问未授权文件夹，脚本会被强制停止
// 即使获取了文件夹权限，fbtoken等敏感文件也是禁止访问的（脚本会被强制停止）

// 加载文件现有内容
let logData = fs.readFile("日志记录/日志.txt");

setInterval(()=>{
	// 每隔十秒保存一次
	engine.message("保存日志到" + fs.getAbsPath("日志记录/日志.txt"));
	fs.writeFile("日志记录/日志.txt", logData);
}, 10000);

// 添加一行记录
function LogString(info) {
	let curDate = new Date();
	logData = logData + curDate.toLocaleString() + ": " + info + "\n";
}

LogString("脚本启动");

// 记录聊天信息
game.listenChat((name, msg)=> {
	LogString("chat: " + name + " :" + msg);
});

// 等待连接到 MC
engine.waitConnectionSync();
LogString("成功连接到 MC");

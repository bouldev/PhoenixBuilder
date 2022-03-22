// 本脚演示了建议的指令发送方式

function onConnection() {
	// 发送一条指令，不需要它的结果
	engine.message("oneShotCommand");
	game.oneShotCommand("list");

	engine.message("sendCommnad");
	// 发送一条指令，需要它的结果（建议方式，异步）
	game.sendCommand("list", function (result) {
		engine.message("Async Result: " + JSON.stringify(result));

		// 在一个回调里再使用一个回调
		game.sendCommand("list", function (result) {
			engine.message("Async in Async Result: " + JSON.stringify(result));
		});
	})

	engine.message("sendCommnadSync");
	//  发送一条指令，需要它的结果（强烈不建议）
	// 如果服务器没有返回结果，脚本将被卡死!
	// 一些指令本身是有结果的，但是，在某些特殊情况下，
	// 不会收到结果，这也很危险

	engine.message("Sync");
	let result2 = game.sendCommandSync("list");
	engine.message("Sync Result: " + JSON.stringify(result2));
}

engine.waitConnection(onConnection);

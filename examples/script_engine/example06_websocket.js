// 本脚本演示了websocket功能，websocket是群服互通的关键
// 假设一台websocket 服务器运行在地址 ws://localhost:8888/ws_test 上
// 我们现在要与其通信

engine.setName("ws client")

// 当接收到新消息时，这个函数会被调用


// 连接到 ws://localhost:8888/ws_test 上
let websocket=new ws("ws://localhost:8888/ws_test");
websocket.onclose=()=> {
	engine.message("Websocket connection closed.");
}
websocket.onconnection=(connection)=> {
	// For ws client, the [connection] is the same object with [websocket].
	try {
		engine.message("Connected to server!");
		websocket.send("msg 1");
		websocket.send("msg 2");
		websocket.send("msg 3",1);
	}catch(e) {
		engine.message(`Error sending message to ws server: ${e.toString()}.`);
	}
	return;
};
function onNewMessage(message, type) {
	// msgType 为消息类型，一般为1，代表是文本消息（字符串）
	engine.message(`Received message: ${message} with type ${type}.`);
	if(message=="ping") {
		websocket.send("pong");
	}else if(message=="bye") {
		websocket.close();
	}
}
websocket.onmessage=onNewMessage;
websocket.onerror=(err)=> {
	engine.message(`Oops, error trapped: ${err}`);
}

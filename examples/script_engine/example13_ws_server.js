
engine.setName("ws server")

// 当收到新消息时，这个函数会被调用
function onMessage(msgType, msg, sendFn, closeFn) {
    engine.message("recv Msg: " + msgType + ": " + msg)
    sendFn(1, "server_echo: " + msg)
    engine.message("Server Send Successfully")
}

// 当有新连接时 这个函数会被调用
function onConnect(sendFn, closeFn) {
    engine.message("New Connection!")

    // 通过这个函数可以发送数据
    sendFn(1, "Hello Client!")
    return function (msgType, msg) { onMessage(msgType, msg, sendFn, closeFn) }
}

// 可以通过 ws://localhost:8888/ws_test 连接
// 即，与例6相同
ws.serve(":8888", "/ws_test", onConnect)

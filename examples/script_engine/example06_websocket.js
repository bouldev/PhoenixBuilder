// 本脚本演示了websocket功能，websocket是群服互通的关键
// 假设一台webscoket 服务器运行在地址 ws://localhost:8888/ws_test 上
// 我们现在要与其通信

engine.setName("ws client")

// 当接收到新消息时，这个函数会被调用
function onNewMessage(msgType, newMessage) {
    if (msgType === null && newMessage === null) {
        // 网络连接中发生错误时有发生,如果可以被这个错误无法被catch
        // 例如回调函数
        // 那么回调的所有值都会被设为 null
        engine.message("连接已经断开！")
        return
    }
    // msgType 为消息类型，一般为1，代表是文本消息（字符串）
    engine.message("type: " + msgType + " message: " + newMessage)
}

// 连接到 ws://localhost:8888/ws_test 上
try {
    sendFn = ws.connect("ws://localhost:8888/ws_test", onNewMessage)
} catch (e) {
    // 网络连接中发生错误时有发生,如果可以被这个错误可以被catch
    // 那么错误总是以 exception 发出的
    engine.message("捕捉了错误 " + e)
} finally {
    engine.message("继续执行")
}

// 使用返回的发送函数向服务器发送消息
// 1 为 msgType， 即消息类型，一般为1，代表是文本消息（字符串）
try {
    sendFn(1, "hello ws 1!")
    sendFn(1, "hello ws 2!")
    sendFn(1, "hello ws 3!")
} catch (e) {
    engine.message("捕捉了错误 " + e)
}

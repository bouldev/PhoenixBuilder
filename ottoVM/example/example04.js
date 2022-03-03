// example04.js
// 本脚本演示了websocket功能
// 假设一台webscoket 服务器运行在地址 ws://localhost:8888/ws_test 上
// 我们现在要与其通信


// 当接收到新消息时，这个函数会被调用
function onNewMessage(newMessage) {
    FB_Println(newMessage)
}

// 连接到 ws://localhost:8888/ws_test 上
sendFn=FB_websocketConnectV1("ws://localhost:8888/ws_test",onNewMessage)

// 使用返回的发送函数向服务器发送消息
sendFn("hello ws!")
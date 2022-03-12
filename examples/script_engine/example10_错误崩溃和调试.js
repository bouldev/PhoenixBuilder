// 你可以使用该函数主动崩溃脚本
engine.crash("在这里脚本崩溃了!")

// 当重复使用 script 指令加载同一个脚本时，前一个会被停止
// script example.js // 第一次加载
// 修改 example.js
// script example.js // 第二次加载时，第一次加载的脚本会被终止

//错误处理
//所有以FB_开头的函数，如果错误可以被 try...catch 捕获，则错误会被throw出去
//如果无法被try...catch 捕获（回调函数），则会在对应位置传入 null
//参考webscoket的例子，连接时，如果无法顺利连接，则会抛出错误，应该使用 try...catch语句
//如果无法收到新消息，由于无法在回调函数内捕捉回调函数外的错误，故传入null


// 当接收到新消息时，这个函数会被调用 （错误时收到null）
function onNewMessage(msgType, newMessage) {
    if (msgType === null && newMessage === null) {
        // 网络连接中发生错误时有发生,如果可以被这个错误无法被catch
        // 例如回调函数
        // 那么回调的所有值都会被设为 null
        engine.message("连接已经断开！")
        return
    }
    // 正常收到消息
    engine.message("type: " + msgType + " message: " + newMessage)
}

// 连接到 ws://localhost:8888/ws_test 上，(错误需要捕获)
try {
    ws.connect("ws://localhost:8888/ws_test", onNewMessage)
} catch (e) {
    // 网络连接中发生错误时有发生,如果可以被这个错误可以被catch
    // 那么错误总是以 exception 发出的
    engine.message("捕捉了错误 " + e)
} finally {
    engine.message("继续执行")
}

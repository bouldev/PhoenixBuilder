// FB提供的函数中,以下四个函数为同步函数
// engine.waitConnectionSync()
// engine.questionSync(hint)
// engine.sendCommandSync(mcCmd)
// engine.requestFilePermission(dir)
// 所谓同步函数，就是脚本会完全停止，直到获得结果

// 其中，三个函数有异步版本，所谓异步，即脚本不会停止
// 当获得结果时，函数会被回调


const afterGettedUserInput = function (userInput) {
    engine.message("成功获得了用户输入！" + userInput)
}

const afterGettedCmdResult = function (result) {
    engine.message("成功获得了指令结果！" + result)

    // 当获得用户输入后，afterGettedUserInput会被回调
    engine.requestFilePermission("这里是为什么要访问", afterGettedUserInput)
}

const afterConnected = function () {
    engine.message("成功连接到MC了！")

    // 当获得指令结果后，afterGettedCmdResult会被回调
    engine.sendCommand("list", afterGettedCmdResult)
}

// 当连接到MC后，afterConnected会被回调
engine.waitConnection(afterConnected)

engine.message("和engine.waitConnectionSync不同，即使没有连接到FB，我也会执行")

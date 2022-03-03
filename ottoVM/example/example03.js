// example03.js
// 本脚本演示了一个日志功能，主要用来展示文件读写
// 演示了 FB_setInterval，FB_ReadFile，FB_SaveFile 的功能

// 一般情况下，应该使用 Append，但是考虑到跨平台，有的系统无法提供append，故只提供
// Save/Read 功能
logData=FB_ReadFile("日志.txt")

FB_setInterval(function () {
    // 每隔十秒保存一次
    console.log("Save Log")
    FB_SaveFile("日志.txt",logData)
},10000)

// 添加一行记录
function LogString(info) {
    newDate = new Date();
    logData=logData+newDate.toLocaleString()+": "+info+"\n"
}

LogString("脚本启动")

// 记录聊天信息
FB_RegChat(function (name,msg) {
    LogString("chat: "+name+" :"+msg)
})

// 等待连接到 MC
FB_WaitConnect()
LogString("成功连接到 MC")
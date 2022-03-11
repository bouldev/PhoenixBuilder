// 本脚演示了时间记分板的校正
// 演示了 engine.setName engine.waitConnectionSync，engine.questionSync，engine.message，game.sendCommand 的功能
// 假设用户有一个记分板，记分板里有 year, month, day, hour, minute 四个项目
// 需要与现实时间同步

// 这个不是必须的，不设置时会以脚本文件名作为名字
engine.setName("时间校准")

// 等待连接到 MC
engine.waitConnectionSync()
engine.message("已经连接到服务器!")

// 请求用户输入信息 (时间相关记分板的名字)
scoreBoardName = engine.questionSync("时间记分板的名字是?")

// js: 计算时间
let nowTime = new Date()
let nowYear = nowTime.getFullYear()
let nowMonth = nowTime.getMonth()
let nowDay = nowTime.getDay()
let nowHour = nowTime.getHours()
let nowMinute = nowTime.getMinutes()

// 发送指令
// sendCommand不指定第二个参数时等同于oneShotCommand
game.sendCommand("scoreboard objectives add " + scoreBoardName + " dummy 时间记分板")
game.oneShotCommand("scoreboard players set year " + scoreBoardName + " " + nowYear)
game.oneShotCommand("scoreboard players set month " + scoreBoardName + " " + nowMonth)
game.oneShotCommand("scoreboard players set day " + scoreBoardName + " " + nowDay)
game.oneShotCommand("scoreboard players set hour " + scoreBoardName + " " + nowHour)
game.oneShotCommand("scoreboard players set minute " + scoreBoardName + " " + nowMinute)

// 向用户发送提示信息
engine.message("时间记分板校准完成！")

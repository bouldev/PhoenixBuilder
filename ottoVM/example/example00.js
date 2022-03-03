// example00.js
// 本脚演示了时间记分板的校正
// 演示了 FB_WaitConnect，FB_RequireUserInput，FB_Println，FB_SendMCCmd 的功能
// 假设用户有一个记分板，记分板里有 year, month, day, hour, minute 四个项目
// 需要与现实时间同步

// 等待连接到 MC
FB_WaitConnect()

// 请求用户输入信息 (时间相关记分板的名字)
scoreBoardName=FB_RequireUserInput("时间记分板的名字是?")

// js: 计算时间
nowTime=new Date()
nowYear=nowTime.getFullYear()
nowMonth=nowTime.getMonth()
nowDay=nowTime.getDay()
nowHour=nowTime.getHours()
nowMinute=nowTime.getMinutes()

// 发送指令
FB_SendMCCmd("scoreboard players set "+scoreBoardName+" year "+nowYear)
FB_SendMCCmd("scoreboard players set "+scoreBoardName+" month "+nowMonth)
FB_SendMCCmd("scoreboard players set "+scoreBoardName+" day "+nowDay)
FB_SendMCCmd("scoreboard players set "+scoreBoardName+" hour "+nowHour)
FB_SendMCCmd("scoreboard players set "+scoreBoardName+" minute "+nowMinute)

// 向用户发送提示信息
FB_Println("时间记分板校准完成！")
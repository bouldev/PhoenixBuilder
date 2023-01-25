# 插件: 关
# 需要使用的请把这个"关"改为"开"


from omega_side.python3_omega_sync import API
from omega_side.python3_omega_sync import frame as omega
from omega_side.python3_omega_sync.bootstrap import install_lib
from omega_side.python3_omega_sync.protocol import *
import time
#导入库

class version_公告系统:
    author  = "Mono"
    version = "2.0 for omegaside"

#-----设置------
#设置计分板名称
scoreboardName_ad = "公告"
#设置QQ群号
qqgroupCode       = "群号"



def Mono_plugin_ad(api:API):
    
    def Mono_ad_system(num:int = 1):
        if num >0 :
            result         = api.do_get_players_list(cb=None)
            playernumbers  = len(result)
            Timedate       = time.localtime(time.time())
            Noweek         = time.asctime(Timedate)
            TimeList       = list(Timedate)
            TimeDataTitle  ="未知"
            TimeTitlecolor ="d"
            if TimeList[3] >= 0:
                if TimeList[3] >= 0 and TimeList[3] <6:
                    TimeTitlecolor = "9"
                    TimeDataTitle  = "清晨"
                if TimeList[3] >=6 and TimeList[3] < 11:
                    TimeTitlecolor = "a"#绿
                    TimeDataTitle  = "早晨"
                if TimeList[3] >=11 and TimeList[3] < 13:
                    TimeTitlecolor = "c"#红
                    TimeDataTitle  = "午时"
                if TimeList[3] >=13 and TimeList[3] < 17:
                    TimeTitlecolor = "g"#淡黄
                    TimeDataTitle  = "下午"
                if TimeList[3] >=17  and TimeList[3] < 22:
                    TimeTitlecolor = "b"#蓝
                    TimeDataTitle  = "晚上"
                if TimeList[3] >=22 :
                    TimeTitlecolor = "3"#青
                    TimeDataTitle  = "深夜"
            #以实际情况打开下面内容
            # try:
            #     result_admin=api.do_send_ws_cmd("/testfor @a[tag=op,tag=!omg]",cb=None)["result"]["OutputMessages"][0]["Parameters"][0].split(", ")
            # except IndexError:
            #     result_admin=[]
            if num == 1:
                style_1={
                    "1":["§r●§7%s/%s/%s %s"%(TimeList[0],TimeList[1],TimeList[2],Noweek[0:3]),9],
                    "2":["§r●§7%s §%s%s:%s"%(TimeDataTitle,TimeTitlecolor,TimeList[3],TimeList[4]),8],
                    #"3":["§r●§a在线管理: §r§l%s"%len(result_admin),7],
                    "3":["§r●§7TPS: §a20.0",7],#这个tps omgside没有提供参数,视情况使用上面的那个
                    "4":["§r●§7在线玩家:§e%s"%(playernumbers) , 6],
                    "5":["§r●§7输入§f§lomg§r§7打开菜单",5],
                    "6":["§r●§7交流号:%s"%(qqgroupCode),4]
                }
                api.do_send_ws_cmd('/scoreboard players reset * %s'%scoreboardName_ad,cb=None)
                api.do_send_ws_cmd(f"""/scoreboard players set "{style_1["1"][0]}" {scoreboardName_ad} {style_1["1"][1]}""",cb=None)
                api.do_send_ws_cmd(f"""/scoreboard players set "{style_1["2"][0]}" {scoreboardName_ad} {style_1["2"][1]}""",cb=None)
                api.do_send_ws_cmd(f"""/scoreboard players set "{style_1["3"][0]}" {scoreboardName_ad} {style_1["3"][1]}""",cb=None)
                api.do_send_ws_cmd(f"""/scoreboard players set "{style_1["4"][0]}" {scoreboardName_ad} {style_1["4"][1]}""",cb=None)
                api.do_send_ws_cmd(f"""/scoreboard players set "{style_1["5"][0]}" {scoreboardName_ad} {style_1["5"][1]}""",cb=None)
                api.do_send_ws_cmd(f"""/scoreboard players set "{style_1["6"][0]}" {scoreboardName_ad} {style_1["6"][1]}""",cb=None)
                del style_1,TimeDataTitle,TimeTitlecolor,result,playernumbers,Timedate,Noweek,TimeList
            elif num == 2:
                ...#更多样式见我做的dotcs版的公告系统
    api.execute_with_repeat(Mono_ad_system,repeat_time=30)#30秒刷新一次,omgside这个有延迟因此如果时间太少会频闪的
omega.add_plugin(plugin=Mono_plugin_ad)

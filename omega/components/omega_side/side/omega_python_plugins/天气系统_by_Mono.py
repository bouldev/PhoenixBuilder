# 插件: 开


#导入库
from omega_side.python3_omega_sync import API
from omega_side.python3_omega_sync import frame as omega
from omega_side.python3_omega_sync.bootstrap import install_lib
from omega_side.python3_omega_sync.protocol import *
#安装库
install_lib(lib_name="numpy",lib_install_name="numpy")
import numpy,time,os,json,random


#版本/作者
class version_weather:
    time     = "2023.1.23"            #创建时间
    author   = "Mono"                 #作者
    version  = "2.0 for omegaside"    #版本

#初始化
def initialize():
    if not os.path.exists(os.path.join('data','weather.json')):
        nowtime=str(time.strftime("%Y %m %d %H %M %S")).split(" ")
        nowtime_="".join(nowtime)
        data={"名称":"天气系统","描述":"天气系统的存储文件","信息":{"初始时间":"%s"%(nowtime_),"季节":"春","已过天数":0,"酸雨天":[5,20],"天气":"晴","low":0,"high":5,"温度":2}}
        with open(os.path.join('data','weather.json'),"w",encoding="utf-8") as f:
            f.write(json.dumps(data,ensure_ascii=False,indent=4))
initialize()
def Mono_plugin_season(api:API):
    #主要功能函数
    class weather:
        def acid_rain():
            """
            酸雨天
            ---

            """
            if 1:
                with open(os.path.join('data','weather.json'),"r",encoding="utf-8") as f:
                    data=f.read()
                data=json.loads(data)

                if data["信息"]["已过天数"] in data["信息"]["酸雨天"]:
                    #检测在线玩家
                    response=api.do_get_players_list(cb=None)
                    allplayers=[]
                    for i in response:
                        allplayers.append(i.name)
                    #检测玩家是否带头盔
                    if 1:
                        result_1=api.do_send_ws_cmd("/testfor @a[hasitem={item=leather_helmet,location=slot.armor.head,slot=0}]",cb=None)["result"]["OutputMessages"][0]["Parameters"]
                        if len(result_1) == 0:
                            pass
                        elif ", " in result_1[0]:
                            result_1[0]=result_1[0].split(", ")
                            for i in result_1[0]:
                                allplayers.remove(i)
                        else:
                            allplayers.remove(result_1[0])
                    
                    if 1:
                        result_1=api.do_send_ws_cmd("/testfor @a[hasitem={item=chainmail_helmet,location=slot.armor.head,slot=0}]",cb=None)["result"]["OutputMessages"][0]["Parameters"]
                        if len(result_1) == 0:
                            pass
                        elif ", " in result_1[0]:
                            result_1[0]=result_1[0].split(", ")
                            for i in result_1[0]:
                                allplayers.remove(i)
                        else:
                            allplayers.remove(result_1[0])
                    if 1:
                        result_1=api.do_send_ws_cmd("/testfor @a[hasitem={item=iron_helmet,location=slot.armor.head,slot=0}]",cb=None)["result"]["OutputMessages"][0]["Parameters"]
                        if len(result_1) == 0:
                            pass
                        elif ", " in result_1[0]:
                            result_1[0]=result_1[0].split(", ")
                            for i in result_1[0]:
                                allplayers.remove(i)
                        else:
                            allplayers.remove(result_1[0])
                    
                    if 1:
                        result_1=api.do_send_ws_cmd("/testfor @a[hasitem={item=golden_helmet,location=slot.armor.head,slot=0}]",cb=None)["result"]["OutputMessages"][0]["Parameters"]
                        if len(result_1) == 0:
                            pass
                        elif ", " in result_1[0]:
                            result_1[0]=result_1[0].split(", ")
                            for i in result_1[0]:
                                allplayers.remove(i)
                        else:
                            allplayers.remove(result_1[0])
                    
                    if 1:
                        result_1=api.do_send_ws_cmd("/testfor @a[hasitem={item=diamond_helmet,location=slot.armor.head,slot=0}]",cb=None)["result"]["OutputMessages"][0]["Parameters"]
                        if len(result_1) == 0:
                            pass
                        elif ", " in result_1[0]:
                            result_1[0]=result_1[0].split(", ")
                            for i in result_1[0]:
                                allplayers.remove(i)
                        else:
                            allplayers.remove(result_1[0])
                    
                    if 1:
                        result_1=api.do_send_ws_cmd("/testfor @a[hasitem={item=netherite_helmet,location=slot.armor.head,slot=0}]",cb=None)["result"]["OutputMessages"][0]["Parameters"]
                        if len(result_1) == 0:
                            pass
                        elif ", " in result_1[0]:
                            result_1[0]=result_1[0].split(", ")
                            for i in result_1[0]:
                                allplayers.remove(i)
                        else:
                            allplayers.remove(result_1[0])
                    #检测头上无方块的玩家并给予效果&扣除盔甲耐久
                    for i in allplayers:    
                        api.do_send_ws_cmd("/tp @s 0 100 0",cb=None)
                        api.do_send_ws_cmd("/fill 0 256 0 0 0 0 air",cb=None)
                        reslut=api.do_send_ws_cmd(f"/execute @a[name={i}] ~~~ testforblocks 0 ~1 0 0 256 0 ~~1~",cb=None)
                        api.do_send_ws_cmd(f"/damage @a[name={i}] 0 projectile",cb=None)
                        if reslut.result.SuccessCount ==1:
                            api.do_send_ws_cmd(f"/effect @a[name={i}] poison 4 0 true",cb=None)
            
        

        def effect_season():
            #季节效果
            while True:
                time.sleep(2)
                with open(os.path.join('data','weather.json'),"r",encoding="utf-8") as f:
                    data=f.read()
                try:
                    data=json.loads(data)
                
                    temp = data["信息"]["温度"]
                    season = data["信息"]["季节"]
                    #检测温度是否在合适值
                    if temp >=-30:
                        if temp >=45:
                            api.do_send_ws_cmd("/effect @a nausea 4 2",cb=None)
                        elif temp <=-10:
                            api.do_send_ws_cmd("/effect @a mining_fatigue 2 0",cb=None)
                    if season != "":
                        if season == "春":
                            api.do_send_ws_cmd("/effect @a regeneration 5 0 true",cb=None)
                        elif season == "夏":
                            api.do_send_ws_cmd("/effect @a haste 5 0 true",cb=None)
                        elif season == "秋":
                            api.do_send_ws_cmd("/effect @a resistance 5 0 true",cb=None)
                        elif season == "冬":
                            api.do_send_ws_cmd("/effect @a fire_resistance 5 0 true",cb=None)
                except:
                    pass
        def temp():
            if 1:
                #温度浮动
                with open(os.path.join('data','weather.json'),"r",encoding="utf-8") as f:
                    data=f.read()
                data=json.loads(data)
                if -30<= data["信息"]["温度"] <= -10:
                    data["信息"]["温度"] = random.randint(data["信息"]["low"],data["信息"]["high"])
                if -10< data["信息"]["温度"] <= 10:
                    data["信息"]["温度"] = random.randint(data["信息"]["low"],data["信息"]["high"])
                if 10< data["信息"]["温度"] <= 20:
                    data["信息"]["温度"] = random.randint(data["信息"]["low"],data["信息"]["high"])
                if 20< data["信息"]["温度"] <= 30:
                    data["信息"]["温度"] = random.randint(data["信息"]["low"],data["信息"]["high"])
                if 30< data["信息"]["温度"] <= 40:
                    data["信息"]["温度"] = random.randint(data["信息"]["low"],data["信息"]["high"])
                if 40< data["信息"]["温度"] <= 50:
                    data["信息"]["温度"] = random.randint(data["信息"]["low"],data["信息"]["high"])
                with open(os.path.join('data','weather.json'),"w",encoding="utf-8") as f:
                    f.write(json.dumps(data,ensure_ascii=False,indent=4))
        def seasonsRepeat():
            """
            四季循环
            ---
            春温度0--20
            夏温度15---50
            秋10--30
            冬-30---10
            每季切换时间30天(现实)
            
            """
            try:
                with open(os.path.join('data','weather.json'),"r",encoding="utf-8") as f:
                    data=f.read()
                try :
                    data=json.loads(data)
                
                except :
                    pass
                today_weather=data["信息"]["天气"]
                #判断季节
                if data["信息"]["季节"] == "春":
                    result_date=json.loads(api.do_send_ws_cmd("/time add 0",cb=None)["result"]["DataSet"])["time"]
                    data["信息"]["游戏时间"] = result_date
                    #获取时间
                    if result_date < 121 and result_date > 1:
                        sampleList = ["晴","多云","阴","雨","暴雨"]
                        data["信息"]["天气"] = numpy.random.choice(sampleList, 1, p=[0.33, 0.15, 0.09, 0.35, 0.08])[0]
                        today_weather=data["信息"]["天气"]
                        low_ = random.randint(0,12)
                        high_ =random.randint(13,20)
                        temp = random.randint(low_,high_)
                        api.do_send_ws_cmd("/playsound note.harp @a",cb=None)
                        if data["信息"]["已过天数"] < 30 :
                            data["信息"]["已过天数"] = data["信息"]["已过天数"]+1
                        elif data["信息"]["已过天数"] == 30:
                            data["信息"]["已过天数"] = 1
                            data["信息"]["季节"] = "夏"
                            data["信息"]["酸雨天"] = [random.randint(0,30),random.randint(0,30)]
                        if data["信息"]["已过天数"] in data["信息"]["酸雨天"]:
                            today_weather = "酸雨"
                            data["信息"]["天气"]="酸雨"
                        api.do_send_player_msg("@a",f"""§a{"-"*20}\n§r●  今日天气 : §e{today_weather}§r\n●  最高温度 :§e {high_}§r\n●  最低温度 : §e{low_}§r\n●  当前温度 :  §b{temp}§r\n§a{"-"*20}""",cb=None)
                        api.do_set_player_actionbar("@a",f'§a{data["信息"]["季节"]}季§r,第§b{data["信息"]["已过天数"]}§r天',cb=None)
                        api.do_set_player_title("@a",f"""§a{data["信息"]["季节"]}""",cb=None)
                        api.do_send_ws_cmd(f"""/title @a subtitle §f第§b{data["信息"]["已过天数"]}§f天""",cb=None)
                        api.do_send_ws_cmd("/time set 150",cb=None)
                        data["信息"]["温度"] = temp
                        data["信息"]["high"] = high_
                        data["信息"]["low"] = low_
                    with open(os.path.join('data','weather.json'),"w",encoding="utf-8") as f:
                        f.write(json.dumps(data,ensure_ascii=False,indent=4))
                    if today_weather !="":
                        if data["信息"]["已过天数"] in data["信息"]["酸雨天"]:
                            api.do_send_ws_cmd("/weather rain",cb=None)
                        elif "晴" == today_weather :
                            api.do_send_ws_cmd("/weather clear",cb=None)
                        elif "多云" == today_weather :
                            api.do_send_ws_cmd("/weather clear",cb=None)
                        elif "阴" == today_weather :
                            api.do_send_ws_cmd("/weather clear",cb=None)
                        elif "雨" == today_weather :
                            api.do_send_ws_cmd("/weather rain",cb=None)
                        elif "暴雨" == today_weather :
                            api.do_send_ws_cmd("/weather thunder",cb=None)
                        
                elif data["信息"]["季节"] == "夏":
                    result_date=json.loads(api.do_send_ws_cmd("/time add 0",cb=None)["result"]["DataSet"])["time"]
                    data["信息"]["游戏时间"] = result_date
                    if result_date < 121 and result_date > 1:
                        sampleList = ["晴","多云","阴","雨","暴雨"]
                        data["信息"]["天气"]  = numpy.random.choice(sampleList, 1, p=[0.45, 0.13, 0.09, 0.23, 0.1])[0]
                        today_weather=data["信息"]["天气"]
                        low_ = random.randint(15,34)
                        high_ =random.randint(35,50)
                        temp = random.randint(low_,high_)
                        api.do_send_ws_cmd("/playsound note.harp @a",cb=None)
                        if data["信息"]["已过天数"] < 30 :
                            data["信息"]["已过天数"] = data["信息"]["已过天数"]+1
                        elif data["信息"]["已过天数"] == 30:
                            data["信息"]["已过天数"] = 1
                            data["信息"]["季节"] = "秋"
                            data["信息"]["酸雨天"] = [random.randint(0,30),random.randint(0,30)]
                        if data["信息"]["已过天数"] in data["信息"]["酸雨天"]:
                            today_weather = "酸雨"
                            data["信息"]["天气"]="酸雨"
                        api.do_send_player_msg("@a",f"""§a{"-"*20}\n§r●  今日天气 : §e{today_weather}§r\n●  最高温度 :§e {high_}§r\n●  最低温度 : §e{low_}§r\n●  当前温度 :  §b{temp}§r\n§a{"-"*20}""",cb=None)
                        api.do_set_player_actionbar("@a",f'§a{data["信息"]["季节"]}季§r,第§b{data["信息"]["已过天数"]}§r天',cb=None)
                        api.do_set_player_title("@a",f"""§a{data["信息"]["季节"]}""",cb=None)
                        api.do_send_ws_cmd(f"""/title @a subtitle §f第§b{data["信息"]["已过天数"]}§f天""",cb=None)
                        api.do_send_ws_cmd("/time set 150",cb=None)
                        data["信息"]["温度"] = temp
                        data["信息"]["high"] = high_
                        data["信息"]["low"] = low_
                    with open(os.path.join('data','weather.json'),"w",encoding="utf-8") as f:
                        f.write(json.dumps(data,ensure_ascii=False,indent=4))
                    if today_weather !="":
                        if data["信息"]["已过天数"] in data["信息"]["酸雨天"]:
                            api.do_send_ws_cmd("/weather rain",cb=None)
                        elif "晴" == today_weather :
                            api.do_send_ws_cmd("/weather clear",cb=None)
                        elif "多云" == today_weather :
                            api.do_send_ws_cmd("/weather clear",cb=None)
                        elif "阴" == today_weather :
                            api.do_send_ws_cmd("/weather clear",cb=None)
                        elif "雨" == today_weather :
                            api.do_send_ws_cmd("/weather rain",cb=None)
                        elif "暴雨" == today_weather :
                            api.do_send_ws_cmd("/weather thunder",cb=None)
                elif data["信息"]["季节"] == "秋":
                    result_date=json.loads(api.do_send_ws_cmd("/time add 0",cb=None)["result"]["DataSet"])["time"]
                    data["信息"]["游戏时间"] = result_date
                    if result_date < 121 and result_date > 1:
                        sampleList = ["晴","多云","阴","雨","暴雨"]
                        data["信息"]["天气"]  = numpy.random.choice(sampleList, 1, p=[0.25, 0.25, 0.19, 0.23, 0.08])[0]
                        today_weather=data["信息"]["天气"]
                        low_ = random.randint(10,19)
                        high_ =random.randint(20,30)
                        temp = random.randint(low_,high_)
                        api.do_send_ws_cmd("/playsound note.harp @a",cb=None)
                        if data["信息"]["已过天数"] < 30 :
                            data["信息"]["已过天数"] = data["信息"]["已过天数"]+1
                        elif data["信息"]["已过天数"] == 30:
                            data["信息"]["已过天数"] = 1
                            data["信息"]["季节"] = "冬"
                            data["信息"]["酸雨天"] = [random.randint(0,30),random.randint(0,30)]
                        if data["信息"]["已过天数"] in data["信息"]["酸雨天"]:
                            today_weather = "酸雨"
                            data["信息"]["天气"]="酸雨"
                        api.do_send_player_msg("@a",f"""§a{"-"*20}\n§r●  今日天气 : §e{today_weather}§r\n●  最高温度 :§e {high_}§r\n●  最低温度 : §e{low_}§r\n●  当前温度 :  §b{temp}§r\n§a{"-"*20}""",cb=None)
                        api.do_set_player_actionbar("@a",f'§a{data["信息"]["季节"]}季§r,第§b{data["信息"]["已过天数"]}§r天',cb=None)
                        api.do_set_player_title("@a",f"""§a{data["信息"]["季节"]}""",cb=None)
                        api.do_send_ws_cmd(f"""/title @a subtitle §f第§b{data["信息"]["已过天数"]}§f天""",cb=None)
                        api.do_send_ws_cmd("/time set 150",cb=None)
                        data["信息"]["温度"] = temp
                        data["信息"]["high"] = high_
                        data["信息"]["low"] = low_
                    with open(os.path.join('data','weather.json'),"w",encoding="utf-8") as f:
                        f.write(json.dumps(data,ensure_ascii=False,indent=4))
                    if today_weather !="":
                        if data["信息"]["已过天数"] in data["信息"]["酸雨天"]:
                            api.do_send_ws_cmd("/weather rain",cb=None)
                        elif "晴" == today_weather :
                            api.do_send_ws_cmd("/weather clear",cb=None)
                        elif "多云" == today_weather :
                            api.do_send_ws_cmd("/weather clear",cb=None)
                        elif "阴" == today_weather :
                            api.do_send_ws_cmd("/weather clear",cb=None)
                        elif "雨" == today_weather :
                            api.do_send_ws_cmd("/weather rain",cb=None)
                        elif "暴雨" == today_weather :
                            api.do_send_ws_cmd("/weather thunder",cb=None)
                elif data["信息"]["季节"] == "冬":
                    result_date=json.loads(api.do_send_ws_cmd("/time add 0",cb=None)["result"]["DataSet"])["time"]
                    data["信息"]["游戏时间"] = result_date
                    if result_date < 121 and result_date > 1:
                        sampleList = ["晴","多云","阴","雨","暴雨"]
                        data["信息"]["天气"]  = numpy.random.choice(sampleList, 1, p=[0.25, 0.15, 0.19, 0.23, 0.18])[0]
                        today_weather=data["信息"]["天气"]
                        low_ = random.randint(-30,-7)
                        high_ =random.randint(-6,10)
                        temp = random.randint(low_,high_)
                        api.do_send_ws_cmd("/playsound note.harp @a",cb=None)
                        if data["信息"]["已过天数"] < 30 :
                            data["信息"]["已过天数"] = data["信息"]["已过天数"]+1
                        elif data["信息"]["已过天数"] == 30:
                            data["信息"]["已过天数"] = 1
                            data["信息"]["季节"] = "春"
                            data["信息"]["酸雨天"] = [random.randint(0,30),random.randint(0,30)]
                        if data["信息"]["已过天数"] in data["信息"]["酸雨天"]:
                            today_weather = "酸雨"
                            data["信息"]["天气"]="酸雨"
                        api.do_send_player_msg("@a",f"""§a{"-"*20}\n§r●  今日天气 : §e{today_weather}§r\n●  最高温度 :§e {high_}§r\n●  最低温度 : §e{low_}§r\n●  当前温度 :  §b{temp}§r\n§a{"-"*20}""",cb=None)
                        api.do_set_player_actionbar("@a",f'§a{data["信息"]["季节"]}季§r,第§b{data["信息"]["已过天数"]}§r天',cb=None)
                        api.do_set_player_title("@a",f"""§a{data["信息"]["季节"]}""",cb=None)
                        api.do_send_ws_cmd(f"""/title @a subtitle §f第§b{data["信息"]["已过天数"]}§f天""",cb=None)
                        api.do_send_ws_cmd("/time set 150",cb=None)
                        data["信息"]["温度"] = temp
                        data["信息"]["high"] = high_
                        data["信息"]["low"] = low_
                    with open(os.path.join('data','weather.json'),"w",encoding="utf-8") as f:
                        f.write(json.dumps(data,ensure_ascii=False,indent=4))
                    if today_weather !="":
                        if data["信息"]["已过天数"] in data["信息"]["酸雨天"]:
                            api.do_send_ws_cmd("/weather rain",cb=None)
                        elif "晴" == today_weather :
                            api.do_send_ws_cmd("/weather clear",cb=None)
                        elif "多云" == today_weather :
                            api.do_send_ws_cmd("/weather clear",cb=None)
                        elif "阴" == today_weather :
                            api.do_send_ws_cmd("/weather clear",cb=None)
                        elif "雨" == today_weather :
                            api.do_send_ws_cmd("/weather rain",cb=None)
                        elif "暴雨" == today_weather :
                            api.do_send_ws_cmd("/weather thunder",cb=None)
            except:
                pass
    api.execute_with_repeat(weather.acid_rain,repeat_time=2)
    api.execute_with_repeat(weather.effect_season,repeat_time=3)
    api.execute_with_repeat(weather.seasonsRepeat,repeat_time=3)
    api.execute_with_repeat(weather.temp,repeat_time=50)
    #聊天栏获取当前温度
    def get_temp(player_input:PlayerInput):
        with open(os.path.join('data','weather.json'),"r",encoding="utf-8") as f:
            data=f.read()
        data=json.loads(data)
        temp=data["信息"]["温度"]
        today_weather=data["信息"]["天气"]
        high=data["信息"]["high"]
        low=data["信息"]["low"]
        if data["信息"]["已过天数"] in data["信息"]["酸雨天"]:
            today_weather="酸雨"
        api.do_send_player_msg(player_input.Name,f"""§a{"-"*20}\n§r●  今日天气 : §e{today_weather}§r\n●  最高温度 :§e {high}§r\n●  最低温度 : §e{low}§r\n●  当前温度 :  §b{temp}§r\n§a{"-"*20}""".ljust(50),cb=None)
    #将功能插入到omg菜单中 
    api.listen_omega_menu(triggers=["温度","temp"],argument_hint="",usage="获取当前温度 [温度/temp]",cb=None,on_menu_invoked=get_temp)

omega.add_plugin(plugin=Mono_plugin_season)

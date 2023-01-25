# 插件: 关
# 需要使用的请把这个"关"改为"开"
from omega_side.python3_omega_sync import API
from omega_side.python3_omega_sync import frame as omega
from omega_side.python3_omega_sync.bootstrap import install_lib
from omega_side.python3_omega_sync.protocol import *
import uuid,ast,json,random,time,os




class version_redbag:
    author     =    "Mono"
    version    =    "2.0 for omegaside"

#生成文件
os.path.join('data','')
if not os.path.exists(os.path.join('data','redbag_Data.json')):
    with open(os.path.join('data','redbag_Data.json'),"w",encoding="utf-8") as f:
        f.write(json.dumps({"名称":"红包系统数据存储","描述":"储存谁发了红包和发了多少红包(这不是配置文件哦)","信息":{}},ensure_ascii=False,indent=4))
if not os.path.isdir(os.path.join('data','player')):
    os.mkdir(os.path.join('data','player'))
if not os.path.isdir(os.path.join('data','红包系统')):
    os.mkdir(os.path.join('data','红包系统'))



if not os.path.exists(os.path.join('data','红包系统','组件_红包系统.json')):
    with open(os.path.join('data','红包系统','组件_红包系统.json'),"w",encoding="utf-8") as f:
        f.write(json.dumps({"名称":"红包系统","描述":"红包系统配置文件,跨服红包暂不支持","配置":{"计分板名称":"存储","默认祝福语":"恭喜发财","玩家领取红包后执行命令":["/time add 0"],"禁用跨服红包":True,"禁用跨服红包发来的指令":True}},ensure_ascii=False,indent=4))
with open(os.path.join('data','红包系统','组件_红包系统.json'),"r",encoding="utf-8") as f:
    data=f.read()
    data=json.loads(data)
    scoreboardName= data["配置"]["计分板名称"]
    defaultBlessingWords=data["配置"]["默认祝福语"] 
#这是复制的dotcs社区版的一个函数
def getPlayerData(dataName: str, playerName: str, writeNew: str = "") -> (str | int | float):
        """
        获取玩家本地数据的函数
        读取文件: data\player\playerName\dataName.txt

        参数:
            dataName: str -> 数据名称
            playerName: str -> 玩家名称
            writeNew: str -> 若数据不存在, 写入的数据
        返回: str | int | float -> 文件读取结果
        """
        dataName = dataName.replace("\\", "/")
        fileDir = "data/player/%s/%s.txt" % (playerName, dataName)
        pathDir = ""
        pathAll = ("data/player/%s/%s" % (playerName, dataName)).split("/")
        pathAll.pop(-1)
        for i in pathAll:
            pathDir += "%s/" % i
        if not os.path.isdir(pathDir):
            pathToCreate = ""
            for i in pathDir.split("/"):
                try:
                    pathToCreate += "%s/" % i
                    os.mkdir(pathToCreate)
                except:
                    pass
        if not os.path.isfile(fileDir):
            with open(fileDir, "w", encoding = "utf-8", errors = "ignore") as file:
                file.write(writeNew)
        with open(fileDir, "r", encoding = "utf-8", errors = "ignore") as file:
            data = file.read()
        if "." not in data:
            try:
                data = int(data)
            except:
                pass
        else:
            try:
                data = float(data)
            except:
                pass
        return data

class redbag:
    def algorithm(total_amount:int,total_people:int):
        """
        红包金额计算
        total_amount -> 红包金额
        total_people: -> 红包数量

        返回dict
        """
        if total_amount< 0 or total_amount==0:
            return {"count":0,"error":"§a小世界§e>> §c错误:红包金额小于或者等于0","value":False}
        elif total_people<0 or total_people ==0:
            return {"count":0,"error":"§a小世界§e>> §c错误:红包数量小于或者等于0","value":False}
        elif total_amount < total_people:
            return {"count":0,"error":"§a小世界§e>> §c错误:红包金额不能小于红包数量","value":False}
        amount_list=[]
        rest_aomunt=total_amount
        rest_people=total_people
        for i in range(0,total_people-1):
            amount=random.randint(1,int(rest_aomunt/rest_people * 2)-1)
            rest_aomunt=rest_aomunt-amount
            rest_people -= 1
            amount_list.append(amount)
        amount_list.append(rest_aomunt)
        print(amount_list)
        return {"count":1,"reduce":amount_list}
        
    def send(redbagSender:str,coin_num:int,num:int,leavingAMessage:str=defaultBlessingWords):
        """
        发送红包
        coin_num -> 金额
        num -> 数量
        leavingAMessage -> 祝福语
        错误则返回dict
        """

        if coin_num< 0 or coin_num==0:
            return {"count":0,"error":"§a小世界§e>> §c错误:红包金额小于或者等于0§r","value":False}
        elif num<0 or num ==0:
            return {"count":0,"error":"§a小世界§e>> §c错误:红包数量小于或者等于0§r","value":False}
        else:
            if not os.path.isdir(f"data/player/{redbagSender}"):
                os.mkdir(f"data/player/{redbagSender}")
            redbagdata=getPlayerData('redbag',redbagSender)
            if redbagdata == "":
                redbagSendTime=time.strftime("%Y年%m月%d日 %H时%M分%S秒")
                coin_Num=coin_num
                Num=num
                calculationReturnResult=redbag.algorithm(coin_Num,Num)
                if calculationReturnResult['count'] == 0:
                    return calculationReturnResult
                elif calculationReturnResult["count"] == 1:
                    顺序=calculationReturnResult['reduce']
                writedata={"玩家名":redbagSender,"红包金额":coin_num,"红包数量":num,"附加的消息":leavingAMessage,"发送的日期":f"{redbagSendTime}","领取顺序":顺序,"已领玩家":[]}
                writedata_one= {"1":writedata}
                
                with open(os.path.join('data','player',f"{redbagSender}",'redbag.txt'),"w",encoding="utf-8") as f:
                    f.write(json.dumps(writedata_one,ensure_ascii=False))
                with open(os.path.join('data','redbag_Data.json'),"r",encoding="utf-8") as f:
                    redbag_GetData=f.read()
                redbag_GetData=ast.literal_eval(redbag_GetData)
                TheOnlyCode=str(uuid.uuid4()).split("-")[0]
                redbag_GetData["信息"][f'{TheOnlyCode}'] = [f"{redbagSender}","1"]
                with open(os.path.join('data','redbag_Data.json'),"w",encoding="utf-8") as f:
                    f.write(json.dumps(redbag_GetData,ensure_ascii=False,indent=4))
                return {"count":1,"value":True,"code":TheOnlyCode}
            else:
                redbagSendTime=time.strftime("%Y年%m月%d日 %H时%M分%S秒")
                calculationReturnResult=redbag.algorithm(coin_num,num)
                if calculationReturnResult['count'] == 0:
                    return calculationReturnResult
                elif calculationReturnResult["count"] == 1:
                    顺序=calculationReturnResult['reduce']
                writedata={"玩家名":redbagSender,"红包金额":coin_num,"红包数量":num,"附加的消息":leavingAMessage,"发送的日期":f"{redbagSendTime}","领取顺序":顺序,"已领玩家":[]}
                redbagdata=ast.literal_eval(redbagdata)
                redbag_addnum=str(len(redbagdata)+1)
                redbagdata[redbag_addnum] = writedata
                with open(os.path.join('data','player',f"{redbagSender}",'redbag.txt'),"w",encoding="utf-8") as f:
                    f.write(json.dumps(redbagdata,ensure_ascii=False))
                with open(os.path.join('data','redbag_Data.json'),"r",encoding="utf-8") as f:
                    redbag_GetData=f.read()
                redbag_GetData=ast.literal_eval(redbag_GetData)
                TheOnlyCode=str(uuid.uuid4()).split("-")[0]
                redbag_GetData["信息"]["%s"%TheOnlyCode] = [redbagSender,str(redbag_addnum)]
                with open(os.path.join('data','redbag_Data.json'),"w",encoding="utf-8") as f:
                    f.write(json.dumps(redbag_GetData,ensure_ascii=False,indent=4))
                
                return {"count":1,"value":True,"code":TheOnlyCode}
    def receive(code:str,player:str):
        """
        领取红包
        code -> 红包唯一识别码(uuid)
        player -> 领取玩家名
        """
        with open(os.path.join('data','redbag_Data.json'),"r",encoding="utf-8") as f:
            redbag_GetData=f.read()
        redbag_GetData=ast.literal_eval(redbag_GetData)
        if code not in redbag_GetData['信息']:
            return {"count":0,"error":f"§6红包系统§e>> §c未找到红包","value":False}
        else:
            pass
        redbagSender=redbag_GetData['信息'][code][0]
        redbagNum = redbag_GetData['信息'][code][1]
        
        if not os.path.isdir(os.path.join('data','player',f"{redbagSender}")):
            os.mkdir(os.path.join('data','player',f"{redbagSender}"))
        GetPlayerRedbagData=getPlayerData("redbag",redbagSender)
        GetPlayerRedbagData=ast.literal_eval(GetPlayerRedbagData)
        if GetPlayerRedbagData[f"{redbagNum}"]["红包数量"] == 0: #红包金额 红包数量 顺序
            return {"count":0,"error":f"§6{redbagSender}§e>> §r你的手速太慢了,§c红包§r已经领完了","value":False}
        elif player in GetPlayerRedbagData[f'{redbagNum}']["已领玩家"]:
            return {"count":0,"error":f"§6{redbagSender}§e>> §c你的已经领取过了","value":False}
        else:
            RemCoin=GetPlayerRedbagData[f'{redbagNum}']["领取顺序"][0]
            GetPlayerRedbagData[f'{redbagNum}']["红包金额"] -= RemCoin
            GetPlayerRedbagData[f'{redbagNum}']["红包数量"] -= 1
            del(GetPlayerRedbagData[f'{redbagNum}']["领取顺序"][0])
            GetPlayerRedbagData[f'{redbagNum}']["已领玩家"].append(player)
            with open(os.path.join('data','player',f"{redbagSender}",'redbag.txt'),"w",encoding="utf-8") as f:
                f.write(json.dumps(GetPlayerRedbagData,ensure_ascii=False))
            return {"count":1,"command":f'/scoreboard players add {player} {scoreboardName} {RemCoin}',"reduce":f"§6红包系统§e>> §r成功领取§e{RemCoin}¥","value":True}
    #暂未使用
    # def back(player:str):
    #     if not os.path.isdir(f"data/player/{player}"):
    #         os.mkdir(f"data/player/{player}")
    #     backtoplayerData=getPlayerData("redbag",player)
    #     if len(backtoplayerData) !=0:
    #         backtoplayerData=ast.literal_eval(backtoplayerData)
    #         for i in backtoplayerData:
    #             for i in backtoplayerData[i]["领取顺序"]:
    #                 API.do_send_wo_cmd(f"/scoreboard players add {player} {scoreboardName} {i}",cb=None)
    #                 API.do_send_player_msg(player,f"§6红包系统§a>> §f归还红包§e{i}¥")
    #         return {"count":1,"value":True}
    #     else:
    #         return {"count":1,"value":False}



def Mono_plugin_redbag(api:API):
    api.do_echo("MonoMenu红包系统_omeagside已启动",cb=None)
    def on_menu_invoked_redbag_one(player_input:PlayerInput):
        player=player_input.Name
        if len(player_input.Msg) in [0,1]: #只输入了 omg发红包
            api.do_send_player_msg(player,"§r输入§bomg发红包 <金额> <数量> <祝福语> 发送红包",cb=None)#\n§r输入§bomg领/抢 <代号> §r领取红包
        elif len(player_input.Msg) in [2,3]: #输入了omg发红包 <金额> <数量> <或者也许输入了祝福语>
            if len(player_input.Msg) ==2: #如果没有祝福语
                response=api.do_get_scoreboard(cb=None)
                api.do_send_player_msg(player_input.Name,f"§r你当前拥有金币:§e{response[scoreboardName][player_input.Name]}§r",cb=None)
                if response[scoreboardName][player_input.Name] < int(player_input.Msg[0]): #如果玩家输入的红包金额大于自己的金额的话
                    api.do_send_player_msg(player_input.Name,"§c金额不足§r",cb=None)
                elif response[scoreboardName][player_input.Name] >= int(player_input.Msg[0]):
                    #api.do_send_player_msg(player_input.Name,"§a准备发送ing§r",cb=None)
                    result=redbag.send(player,int(player_input.Msg[0]),int(player_input.Msg[1]))
                    if result["value"]:
                        api.do_send_player_msg(player,"§6红包系统§e>> §3发送成功§r",cb=None)
                        #api.do_send_player_msg(player="@a",msg="测试001",cb=None)
                        api.do_send_player_msg(player="@a",msg=f"§6{player}§e>> §f发送了§c红包§e{player_input.Msg[0]}¥ 0/{player_input.Msg[1]} 祝福:§e{defaultBlessingWords}§f\n§f红包代号:{result['code']}",cb=None)
                        api.do_send_wo_cmd('/title @a title §r',cb=None)
                        api.do_send_wo_cmd(f'/title @a subtitle §e{player}§f发送了§c红包§f\n输入§eomg抢 <代号>§f领取',cb=None)
                    elif result["value"] == False:
                        api.do_send_player_msg(player,result["error"])
                
            elif len(player_input.Msg) ==3: #有祝福语
                response=api.do_get_scoreboard(cb=None)
                api.do_send_player_msg(player_input.Name,f"§r你当前拥有金币:§e{response[scoreboardName][player_input.Name]}§r",cb=None)
                if response[scoreboardName][player_input.Name] < int(player_input.Msg[1]): #如果玩家输入的红包金额大于自己的金额的话
                    api.do_send_player_msg(player_input.Name,"§c金额不足§r",cb=None)
                elif response[scoreboardName][player_input.Name] >= int(player_input.Msg[1]):
                    result=redbag.send(player,int(player_input.Msg[0]),int(player_input.Msg[1]))
                    if result["value"]:
                        api.do_send_player_msg(player,"§6红包系统§e>> §3发送成功§r",cb=None)
                        api.do_send_player_msg(player="@a",msg=f"§6{player}§e>> §f发送了§c红包§e{player_input.Msg[0]}¥ 0/{player_input.Msg[1]} 祝福:§e{player_input.Msg[2]}§f\n§f红包代号:{result['code']}",cb=None)
                        api.do_send_wo_cmd('/title @a title §r',cb=None)
                        api.do_send_wo_cmd(f'/title @a subtitle §e{player}§f发送了§c红包§f\n输入§eomg抢 <代号>§f领取',cb=None)
                    elif result["value"] == False:
                        api.do_send_player_msg(player,result["error"],cb=None)
        
                
    def on_menu_invoked_redbag_two(player_input:PlayerInput):
        player=player_input.Name
        if len(player_input.Msg) != 1:
            api.do_send_player_msg(player,"§r输入§bomg领/抢 <代号> 领取红包",cb=None)
        elif len(player_input.Msg) == 1:
            api.do_send_player_msg(player_input.Name,f"§a领取代号:§e{player_input.Msg[0]}§r\n§r",cb=None)
            result=redbag.receive(player_input.Msg[0],player_input.Name)
            if result["value"]:
                api.do_send_wo_cmd(result['command'],cb=None)
                api.do_send_player_msg(player,result["reduce"],cb=None)
                
            else:
                api.do_send_player_msg(player,result["error"],cb=None)

    api.listen_omega_menu(triggers=["发红包"],argument_hint="",usage="发红包 [金额] [数量] [祝福语(选填)]",cb=None,on_menu_invoked=on_menu_invoked_redbag_one)
    api.listen_omega_menu(triggers=["领","抢"],argument_hint="",usage="领红包 领/抢 [代号]",cb=None,on_menu_invoked=on_menu_invoked_redbag_two)

omega.add_plugin(plugin=Mono_plugin_redbag)


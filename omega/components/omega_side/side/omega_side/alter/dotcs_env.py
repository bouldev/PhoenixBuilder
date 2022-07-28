from collections import defaultdict
from posixpath import split
from threading import Event
from typing import *
import time
import os
import json

from omega_side.python3_omega_sync.protocol import CmdResult, PlayerInfo
from ..python3_omega_sync import API
from ..python3_omega_sync import frame as omega
from ..python3_omega_sync import bootstrap

orig_print=print
def alter_print(*args,**kwargs):
    kwargs["flush"]=True
    orig_print(*args,**kwargs)
print=orig_print

def color(text: str, output: bool = True, end: str = "\n", replace: bool = False, replaceByNext: bool = False) -> str:
    # 所有在omega 环境下print 出来的东西会自动变颜色，且会自动转为日志，所以这两个函数并没有什么意义
    if output:
        print(("\r" if replace else "")+ text,end="" if replaceByNext else "\n")
    return text

def log(text: str, filename: str = None, mode: str = "a", encoding: str = "utf-8", errors: str = "ignore", output: bool = True, sendtogamewithRitBlk: bool = False, sendtogamewithERROR: bool = False, sendtogrp: bool = False) -> None:
    print(text)

def countdown(delay: Union[int,float], msg: str = "Countdown", untilPaid:bool = False):
    deadline = time.time()+delay
    current_time=time.time()
    while current_time+1<deadline :
        print(f"\r{msg:s}: {deadline-current_time:>5.2f}s",end="")
        time.sleep(1)
        current_time=time.time()
    print()
    
def exitChatbarMenu(killFB: bool = True, delay: Union[int,float] = 3, reason: str = None):
    exit(0)
    
def is_port_used(port: int)->bool:
    pass

def FBkill():
    pass

def runFB(killFB: bool = True):
    pass

def Byte2KB(byteSize: float) -> str:
    for i,unit in enumerate(("B", "KB", "MB", "GB", "TB", "PB", "EB")):
        if byteSize<2**(10*(i+1)):  return f"{byteSize/(2**(10*i)):.2f}{unit}"

def fileDownload(url: str, path: str, timeout: Union[int,float] = 3, freshSize: int = 10240) -> dict:
    try:
        bootstrap.download_file(url=url,local_filename=path,chunk_size=freshSize,timeout=timeout)
    except Exception as e:
        return {"status": "fail","reason":e}
    return {"status": "success"}

sendtogroup = ""
QQgroup = ""

api:API=None
init_lock=Event()

def api_taker(_api:API):
    global api 
    init_lock.set()
    api=_api

omega.add_plugin(api_taker)
bootstrap.execute_func_in_thread_with_auto_restart(omega.run)
init_lock.wait()
print("\033[32mOmega DotCS Emulator 已经链接到 Omega 框架中\033[0m")

def nop(*args,**kwargs):
    pass

def sendtogroup(where="group", number:int=0, message:str=""):
    api.do_send_qq_msg(msg=message,cb=nop) 

def tellrawText(who: str, dispname:str = None, text: str = None):
    global api
    api.do_send_player_msg(who,text if dispname is None else f"<{dispname}> "+text,cb=nop)
    
def tellrawScore(scoreboardName: str, targetName: str) -> str:
    return '{"score":{"name":"'+targetName+'","objective":"'+scoreboardName+'}"}}'

def getPlayerData(dataName: str, playerName: str, writeNew: str = "") -> any:
    response=api.do_get_player_data(player=playerName,entry=dataName,cb=None)
    if response.found:
        try:
            return int(response.data)
        except:
            return response.data
    else:
        api.do_set_player_data(player=playerName,entry=dataName,data=writeNew,cb=None)
        try:
            return int(writeNew)
        except:
            return writeNew

def addPlayerData(dataName: str, playerName: str, dataValue, dataType: str = "int", writeNew: str = ""):
    response=api.do_get_player_data(player=playerName,entry=dataName,cb=None)
    if dataType == "int":
        if response.found:
            if response.data=="":
                response.data=0
            new_val=response.data+dataValue
        else:
            new_val=dataValue
        api.do_set_player_data(player=playerName,entry=dataName,data=new_val,cb=None)
        return new_val
    new_val=response.data
    if response.found:
        new_val+=dataValue     
    else:
        new_val=dataValue
    api.do_set_player_data(player=playerName,entry=dataName,data=new_val,cb=None)

def setPlayerData(dataName: str, playerName: str, dataValue:any, writeNew: any = ""):
    # response=api.do_get_player_data(player=playerName,entry=dataName,cb=None)
    # if response.found:
    api.do_set_player_data(player=playerName,entry=dataName,data=dataValue,cb=None)
    # else:
        #api.do_set_player_data(player=playerName,entry=dataName,data=writeNew,cb=None)
        
def getType(sth):
    return sth.__class__.__name__

def float2int(number: float, way: int = 1) -> int:
    return (round,int,lambda x:round(x+0.5))[way](number)

def second2minsec(sec: int) -> str:
    hour,min,sec=sec//3600,(sec//60)%60,sec%60
    return f"{hour:0>2d}:{min:0>2d}:{sec:0>2d}"

def getTarget(sth: str, timeout= 1) -> list:
    response=api.do_send_ws_cmd(f"tell @s {sth}",cb=None)
    try:
        return response.result.OutputMessages[0].Parameters[1:].split(", ")
    except Exception as e:
        return Exception("Target not found.")


def getScore(scoreboardNameToGet: str, targetNameToGet: str):
    msgs = sendcmd("/scoreboard players list %s" % targetNameToGet, True).OutputMessages
    player_scoreboard_result = defaultdict(dict)
    scoreboard_palyer_result = defaultdict(dict)
    current_player_name=""
    for msg in msgs:
        msg_type = msg.Message
        if msg_type == "commands.scoreboard.players.list.player.empty":
            continue
        elif msg_type == "§a%commands.scoreboard.players.list.player.count":
            current_player_name = msg.Parameters[1][1:]
        elif msg_type == "commands.scoreboard.players.list.player.entry":
            if current_player_name == "commands.scoreboard.players.offlinePlayerName":
                continue
            scoreboard_name = msg.Parameters[2]
            targetScore = int(msg.Parameters[0])
            player_scoreboard_result[current_player_name][scoreboard_name] = targetScore
            scoreboard_palyer_result[scoreboard_name][current_player_name] = targetScore
    if not(player_scoreboard_result or scoreboard_palyer_result):
        raise Exception("Failed to get the score.")
    try:
        if targetNameToGet == "*" or targetNameToGet.startswith("@"):
            if scoreboardNameToGet == "*":
                return [player_scoreboard_result, scoreboard_palyer_result]
            else:
                return scoreboard_palyer_result[scoreboardNameToGet]
        else:
            if scoreboardNameToGet == "*":
                return player_scoreboard_result[targetNameToGet]
            else:
                return player_scoreboard_result[targetNameToGet][scoreboardNameToGet]
    except KeyError as err:
        raise Exception(f"Failed to get score: {err}")


def getPos(targetNameToGet: str) -> dict:
    response = sendwscmd(f"/querytarget {targetNameToGet}", True).OutputMessages[0]
    if response.Success == 0:
        raise Exception("Failed to get the position.")
    translated_position = {}
    player_list_response=api.do_get_players_list(cb=None)
    uuid_name_mapping={}
    for player in player_list_response:
        uuid_name_mapping[player.uuid]=player.name
    for parameter in json.loads(response.Parameters[0]):
        parameter["position"]["y"]-= 1.6200103759765
        translated_position[uuid_name_mapping[parameter["uniqueId"]]]=parameter
    if targetNameToGet == "@a":
        return translated_position
    else:
        if len(translated_position) != 1:
            raise Exception("Failed to get the position.")
        if targetNameToGet.startswith("@a"):
            return list(translated_position.values())[0]
        else:
            return translated_position[targetNameToGet]


def getItem(targetName: str, itemName: str, itemSpecialID: int = -1) -> int:
    result = sendcmd(f"/clear {targetName} {itemName} {itemSpecialID} 0", True)
    if result.OutputMessages[0].Message == "commands.generic.syntax":
        raise Exception("Item name error.")
    if result.OutputMessages[0].Message == "commands.clear.failure.no.items":
        return 0
    else:
        return int(result.OutputMessages[0].Parameters[1])


def getStatus(statusName:str)->any:
    response=api.do_get_player_data(player="_dotcs_status",entry=statusName,cb=None)
    if response.found:
        return response.data
    else:
        return None

def setStatus(statusName: str, status:any):
    api.do_set_player_data(player="_dotcs_status",entry=statusName,data=status,cb=None)

def QRcode(text: str, where: str = "console", who: str= None) -> None:
    bootstrap.install_lib("qrcode")
    import qrcode
    if where not in ("console","server") or (where=="server" and who is None):
        raise Exception("invalid argument")
    if where=="console":
        block={True:"\033[0m  ",False:"\033[0;37;7m  "} 
        display_line=print
    else:
        block= {True:"§0▓",False:"§f▓"}
        display_line=lambda x:tellrawText(who, text = "§l"+x)
    qr = qrcode.QRCode()
    qr.add_data(text)
    for row in qr.get_matrix():
        display_line("".join([block[col] for col in row]))

def sendcmd(cmd: str, waitForResponse: bool = False)->Union[None,CmdResult]:
    cmd=cmd[1:] if cmd.startswith("/") else cmd 
    if waitForResponse:
        response=api.do_send_player_cmd(cmd,cb=None)
        return response.result
    else:
        api.do_send_player_cmd(cmd,cb=nop)

def sendwscmd(cmd: str, waitForResponse: bool = False)->Union[None,CmdResult]:
    cmd=cmd[1:] if cmd.startswith("/") else cmd 
    if waitForResponse:
        response=api.do_send_ws_cmd(cmd,cb=None)
        return response.result
    else:
        api.do_send_ws_cmd(cmd,cb=nop)

def sendwocmd(cmd:str):
    cmd=cmd[1:] if cmd.startswith("/") or cmd.startswith("!") else cmd 
    api.do_send_wo_cmd(cmd,cb=None)

def sendfbcmd(cmd:str):
    api.do_send_fb_cmd(cmd,cb=None)

def strInList(str: str, list: list) -> bool:
        for i in list:
            if str in i: return True
        return False

server="omega_server"
version="omega_adapt_dotcs"

allplayers=[]
all_players_dict={}
msgList = []
rev = ""
robotname = ""
timesErr = 0
msgRecved = False
entityRuntimeID2playerName = {}
XUID2playerName = {}
msgLastRecvTime = time.time()
itemNetworkID2NameDict = {}
itemNetworkID2NameEngDict = {}
adminhigh = []
needToGetMainhandItem = False
needToGetArmorItem = False
needToGetMainhandAndArmorItem = False
targetMainhandAndArmor = ""
itemMainhandAndArmor = ""
targetArmor = ""
targetMainhand = ""

def on_player_login(player:PlayerInfo):
    global allplayers,all_players_dict,XUID2playerName
    if player.name not in all_players_dict.keys():
        all_players_dict[player.name]=True
        allplayers.append(player.name)
    XUID2playerName[player.uuid]=player.name

def on_player_logout(player:PlayerInfo):
    global allplayers,all_players_dict
    if player.name in all_players_dict.keys():
        del all_players_dict[player.name]
        allplayers=list(all_players_dict.keys())

api.listen_player_login(cb=None,on_player_login_cb=on_player_login)
api.listen_player_logout(cb=None,on_player_logout_cb=on_player_logout)

try:
    robotname=api.do_send_ws_cmd("testfor @s",cb=None).result.OutputMessages[0].Parameters[0]
except Exception as e:
    print(e)
response=api.do_get_item_mapping(cb=None)
for runtime_id,item in response.mapping.items():
    itemNetworkID2NameDict[int(runtime_id)]=item["name"].replace("minecraft:","")
    itemNetworkID2NameEngDict[int(runtime_id)]=item["name"].replace("minecraft:","")
    
def simplify_name(name:str):
    try:
        name = name.replace(">§r", "").split("><")[1]
        return name
    except:
        return name

player_message_dotcs_cbs=[]
def listen_player_message(cb:Callable[[str,str,str],None]):
    player_message_dotcs_cbs.append(cb)

def launch_player_message_cbs(text_type,player_name,msg):
    for cb in player_message_dotcs_cbs:
        cb(text_type,player_name,msg)

player_prejoin_cbs=[]
def listen_player_prejoin(cb:Callable[[str,str,str],None]):
    player_prejoin_cbs.append(cb)

player_join_cbs=[]
def listen_player_join(cb:Callable[[str,str,str],None]):
    player_join_cbs.append(cb)

player_leave_cbs=[]
def listen_player_leave(cb:Callable[[str,str,str],None]):
    player_leave_cbs.append(cb)

player_death_cbs=[]
def listen_player_death(cb:Callable[[str,str,str,str],None]):
    player_death_cbs.append(cb)

def launch_prejoin_cbs(text_type,player_name,msg):
    for cb in player_prejoin_cbs:
        cb(text_type,player_name,msg)

def launch_join_cbs(text_type,player_name,msg):
    for cb in player_join_cbs:
        cb(text_type,player_name,msg)

def launch_leave_cbs(text_type,player_name,msg):
    for cb in player_leave_cbs:
        cb(text_type,player_name,msg)


def launch_death_cbs(text_type,player_name,msg,killer):
    for cb in player_leave_cbs:
        cb(text_type,player_name,msg,killer)

def player_message_listener(pkt):
    global allplayers,all_players_dict,msgList,rev,robotname,timesErr,msgRecved,entityRuntimeID2playerName,XUID2playerName,msgLastRecvTime,itemNetworkID2NameDict,itemNetworkID2NameEngDict,needToGetMainhandItem,needToGetArmorItem,needToGetArmorItem,needToGetMainhandAndArmorItem,targetMainhandAndArmor,itemMainhandAndArmor,targetArmor,targetMainhand
    text_type,player_name,msg=pkt["TextType"],pkt["SourceName"],pkt["Message"]
    player_name=simplify_name(player_name)
    if "alive" in msg:
        return
    if text_type == 8:
        msg = msg.split("] ", 1)
        if len(msg)>0:
            msg=msg[1]
        else:
            msg=""
    elif text_type == 9:
        msg = msg.replace('{"rawtext":[{"text":"', "").replace('"}]}', "").replace('"},{"text":"', "").replace(r"\n", "\n"+str(text_type)+" ").replace("§l", "")
        if len(msg)>0 and msg[-1] == "\n":
            msg = msg[:-1]
    elif text_type ==2:
        if msg == "§e%multiplayer.player.joining":
            playername = pkt["Parameters"][0]
            bootstrap.execute_func_in_thread_with_auto_restart(launch_prejoin_cbs,text_type,player_name,msg)
        elif msg == "§e%multiplayer.player.joined":
            playername = pkt["Parameters"][0]
            if playername not in all_players_dict.keys():
                all_players_dict[playername]=True
                allplayers.append(playername)
            bootstrap.execute_func_in_thread_with_auto_restart(launch_join_cbs,text_type,player_name,msg)

        elif msg == "§e%multiplayer.player.left":
            playername = pkt["Parameters"][0]
            if playername in all_players_dict.keys():
                    del all_players_dict[playername]
                    allplayers=list(all_players_dict.keys())
            bootstrap.execute_func_in_thread_with_auto_restart(launch_leave_cbs,text_type,player_name,msg)
            
        elif msg[0:6] == "death.":
            playername = pkt["Parameters"][0]
            if len(pkt["Parameters"]) == 2:
                killer = pkt["Parameters"][1]
            else:
                killer = None
            bootstrap.execute_func_in_thread_with_auto_restart(launch_leave_cbs,text_type,player_name,killer)
    elif text_type in (1,7,8):                  
        for cb in player_message_dotcs_cbs:
            cb(text_type,player_name,msg)
        bootstrap.execute_func_in_thread_with_auto_restart(launch_player_message_cbs,text_type,player_name,msg)

api.listen_mc_packet("IDText",cb=None,on_new_packet_cb=player_message_listener)

def add_player_listener(jsonPkt):
    global allplayers,all_players_dict,msgList,rev,robotname,timesErr,msgRecved,entityRuntimeID2playerName,XUID2playerName,msgLastRecvTime,itemNetworkID2NameDict,itemNetworkID2NameEngDict,needToGetMainhandItem,needToGetArmorItem,needToGetArmorItem,needToGetMainhandAndArmorItem,targetMainhandAndArmor,itemMainhandAndArmor,targetArmor,targetMainhand
    try:
        entityRuntimeID2playerName[jsonPkt["EntityRuntimeID"]] = jsonPkt["Username"]
    except:
        pass
    if needToGetMainhandItem and targetMainhand == jsonPkt["Username"]:
        itemMainhand = jsonPkt["HeldItem"]["Stack"]
        try:
            itemMainhand["itemName"] = itemNetworkID2NameDict[str(jsonPkt["HeldItem"]["Stack"]["NetworkID"])]
            itemMainhand["itemCmdName"] = itemNetworkID2NameEngDict[str(jsonPkt["HeldItem"]["Stack"]["NetworkID"])]
        except:
            itemMainhand["itemName"] = "未知"
            itemMainhand["itemCmdName"] = "unknown"
        needToGetMainhandItem = False
    if needToGetMainhandAndArmorItem and targetMainhandAndArmor == jsonPkt["Username"]:
        itemMainhandAndArmor["mainHand"] = jsonPkt["HeldItem"]["Stack"]
        try:
            itemMainhandAndArmor["mainHand"]["itemName"] = itemNetworkID2NameDict[str(jsonPkt["HeldItem"]["Stack"]["NetworkID"])]
            itemMainhandAndArmor["mainHand"]["itemCmdName"] = itemNetworkID2NameEngDict[str(jsonPkt["HeldItem"]["Stack"]["NetworkID"])]
        except:
            itemMainhandAndArmor["mainHand"]["itemName"] = "未知"
            itemMainhandAndArmor["mainHand"]["itemCmdName"] = "unknown"
api.listen_mc_packet("IDAddPlayer",cb=None,on_new_packet_cb=add_player_listener)


def mob_armour_equipment(jsonPkt):
    global allplayers,all_players_dict,msgList,rev,robotname,timesErr,msgRecved,entityRuntimeID2playerName,XUID2playerName,msgLastRecvTime,itemNetworkID2NameDict,itemNetworkID2NameEngDict,needToGetMainhandItem,needToGetArmorItem,needToGetArmorItem,needToGetMainhandAndArmorItem,targetMainhandAndArmor,itemMainhandAndArmor,targetArmor,targetMainhand
    try:
        entityRuntimeID2playerName[jsonPkt["EntityRuntimeID"]] = jsonPkt["Username"]
    except:
        pass
    if needToGetMainhandItem and targetMainhand == jsonPkt["Username"]:
        itemMainhand = jsonPkt["HeldItem"]["Stack"]
        try:
            itemMainhand["itemName"] = itemNetworkID2NameDict[str(jsonPkt["HeldItem"]["Stack"]["NetworkID"])]
            itemMainhand["itemCmdName"] = itemNetworkID2NameEngDict[str(jsonPkt["HeldItem"]["Stack"]["NetworkID"])]
        except:
            itemMainhand["itemName"] = "未知"
            itemMainhand["itemCmdName"] = "unknown"
        needToGetMainhandItem = False
    if needToGetMainhandAndArmorItem and targetMainhandAndArmor == jsonPkt["Username"]:
        itemMainhandAndArmor["mainHand"] = jsonPkt["HeldItem"]["Stack"]
        try:
            itemMainhandAndArmor["mainHand"]["itemName"] = itemNetworkID2NameDict[str(jsonPkt["HeldItem"]["Stack"]["NetworkID"])]
            itemMainhandAndArmor["mainHand"]["itemCmdName"] = itemNetworkID2NameEngDict[str(jsonPkt["HeldItem"]["Stack"]["NetworkID"])]
        except:
            itemMainhandAndArmor["mainHand"]["itemName"] = "未知"
            itemMainhandAndArmor["mainHand"]["itemCmdName"] = "unknown"
api.listen_mc_packet("IDMobArmourEquipment",cb=None,on_new_packet_cb=mob_armour_equipment)

def repeat_exec(cb:Callable,repeat_time):
    api.execute_with_repeat(cb,repeat_time=repeat_time)

def listen_packet(cb:Callable[[dict],None],packet_id):
    api.listen_mc_packet(pkt_type=packet_id,cb=None,on_new_packet_cb=cb)

install_lib=bootstrap.install_lib
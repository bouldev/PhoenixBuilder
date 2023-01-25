# 插件: 关


from threading import Thread
import time,os
from omega_side.python3_omega_sync import API
from omega_side.python3_omega_sync import frame as omega
from omega_side.python3_omega_sync.bootstrap import install_lib
from omega_side.python3_omega_sync.protocol import *
class version_packet_IDBlockActorData:
    author     =    "Mono"
    version    =    "2.0 for omegaside"


if not os.path.isdir(os.path.join('data','物品放置日志')):
    os.mkdir(os.path.join('data','物品放置日志'))


def MonoMenu_plugin_getblockdata(api:API):
    def testforIDBlockActorData(packet:dict):
        if "NBTData" in packet:
            if "id" in packet["NBTData"]:
                Timedate = time.localtime(time.time())
                TimeList = list(Timedate)
                x,y,z= packet['Position'][0],packet['Position'][1],packet['Position'][2]
                try:
                    PlayerName=api.do_send_ws_cmd(f"/testfor @a[x={x},y={y},z={z},r=10]",cb=None)["result"]["OutputMessages"][0]["Parameters"][0]
                except:
                    PlayerName="unknow"
                if packet["NBTData"]["id"] == "Sign":
                    if packet['NBTData']["Text"] != "":
                        SignText=packet['NBTData']["Text"]
                        if PlayerName!="":
                            with open(os.path.join("data","物品放置日志","[物品放置]log.txt"),"a+",encoding="utf-8") as f:
                                f.write(f"\n[{TimeList[1]}.{TimeList[2]} {TimeList[3]}:{TimeList[4]}:{TimeList[5]}]坐标:[{x},{y},{z}]告示牌内容:{SignText},最近的玩家:{PlayerName}")
                elif packet["NBTData"]["id"] == "ShulkerBox":
                    
                    
                    for i in packet['NBTData']['Items']:
                        shulkerboxItems=f"\n[{TimeList[1]}.{TimeList[2]} {TimeList[3]}:{TimeList[4]}:{TimeList[5]}]坐标:[{x},{y},{z}]"
                        shulkerboxItems += f"物品:{i['Name']} 栏位:{i['Slot']} 数量:{i['Count']},"
                        if "tag" in i:
                            if "display" in i["tag"]:
                                if "Name" in i ["tag"]["display"]:
                                    shulkerboxItems += f'名称:{i["tag"]["display"]["Name"]},'
                            if "ench" in i["tag"]:
                                for x in i["tag"]["ench"]:
                                    shulkerboxItems += f'附魔id:{x["id"]}附魔等级:{x["lvl"]},'
                        if PlayerName !="unknow" :
                            shulkerboxItems+="最近的玩家:%s"%PlayerName
                        with open(os.path.join("data","物品放置日志","[物品放置]log.txt"),"a+",encoding="utf-8") as f:
                            f.write(shulkerboxItems)
    def getIDBlockActorData_pakcet(packet):
        Thread(target=testforIDBlockActorData,args=[packet]).start()
    api.listen_mc_packet(pkt_type="IDBlockActorData",cb=None,on_new_packet_cb=getIDBlockActorData_pakcet)
omega.add_plugin(plugin=MonoMenu_plugin_getblockdata)

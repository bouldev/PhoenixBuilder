# 插件: 开

from threading import Thread
import time
from omega_side.python3_omega_sync import API
from omega_side.python3_omega_sync import frame as omega
from omega_side.python3_omega_sync.bootstrap import install_lib
from omega_side.python3_omega_sync.protocol import *

#声明:此插件原作者为super
class version_井字棋:
    移植_author     =    "Mono"
    version    =    "2.0 for omegaside"
JZQ_Rooms = []

def searchName(key,data) -> list:
    """
    模糊查找器
    key: 关键字
    data: 数据
    :return: list
    """
    data_re = [i for i in data if key in i]
    return data_re

def Mono_plugin_井字棋(api:API):
    class JZQStage:
        def __init__(self):
                self.Stage = [    "basic_0",
                "§7▒§f", "§7▒§f", "§7▒§f",
                "§7▒§f", "§7▒§f", "§7▒§f",
                "§7▒§f", "§7▒§f", "§7▒§f"
                ]

                self.turn = 0

                self.luozi_type = ["§f▒§f", "§0▒§f"]

                self.panding = [
                    (1, 2, 3), (4, 5, 6), (7, 8, 9), 
                    (1, 4, 7), (2, 5, 8), (3, 6, 9),
                    (1, 5, 9), (3, 5, 7)
                ]

                self.__time = 180
        def 轮流(self, turn = False):
            if turn:
                if self.turn == 0:
                    self.turn = 1
                else:
                    self.turn = 0
            return self.turn
        def 落子(self, xpos, ypos, typePlayer):
            pos = self.Stage[(xpos - 1) * 3 + ypos]
            if pos == "§7▒§f":
                self.Stage[(xpos - 1) * 3 + ypos] = self.luozi_type[typePlayer]
                return True
            else:
                return False

        def 判定(self):
            for i in self.panding:
                pos1, pos2, pos3 = i
                if self.Stage[pos1] == self.Stage[pos2] == self.Stage[pos3] and self.Stage[pos1] != "§7▒§f":
                    return True
            return False

        def 判死(self):
            return not "§7▒§f" in self.Stage

        def 重置(self, done = False):
            self.__time = 180
            self.Stage = [        "basic_0",
                "§7▒§f", "§7▒§f", "§7▒§f",
                "§7▒§f", "§7▒§f", "§7▒§f",
                "§7▒§f", "§7▒§f", "§7▒§f"
                ]
            if done:
                self.turn = 0
                

        def Timer(self, time = None):
            if time:
                self.__time += time
            return self.__time

        def display(self):
            basic, ps1, ps2,ps3, ps4, ps5, ps6, ps7, ps8, ps9\
                = self.Stage
            return \
                  ps1 + ps2 + ps3 + "\n" \
                + ps4 + ps5 + ps6 + "\n" \
                + ps7 + ps8 + ps9

        def stage_display(self, index: list, playername: str, end = False):
            api.do_send_ws_cmd("/title {player} actionbar §e§l井字棋\n§b时间: §c{minute}§f:§c{second}§r §6落子:{turn} {iftimedout}\n§l{stage}\n§9© Copyright Yisheng-SuperScript.Co.Ltd".format(
                player = playername,
                minute = "**" if end else self.Timer() // 60,
                second = "**" if end else self.Timer() % 60,
                turn = "§a✔" if Game_JZQ.轮流() == index.index(playername) else "§c✘",
                stage = Game_JZQ.display(),
                iftimedout = "§c时间到!" if self.Timer() == 0 else ""),cb=None
            )
    Game_JZQ = JZQStage()
    def 井字棋(playername:str,msg:str):
        if not JZQ_Rooms:
            try:
                to_who = msg.split()[1]
            except:
                to_who = ""
            response=api.do_get_players_list(cb=None)
            allplayers=[]
            for i in response:
                allplayers.append(i.name)
            playernameList=searchName(to_who,allplayers)
            if len(playernameList) !=1:
                api.do_send_player_msg(playername, "§a井字棋§f>> §c玩家未找到!.",cb=None)
                return
            to_who = playernameList[0]
            if to_who in allplayers and (not playername == to_who):
                JZQ_Rooms.append([playername, to_who])
                api.do_send_player_msg(playername, "§a井字棋§f>> §a成功开启游戏.",cb=None)
            else:
                api.do_send_player_msg(playername, "§a井字棋§f>> §c玩家未找到!.",cb=None)
        else:
            api.do_send_player_msg(playername, "§a井字棋§f>> §c房间里正在游戏中!",cb=None)
    def 下子(playername:str,msg:str):

        for i in JZQ_Rooms:
            if playername in i:
                typePlayer = i.index(playername)
                try:
                    x_xpos = int(msg.split()[1])
                    y_ypos = int(msg.split()[2])
                except:
                    api.do_send_player_msg(playername, "§a井字棋§f>> §c下子格式有误",cb=None)
                if x_xpos < 1 or x_xpos > 3 or y_ypos < 1 or y_ypos > 3:
                    api.do_send_player_msg(playername, "§a井字棋§f>> §c下子位置有误",cb=None)
                elif Game_JZQ.轮流() != typePlayer:
                    api.do_send_player_msg(playername, "§a井字棋§f>> §c没有轮到你落子.",cb=None)
                else:
                    result = Game_JZQ.落子(x_xpos, y_ypos, typePlayer)
                    if result:
                        api.do_send_player_msg(playername, "§a井字棋§f>> §a成功下子.",cb=None)
                        Game_JZQ.轮流(True)
                        api.do_send_player_msg(i[Game_JZQ.轮流()], "§a井字棋§f>> §6对方已落子: §e({x}, {y})§6 到你啦!".format(x = x_xpos, y = y_ypos),cb=None)
                        if Game_JZQ.判定():
                            Game_JZQ.stage_display(i, playername)
                            api.do_send_ws_cmd("/title %s title §e井字棋" % playername,cb=None)
                            api.do_send_ws_cmd("/title %s subtitle §a祝贺!你赢了!" % playername,cb=None)
                            
                            nexPlayer = i[Game_JZQ.轮流()]
                            Game_JZQ.stage_display(i, nexPlayer)
                            api.do_send_ws_cmd("/title %s title §e井字棋" % nexPlayer,cb=None)
                            api.do_send_ws_cmd("/title %s subtitle §7惜败.." % nexPlayer,cb=None)
                            JZQ_Rooms.remove(i)
                            Game_JZQ.重置(True)
                            continue
                        elif Game_JZQ.判死():
                            Game_JZQ.重置()
                    else:
                        api.do_send_player_msg(playername, "§a井字棋§f>> §c这个地方不能下子",cb=None)
    def testforIDText(packet:dict):
        if packet["Message"][0:3] == "井字棋" and packet["TextType"] in [1,7]:
            井字棋(packet["SourceName"],packet["Message"])
        elif packet["Message"][0:2] in ["下子","落子"] and packet["TextType"] in [1,7]:
            下子(packet["SourceName"],packet["Message"])
        ...
    def IDText_pakcet(packet):
        Thread(target=testforIDText,args=[packet]).start()
    api.listen_mc_packet(pkt_type="IDText",cb=None,on_new_packet_cb=IDText_pakcet)
    def rep_井字棋():
        if JZQ_Rooms:
            for i in JZQ_Rooms:
                for player in i:
                    for player in i:
                        Game_JZQ.stage_display(i, player)
                if Game_JZQ.Timer() == 0:
                    api.do_send_player_msg(player, "§a井字棋§f>> §c时间到!游戏结束",cb=None)
                    for player in i:
                        Game_JZQ.stage_display(i, player, True)
                    JZQ_Rooms.remove(i)
                    Game_JZQ.重置(True)
                    continue
                Game_JZQ.Timer(-1)
    api.execute_with_repeat(rep_井字棋,repeat_time=1)

omega.add_plugin(plugin=Mono_plugin_井字棋)
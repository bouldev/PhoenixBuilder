# 插件: 开
import time, threading, traceback
from omega_side.python3_omega_sync import API
from omega_side.python3_omega_sync import frame as omega
from omega_side.python3_omega_sync.bootstrap import install_lib
from omega_side.python3_omega_sync.protocol import *

# install_lib("numpy")
# import numpy as np
# <- SuperScript ->

class Super_AFKGobangBasic:
    """
    SuperGobang v SuperScript|SuperAFK
    TM AND LICENSED BY DOYEN STUDIO(1991-2023).Inc.
    """
    rooms = {}
    waitingCache = {}
    cacheUID = 0
    DESCRIPTION = __doc__
    __version__ = "0.0.1"
    class Room:
        def __init__(this, _1P: str, _2P: str):
            this.playerA = _1P
            this.playerB = _2P
            this.turns = 1
            this.timeleft = 120
            this.startTime = time.time()
            this.stage = SuperGobangStage()
            this.maxTimeout = 0
            this.status = ""

        def turn(this):
            this.turns = [0, 2, 1][this.turns]

        def isTurn(this, player: str):
            return (player == this.playerA and this.turns == 1) or (player == this.playerB and this.turns == 2)

        def resetTimer(this):
            this.startTime = time.time()
            this.timeleft = 120

        def fmtTimeLeft(this):
            time_min, time_sec = divmod(int(time.time()+this.timeleft-this.startTime), 60)
            return "%02d ： %02d" % (time_min, time_sec)

        def PID(this, player: str):
            return 1 if player == this.playerA else (2 if player == this.playerB else None)

        def anotherPlayer(this, player: str):
            "请不要在未确认玩家为该局玩家的时候使用该方法"
            return this.playerA if player == this.playerB else this.playerB

        def setStatus(this, status: str):
            this.status = status
    
    def createRoom(this, roomdata: Room):
        roomUID = hex(GobangRoom.cacheUID)
        GobangRoom.cacheUID += 1
        this.rooms[roomUID] = roomdata
        return roomUID

    def removeRoom(this, roomUID: str):
        del this.rooms[roomUID]

    def getRoom(this, player: str):
        for _k in this.rooms:
            if this.rooms[_k].playerA == player or this.rooms[_k].playerB == player:
                return _k
        return None

    def GameStart(Frame: API, _1P: str, _2P: str):
        Frame.do_send_player_msg(_1P, "§l§7> §a五子棋游戏已开始， 退出聊天栏查看棋盘，§f输入 xz <纵坐标> <横坐标>以下子.", ign_cb)
        Frame.do_send_player_msg(_2P, "§l§7> §a五子棋游戏已开始， 退出聊天栏查看棋盘，§f输入 xz <纵坐标> <横坐标>以下子.", ign_cb)
        Frame.do_set_player_title(_1P, "§e游戏开始", ign_cb)
        Frame.do_set_player_title(_2P, "§e游戏开始", ign_cb)
        Frame.do_set_player_subtitle(_1P, "§a聊天栏输入 下子 <纵坐标> <横坐标> 即可落子", ign_cb)
        Frame.do_set_player_subtitle(_2P, "§a聊天栏输入 下子 <纵坐标> <横坐标> 即可落子", ign_cb)
        linked_room_uid = GobangRoom.createRoom(Super_AFKGobangBasic.Room(_1P, _2P))
        this_room: Super_AFKGobangBasic.Room = GobangRoom.rooms[linked_room_uid]
        while 1:
            time.sleep(0.5)
            nowPlayer = _1P if this_room.isTurn(_1P) else _2P
            actbarText = f"§e§l五子棋 {this_room.fmtTimeLeft()} %s\n{this_room.stage.strfChess()}§9SuperGobang\n§a"
            Frame.do_set_player_actionbar(_1P, actbarText % ("§a我方下子" if this_room.isTurn(_1P) else "§6对方下子"), ign_cb)
            Frame.do_set_player_actionbar(_2P, actbarText % ("§a我方下子" if this_room.isTurn(_2P) else "§6对方下子"), ign_cb)
            if this_room.status == "done":
                Frame.do_send_player_msg("@a", "game done", ign_cb)
                break
            if this_room.timeleft < 20:
                Frame.do_set_player_title(nowPlayer, "§c还剩 20 秒", ign_cb)
                Frame.do_set_player_subtitle(nowPlayer, "§6若仍然没有下子， 将会跳过你的回合", ign_cb)
            if this_room.timeleft <= 0:
                Frame.do_set_player_title(nowPlayer, "§c已跳过你的回合", ign_cb)
                Frame.do_set_player_title(_1P if _1P != nowPlayer else _2P, "§6对方超时， 现在轮到你落子", ign_cb)
                this_room.resetTimer()
                this_room.maxTimeout += 1
                this_room.turn()
                if this_room.maxTimeout > 1:
                    Frame.do_set_player_title(_1P, "§c游戏超时， 本局已结束", ign_cb)
                    Frame.do_set_player_title(_2P, "§c游戏超时， 本局已结束", ign_cb)
                    break
            this_room.timeleft -= 1

    def GameWait(this, Frame: API, _1P: str, _2P: str):
        Frame.do_send_player_msg(_1P, "§7§l> §r§6正在等待对方同意请求..", ign_cb)
        Frame.do_send_player_msg(_2P, f"§7§l> §r§e{_1P}§f向你发送了五子棋对弈邀请 ！", ign_cb)
        Frame.do_send_player_msg(_2P, f"§7§l> §r§a输入wzq y同意， §c输入wzq n拒绝", ign_cb)
        waitStartTime = time.time()
        this.waitingCache[_2P] = None
        while 1:
            time.sleep(0.5)
            if time.time() - waitStartTime > 30:
                Frame.do_send_player_msg(_1P, f"§7§l> §c等待{_2P}的请求超时， 已取消.", ign_cb)
                break
            if this.waitingCache[_2P]:
                if this.waitingCache[_2P] == 1:
                    Super_AFKGobangBasic.GameStart(Frame, _1P, _2P)
                    Frame.do_echo("Game done!", ign_cb)
                    break
                else:
                    Frame.do_send_player_msg(_1P, f"§7§l> §c{_2P}拒绝了您的邀请..", ign_cb)
                del this.waitingCache[_2P]

class SuperGobangStage():
    def __init__(this):
        this.basic()

    def basic(this):
        this.SIZE = 12
        this.field = [[0 for _ in range(this.SIZE)] for _v in range(this.SIZE)]
        this.winner = None
        this.BLACK = 1
        this.WHITE = 2
        this.PosSignLeft = ["①", "②", "③", "④", "⑤", "⑥", "⑦", "⑧", "⑨", "⑩", "⑪", "⑫"]

    def centers(this, l, w ,f:API):
        if (l<3 or l>this.SIZE-2) and (w<3 or w>this.SIZE-2) or not this.getField(l, w):
            return False
        return any([
            (this.getField(l, w) == this.getField(l-1, w) == this.getField(l-2, w) == this.getField(l+1, w) == this.getField(l+2, w)),
            (this.getField(l, w) == this.getField(l, w-1) == this.getField(l, w-2) == this.getField(l, w+1) == this.getField(l, w+2)),
            (this.getField(l, w) == this.getField(l-1, w-1) == this.getField(l-2, w-2) == this.getField(l+1, w+1) == this.getField(l+2, w+2) != 0),
            (this.getField(l, w) == this.getField(l-1, w+1) == this.getField(l-2, w+2) == this.getField(l+1, w-1) == this.getField(l+2, w-2) != 0)
        ])

    def getField(this, l: int, w: int):
        if l not in range(1, this.SIZE + 1) or w not in range(1, this.SIZE + 1): return None
        return this.field[l-1][w-1]

    def setField(this, l, w, chesType):
        if l not in range(1, this.SIZE + 1) or w not in range(1, this.SIZE + 1): return False
        this.field[l-1][w-1] = chesType
        return True

    def get_win(this, api: API):
        for cl in range(1, this.SIZE + 1):
            for cw in range(1, this.SIZE + 1):
                if this.centers(cl, cw, api):
                    return this.getField(cl, cw)
        return None

    def onchess(this, l: int, w: int, player):
        assert not this.getField(l, w), "§c此处不可以再下子哦"
        if not l in range(1, this.SIZE + 1) or not w in range(1, this.SIZE + 1):
            return False
        this.setField(l, w, player)
        return True

    def toSignLeft(this, num: int):
        return this.PosSignLeft[num-1]

    def toSignRight(this, num: int):
        return this.PosSignRight[num-1]


    def strfChess(this):
        fmt: str = "§e   1 2 3 4 5 6 7 8 9 10 1112§r"
        for cl in this.field:
            fmt += "\n{}"
            for cw in cl:
                if cw == 0:
                    fmt += "§7§l▒§r"
                elif cw == 1:
                    fmt += "§0§l▒§r"
                elif cw == 2:
                    fmt += "§f§l▒§r"
        return fmt.format(*[this.toSignLeft(i) for i in range(1, this.SIZE + 1)])

GobangRoom = Super_AFKGobangBasic()

ign_cb = lambda *_, **_v: None

def plugin(Frame: API):
    def on_menu_invoked(chat: PlayerInput):
        if len(chat.Msg) == 1:
            _2P = chat.Msg[0]
            if len(_2P) < 2:
                Frame.do_send_player_msg(chat.Name, "§c模糊搜索玩家名， 输入的名字长度必须大于1", ign_cb)
                return
            allplayers = [single_player.name for single_player in Frame.do_get_players_list(None)]
            new2P = None
            for single_player in allplayers:
                if _2P in single_player:
                    new2P = single_player
                    break
            if not new2P:
                Frame.do_send_player_msg(chat.Name, f"§c未找到名字里含有\"{_2P}\"的玩家.", ign_cb)
                return
            if not GobangRoom.getRoom(chat.Name):
                threading.Thread(target=GobangRoom.GameWait, args = (Frame, chat.Name, new2P)).start()
        
    Frame.listen_omega_menu(triggers=["五子棋", "wzq"],argument_hint="[对手]",usage="开一局五子棋游戏",cb = ign_cb,on_menu_invoked=on_menu_invoked)
    
    def on_chess_cmd(packet: dict[str, str]):
        player, msg = packet["SourceName"], packet["Message"]
        if msg.startswith("下子") or msg.lower().startswith("xz") or msg.startswith("XZ"):
            if GobangRoom.rooms:
                in_room = GobangRoom.getRoom(player)
                if in_room:
                    inRoom: Super_AFKGobangBasic.Room = GobangRoom.rooms[in_room]
                    if inRoom.isTurn(player):
                        try:
                            try:
                                _, posl, posw = msg.split()
                            except:
                                raise AssertionError("§c落子格式不正确； 下子/xiazi/xz <纵坐标> <横坐标>")
                            assert inRoom.stage.onchess(int(posl), int(posw), inRoom.PID(player))
                            Frame.do_send_player_msg(player, "§l§7> §r§a成功下子.", ign_cb)
                            Frame.do_send_player_msg(inRoom.anotherPlayer(player), "§l§7> §r§a到你啦！", ign_cb)
                            inRoom.resetTimer()
                            is_win = inRoom.stage.get_win(Frame)
                            if is_win:
                                Frame.do_set_player_title(player, "§a§l恭喜！", ign_cb)
                                Frame.do_set_player_subtitle(player, "§e本局五子棋您获得了胜利！", ign_cb)
                                Frame.do_send_player_msg(player, "§7§l> §r§e恭喜！ §a本局五子棋您取得了胜利！", ign_cb)
                                Frame.do_send_wo_cmd(f"/execute {inRoom.anotherPlayer(player)} ~~~ playsound random.levelup @s", ign_cb)
                                Frame.do_set_player_title(inRoom.anotherPlayer(player), "§7§l遗憾惜败", ign_cb)
                                Frame.do_set_player_subtitle(inRoom.anotherPlayer(player), "§6下局再接再厉哦！", ign_cb)
                                Frame.do_send_wo_cmd(f"/execute {inRoom.anotherPlayer(player)} ~~~ playsound note.pling @s ~~~ 1 0.5")
                                inRoom.setStatus("done")
                            inRoom.turn()
                        except AssertionError as err:
                            Frame.do_send_player_msg(player, str(err), ign_cb)
                        except:
                            print(traceback.format_exc())
                    else:
                        Frame.do_send_player_msg(player, "§c还没有轮到你落子哦", ign_cb)
                else:
                    Frame.do_send_player_msg(player, "§c需要开启一场五子棋游戏才可以落子", ign_cb)
            else:
                Frame.do_send_player_msg(player, "§c需要开启一场五子棋游戏才可以落子", ign_cb)
        elif msg.lower() == "wzq y":
            if player in GobangRoom.waitingCache.keys():
                GobangRoom.waitingCache[player] = 1
        elif msg.lower() == "wzq n":
            if player in GobangRoom.waitingCache.keys():
                GobangRoom.waitingCache[player] = 2

    Frame.listen_mc_packet(pkt_type="IDText",cb=ign_cb,on_new_packet_cb=on_chess_cmd)

    def player_exit(player: PlayerInfo):
        player = player.name
        if GobangRoom.rooms:
            in_room = GobangRoom.getRoom(player)
            if in_room:
                inRoom: Super_AFKGobangBasic.Room = GobangRoom.rooms[in_room]
                Frame.do_set_player_title(inRoom.anotherPlayer(player), "§c对方已退出游戏，游戏结束", ign_cb)
                inRoom.setStatus("done")


    Frame.listen_player_logout(cb=None,on_player_logout_cb=player_exit)

omega.add_plugin(plugin=plugin)

omega.run(addr=ign_cb)
print("Gobang plugin load succcessully.")

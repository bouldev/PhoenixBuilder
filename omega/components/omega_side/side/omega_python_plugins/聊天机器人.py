# 插件: 关

from omega_side.python3_omega_sync import API
from omega_side.python3_omega_sync import frame as omega
from omega_side.python3_omega_sync.bootstrap import install_lib
from omega_side.python3_omega_sync.protocol import *
import json,requests

def plugin_main(api:API): 
    def on_menu_invoked(player_input:PlayerInput):
        player=player_input.Name
        try:
            msg = str(list(player_input.Msg)[0])
        except:
            api.do_send_player_msg(player,"§c缺少信息!",cb=None)
        try:
            bot = requests.get("http://api.qingyunke.com/api.php?key=free&appid=0&msg="+msg)
            bot = bot.text;bot = json.loads(bot)
            api.do_send_player_msg(player,"bot:"+bot["content"],cb=None)
        except:
            pass
    def plugin_menu():
        api.listen_omega_menu(triggers=["聊天机器人","bot"],argument_hint="",usage="omg bot 信息",cb=None,on_menu_invoked=on_menu_invoked)
    plugin_menu()
omega.add_plugin(plugin=plugin_main)
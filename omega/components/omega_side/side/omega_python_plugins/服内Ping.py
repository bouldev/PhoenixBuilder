# 插件: 关

from omega_side.python3_omega_sync import API
from omega_side.python3_omega_sync import frame as omega
from omega_side.python3_omega_sync.bootstrap import install_lib
from omega_side.python3_omega_sync.protocol import *
import json,requests

def plugin_main(api:API):
    def on_menu_invoked(player_input:PlayerInput):
        try:
            player = player_input.Name
            input_msg=api.do_get_get_player_next_param_input(player,hint="请输入要Ping的网站/IP",cb=None);input_msg = input_msg.input
            ping = requests.post("https://api.yum6.cn/ping.php",data={'host':input_msg[0]})
            ping = ping.text;ping = json.loads(ping)
            if ping["state"] == "1001":
               api.do_send_player_msg(player,"§cPING失败,请检查网站是否禁PING或无法访问",cb=None)
            else: 
                ping = "§a地址:"+ping["host"]+"\nIP:"+ping["ip"]+"\n位置:"+ping["location"]+"\n节点:"+ping["node"]+"\n平均时间:"+ping["ping_time_avg"]+"\n最大时间:"+ping["ping_time_max"]+"\n最小时间:"+ping["ping_time_min"]+"\n状态值:"+ping["state"]+"\n"
                api.do_send_player_msg(player,"=========PING成功=========\n"+ping+"§f==========================",cb=None)
        except:
            print("ping失败,将代理关闭即可")
    def plugin_menu():
        api.listen_omega_menu(triggers=["PING","ping"],argument_hint="",usage="ping 输入omg ping后将会向你询问地址",cb=None,on_menu_invoked=on_menu_invoked)
    plugin_menu()

omega.add_plugin(plugin=plugin_main)
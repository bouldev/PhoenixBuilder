# 插件: 开

import time
from omega_side.python3_omega_sync import API
from omega_side.python3_omega_sync import frame as omega
from omega_side.python3_omega_sync.bootstrap import install_lib
from omega_side.python3_omega_sync.protocol import *

def red_packet_plugin(api:API):
    response=api.do_echo("抢红包插件已经启用",cb=None)
    print(response.msg) 
    # 剩下的晚点再写吧

omega.add_plugin(plugin=red_packet_plugin)
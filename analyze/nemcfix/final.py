#!/usr/bin/env python
# ohhhhhhhhhhhhhhhhhhhhhh
# I'm so excited!!!
# Thank https://tool.lu/pyc/ for helping me completing the final leap!!!
# 2021.10.25 18:40 Ruphane.
time = time
setting = setting
rpc = common.network
('defaultrpc',)
uid = setting.get_login_uid()
#uid = 0
num = uid & 255 ^ (uid & 65280) >> 8
curTime = int(time.time())
num = curTime & 3 ^ (num & 7) << 2 ^ (curTime & 252) << 3 ^ (num & 248) << 8
logger.info('GetLoadingTime uid:%s curTime:%d num:%d', uid, curTime, num)
rpc.CServerRpc().SetloadLoadingTime(num)
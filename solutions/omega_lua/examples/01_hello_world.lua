local omega = require("omega")
local print = omega.system.print
local block_sleep = omega.system.block_sleep
print("hello world!")

print(("你现在位于 %s"):format(omega.system.cwd()))
print(("此设备的系统是 %s"):format(omega.system.os()))
print(("此代码启动的系统时间(unix time)为 %.2f"):format(omega.system.start_time))
print(("自代码启动以来，现在时间为 %.2fs"):format(omega.system.now()))
local nowTime = omega.system.now()
block_sleep(3.0) -- 3秒后退出
print(("自代码启动以来，现在时间为 %.2fs"):format(omega.system.now()))
print(("自sleep以来，经过了 %.2fs"):format(omega.system.now() - nowTime))
print("bye!")

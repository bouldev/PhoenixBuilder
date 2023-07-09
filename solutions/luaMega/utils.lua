local omega = require("omega")

-- backend function aliases
local print = omega.backend.print -- print(info:string)
print("Hello World!")             --example

-- system function aliases
local os = omega.system.os                   -- arch-system:string=os()
print(os())                                  --example

local cwd = omega.system.cwd                 -- cwd:string=cwd()
print(cwd())                                 --example

local now = omega.system.now                 -- second_since_start:number=now()
print(("now time: %.2fs"):format(now()))     --example

local set_timeout = omega.system.set_timeout -- set_timeout(seconds:number,fn:function)
local begin_time = omega.system.now()
set_timeout(1.0, function()
    print(("1s passed (actual %.2fs)"):format(now() - begin_time))
end) --example
set_timeout(2.0, function()
    print(("2s passed (actual %.2fs)"):format(now() - begin_time))
end)                                           --example
print(("shold be printed before 1s passed (time elasped: %.2fs)"):format(now() - begin_time))
local set_interval = omega.system.set_interval -- stop:function=set_interval(seconds:number,fn:function)
local stop_interval = set_interval(0.5, function()
    print(("0.5s passed (actual %.2fs)"):format(now() - begin_time))
end)                            --example
set_timeout(3.1, stop_interval) -- stop after 3s

-- block function aliases
-- 警告，这些block函数会阻塞整个lua文件，包括timeout\interval\任何回调
local block_read = omega.block.user_input -- input:string=block_read()
print(block_read())                       --example
-- 在用户输入前，你什么都做不了，所有的代码都会被阻塞,包括timeout\interval\任何回调

local block_sleep = omega.block.sleep -- block_sleep(seconds:number)
block_sleep(3.0)                      --example
-- 在这3s内，你什么都做不了，所有的代码都会被阻塞,包括timeout\interval\任何回调
print("sleeped 3s passed")            --example
-- 请移步 input.lua 查看如何在阻塞的同时，等待用户输入,
-- 请移步 packet.lua 查看如何在阻塞的同时，等待游戏包

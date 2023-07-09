local omega = require("omega")
local block_read = omega.block.user_input
local print = omega.backend.print
local now = omega.system.now

print("please input:")
print(block_read())

local make_mux_poller = omega.block.make_mux_poller
local block_sleep = omega.block.sleep

print("input in 0.1s")
local poller = make_mux_poller()
local event = poller:poll(block_read):poll(block_sleep, 0.1):block_get_next()
if event.type == block_read then
    print(("get input %s"):format(event.data))
elseif event.type == block_sleep then
    print("no input!")
end

print("input in 3s")
local start_time = now()
local poller = make_mux_poller()
local event = poller:poll(block_read):poll(block_sleep, 3):block_get_next()
if event.type == block_read then
    print(("get input %s in %.2fs"):format(event.data, now() - start_time))
elseif event.type == block_sleep then
    print("no input!")
end

local read_input_with_timeout = function(time_out)
    print(("input in %.2fs"):format(time_out))
    local start_time = now()
    local poller = make_mux_poller()
    local event = poller:poll(block_read):poll(block_sleep, time_out):block_get_next()
    if event.type == block_read then
        print(("get input %s in %.2fs"):format(event.data, now() - start_time))
        return event.data
    elseif event.type == block_sleep then
        print("no input!")
        return
    end
end

local input0s = read_input_with_timeout(0.1)

local input3s = read_input_with_timeout(3)

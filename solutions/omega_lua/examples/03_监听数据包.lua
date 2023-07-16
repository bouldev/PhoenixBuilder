local omega = require("omega")
local packets = omega.packets

-- -- 监听数据包需要使用 packet_poller，创建一个 packet_poller 的方法是 omega.listen.new_packet_poller(...)
local new_packet_poller = omega.listen.new_packet_poller
-- -- 这里有两种方式可以创建 packet_poller

-- -- 1. new_packet_poller(数据包类型1, 数据包类型2, 数据包类型3 ...)
-- -- 即列举出所有你希望监听的数据包类型
local packet_poller = new_packet_poller(packets.Text, packets.CommandOutput)

-- -- 2. new_packet_poller(packets.all, no数据包类型1, no数据包类型2, no数据包类型3 ...)
-- -- 即列举出所有你希望忽略的数据包类型, packets.all 表示所有数据包类型，no数据包类型1 表示忽略数据包类型1
-- -- 这种方式的好处是，当你希望监听的数据包类型很多的时候，你只需要列举出你不希望监听的数据包类型即可
-- -- 例如
-- -- local packet_poller = new_packet_poller(packets.all, packets.noMovePlayer)

-- -- 如果你不清楚有哪些数据包类型，可以使用 packets.all_names 来查看所有数据包类型
-- print(packets.all_names)

-- -- 和之前的 block_sleep, block_input 类似，packet_poller 的许多函数(方法)也是阻塞的，也就是说，当我们调用这个函数的时候，程序会停下来，等待这个函数返回
local packet_count = 0
while packet_poller:block_has_next() do                                  -- 如果有下一个数据包
    local packet = packet_poller:block_get_next()                        -- 获取下一个数据包
    print(("packet name: %s id: %s"):format(packet:name(), packet:id())) -- 打印数据包的名称和id
    -- just 10 packets as an example
    packet_count = packet_count + 1
    if packet_count > 10 and packet:name() == packets.CommandOutput then                           -- 我们只取10个数据包，然后退出
        print(("detail packet %s"):format(packet:json_str(packet)))                                -- json_str 是一个比较慢的函数，但是它可以将数据包转换成json字符串，方便我们查看，它最好只在调试的时候使用
        local packet_userdata = packet:user_data()                                                 -- user_data 可以将数据包转换成lua user_data，方便我们取其中的数据
        print(("detail packet (user_data) %s"):format(packet_userdata))                            -- lua user_data 可以直接打印, 但是诸如 ipair\pair 之类的函数对其不起作用
        print(("detail packet (lua table) %s"):format(ud2lua(packet_userdata)))                    -- 使用 ud2lua 可以将其彻底转为 lua table，ipair和pair可以使用，但是转换需要额外消耗时间，如果不需要用到 pair 和 ipair，建议不要使用
        print(("Origin: %s"):format(packet_userdata.CommandOrigin.Origin))                         -- lua user_data 可以直接取值
        print(("OutputMessages[0].Message: %s"):format(packet_userdata.OutputMessages[1].Message)) -- lua 的索引从1开始，而不是通常的从0开始，请小心
        packet_poller:stop()
    end
end
print("20 packets listened")

-- 用和之前类似的方法，我们可以用mux_poller来监听包括数据包在内的多个事件
-- 如下，我们定义一个多种事件混合的多路复用器
local mux_poller = omega.listen.new_mux_poller()
local packet_poller = nil

-- 第一个阻塞监听事件 event_after，它会在2秒后发生, 并且携带一个参数 action = "start" 表示开始监听数据包
mux_poller:poll(mux_poller.event_after, 2.0, { action = "start" })
-- 第二个阻塞监听事件 event_after，它会在20秒后发生, 并且携带一个参数 action = "stop" 表示停止监听数据包
mux_poller:poll(mux_poller.event_after, 20.0, { action = "stop", reason = "timeout" })
while mux_poller:block_has_next() do                                                             -- 如果有下一个事件
    local event = mux_poller:block_get_next()                                                    -- 获取下一个事件
    if event.type == mux_poller.event_after then                                                 -- 如果是来自 event_after 的事件
        print(event.data.action)                                                                 -- 打印事件的参数 action
        if event.data.action == "start" then                                                     -- 如果参数 action 为 "start"
            print("start packet poller")                                                         -- 开始监听数据包
            packet_poller = new_packet_poller(packets.all, packets.noMovePlayer)                 -- 创建一个 packet_poller，监听所有数据包，但是忽略 noMovePlayer 数据包
            mux_poller:poll(packet_poller)                                                       -- 在复用器中监听数据包，这样就可以同时监听多个事件了
            mux_poller:poll(omega.system.block_input)                                            -- 在复用器中监听用户输入
        elseif event.data.action == "stop" then                                                  -- 如果参数 action 为 "stop"
            print(("stop packet poller because %s"):format(event.data.reason))
            packet_poller:stop()                                                                 -- 停止监听数据包
            mux_poller:stop()                                                                    -- 停止复用器
        end
    elseif event.type == packet_poller then                                                      -- 如果是来自 packet_poller 的事件（数据包）
        local packet = event.data                                                                -- 获取数据包
        print(("* get packet name: %s id: %s"):format(packet:name(), packet:id()))               -- 打印数据包的名称和id
    elseif event.type == omega.system.block_input then                                           -- 如果是来自 block_input 的事件（用户输入）
        print(("user input: %s"):format(event.data))                                             -- 打印用户输入
        mux_poller:poll(mux_poller.event_after, 0.5, { action = "stop", reason = "user_input" }) -- 0.5秒后停止监听数据包
    end
end
print("bye")

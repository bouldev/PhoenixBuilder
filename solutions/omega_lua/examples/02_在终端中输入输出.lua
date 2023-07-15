local omega = require("omega")
local print = omega.system.print
local block_input = omega.system.block_input

print("请随便输入点什么")
local something = block_input()
print(("你刚才输入了 %s"):format(something))

--  现在的问题是，假如我们希望设置一个超时，比如说，如果用户在3秒内没有输入，那么就默认输入了"nothing"。
-- 我们知道，我们可以使用sleep函数来等待一段时间，但是我们不知道怎么在等待的同时，又能够接收用户的输入。
-- 因为，如果我们先使用sleep函数，那么在等待的时候，我们就无法接收用户的输入了。
-- 或者，我们可以先接收用户的输入，但是在等待的时候，我们就无法接收用户的输入了。
-- 就像这样：

omega.system.block_sleep(3.0)
print("3s passed")
print("请随便输入点什么")
local something = block_input()
print(("你刚才输入了 %s"):format(something))

-- 这是因为，这两个函数是阻塞的，也就是说，当我们调用这两个函数的时候，程序会停下来，等待这两个函数返回。
-- 在lua支持中，阻塞函数都被命名为 block_xxx，比如block_sleep，block_input

-- 回到我们的问题，我们希望在等待的时候，又能够接收用户的输入，该怎么做呢？
-- 我们可以使用 mux_poller 来实现这个功能
-- mux_poller 是一个多路复用器，它可以同时监听阻塞多个事件，当其中一个事件发生的时候，它就会返回这个事件。
-- 创建一个 mux_poller 的方法是 omega.listen.new_mux_poller()
local mux_poller = omega.listen.new_mux_poller()
-- 然后，我们可以使用它来监听阻塞事件，比如说，我们可以监听阻塞输入事件 block_input
local event = mux_poller:poll(block_input):poll(omega.system.block_sleep, 0.5):block_get_next()
-- 上面这句代码等效于下面这些代码
-- mux_poller:poll(block_input) -- 监听阻塞输入事件
-- mux_poller:poll(omega.system.block_sleep, 3.0) -- 监听阻塞3秒事件
-- local event = mux_poller:block_get_next() -- 等待第一件发生的事情，然后返回这件事情
-- 如果我们监听的事件都没有发生，那么这个函数就会一直等待下去，直到有一个事件发生。

print("请随便输入点什么")
if event.type == block_input then                  -- 如果是阻塞输入事件
    print(("你输入了: %s"):format(event.data))
elseif event.type == omega.system.block_sleep then -- 如果是阻塞3秒事件
    print("你没有输入")
end
mux_poller:stop() -- 需要注意的时，因为我们只取了一个事件，所以我们调用 stop 表示剩下的事件都不要了。如果不这么干，这个程序就永远不会结束


-- 我们将上面的代码封装成一个函数，这样就可以在任何地方使用了
local function get_input_with_timeout(time_out)
    local mux_poller = omega.listen.new_mux_poller()
    local event = mux_poller:poll(block_input):poll(omega.system.block_sleep, time_out):block_get_next()
    mux_poller:stop()
    if event.type == block_input then
        return event.data
    elseif event.type == omega.system.block_sleep then
        return nil
    end
end

print("请随便输入点什么")
print(("你刚才输入了 %s"):format(get_input_with_timeout(3.0)))


-- 相比 sleep, 还有一个更好的方法， 就是使用 event_after, 相比 sleep, event_after 可以携带一个参数描述事件的类型
local mux_poller = omega.listen.new_mux_poller()
local event = mux_poller:poll(block_input):event_after(0.5, { reason = "timeout" }):block_get_next()
-- 或者这样写， 上下两种写法等价，但是下面写法风格更统一
-- local event = mux_poller:poll(block_input):poll(mux_poller.event_after, 0.5, { reason = "timeout" }):block_get_next()
-- 或者这样写，event_after 携带的参数可以是任何类型，不一定是 table ()
local event = mux_poller:poll(block_input):poll(mux_poller.event_after, 0.5, "timeout"):block_get_next()
if event.type == block_input then
    print(("你输入了: %s"):format(event.data))
elseif event.type == mux_poller.event_after then
    print(("你没有输入: %s %s"):format(event.data, event.data.reason))
end
mux_poller:stop()

-- 现在， 你可以把 mux_poller:stop() 注释掉，然后看看会发生什么。

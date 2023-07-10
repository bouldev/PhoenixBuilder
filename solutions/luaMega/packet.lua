local omega = require("omega")
local print = omega.backend.print
local packets = omega.packets

local make_mux_poller = omega.block.make_mux_poller
local mux_poller = make_mux_poller()
-- mux_poller:poll(mux_poller.event_after, 2.0, "stop")

print(("%s"):format(packets.all_names))

local make_packet_poller = omega.block.make_packet_poller
local packet_poller = make_packet_poller(packets.Text, packets.CommandOutput) -- list of required packets
-- local poller = make_packet_poller(packets.all, packets.noMovePlayer, packets.noMoveActorAbsolute) -- or all packets except ...
local packet_count = 0
while packet_poller:has_next() do
    local packet = packet_poller:block_get_next()
    print(("packet name: %s id: %s"):format(packet:name(), packet:id()))
    -- just 20 packets as an example
    packet_count = packet_count + 1
    if packet_count > 20 and packet:name() == packets.CommandOutput then
        print(("detail packet %s"):format(packets.to_json_string_slow(packet)))
        local lua_table_packet = packets.to_lua_table(packet)
        print(("detail packet (lua table) %s"):format(lua_table_packet))
        print(("Origin: %s"):format(lua_table_packet.CommandOrigin.Origin))
        print(("OutputMessages[0].Message: %s"):format(lua_table_packet.OutputMessages[1].Message)) -- index of lua table starts from 1, be careful
        packet_poller:stop()
    end
end
print("20 packets listened")

local make_mux_poller = omega.block.make_mux_poller


local packet_poller = nil --= make_packet_poller(packets.all, packets.noMovePlayer, packets.noMoveActorAbsolute)
-- note that packets will queue up if you don't poll them, after 128 packets queued, the packet_poller cannot hold more packets


local function example_listen(event, mux_poller)
    if event.type == packet_poller then
        local packet = event.data
        print(("get packet name: %s id: %s"):format(packet:name(), packet:id()))
    elseif event.type == omega.block.sleep then
        print("awake after 3s, poller begin to work!")
        packet_poller = make_packet_poller(packets.all, packets.noMovePlayer) -- list of required packets
        mux_poller:poll(packet_poller)                                        -- begin to poll packet
        mux_poller:poll(omega.block.user_input)                               -- begin to poll user input
    elseif event.type == omega.block.user_input then
        print(("user input: %s"):format(event.data))                          -- print user input
        mux_poller:poll(mux_poller.event_after, 3.5,                          -- work like flag
            { action = "stop", reason = "user_input" })                       -- make a single which occours after 3.5s
        -- or (work like fuction)
        -- mux_poller:event_after(3.5, { action = "stop", reason = "user_input" }) -- make a single which occours after 3.5s
    elseif event.type == mux_poller.event_after and event.data.action == "stop" then
        print(("stop packet poller because %s"):format(event.data.reason))
        packet_poller:stop() -- stop packet poller
        -- 由于所有的源都停止了，所以这个mux_poller也停止了
    end
end


print("sleep 3s")
-- first sleep 3s
local mux_poller = make_mux_poller()
mux_poller:poll(omega.block.sleep, 3)
while mux_poller:has_next() do
    local event = mux_poller:block_get_next()
    example_listen(event, mux_poller)
end
print("packets listened")

print("start after sleep 3s")
local mux_poller = make_mux_poller()
mux_poller:poll(omega.block.sleep, 3)
mux_poller:as_async(function(event) example_listen(event, mux_poller) end)
print("no block in async mode")

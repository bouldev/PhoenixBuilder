# 插件: 关
# 第一行应该指明自己是一个 标准 Omega 插件， "插件: 开" 代表使用这个插件， "插件: 关" 代表不使用这个插件
# 如果第一行什么都不指明，代表这个文件不是一个插件（比如说是文件或者别的文件什么的）
# 这个插件是用来展示所有可用的 api 的

import time
from omega_side.python3_omega_sync import API
from omega_side.python3_omega_sync import frame as omega
from omega_side.python3_omega_sync.bootstrap import install_lib
from omega_side.python3_omega_sync.protocol import *

# 如果你 不用安装额外的库，下面一段话 可以不看
# 如果 你需要安装一些库， 例如  websocket 库，你可以
# install_lib(lib_name="websocket",lib_install_name="websocket-client")
# 其中，lib_name 为 import时的名字,lib_install_name 为 pip install 时的名字
# 一般来说,lib_name 和 lib_install_name 是一样的
# 安装完之后就可以
# import websocket 
 
# 一个示例插件 api 为全部api入口
def plugin(api:API):
    # 发送一条消息到 omega 并接收，可以用来测试连接
    response=api.do_echo("hello",cb=None)
    print(response.msg) # hello
    
    # 执行命令： 框架返回命令执行结果
    response=api.do_send_ws_cmd("/tp @s ~~~",cb=None)
    print(response.result.CommandOrigin) #{'Origin': 5, 'UUID': '7da4befc-08e7-11ed-bcb1-52540039f5d6', 'RequestID': '96045347-a6a3-4114-94c0-1bc4cc561694', 'PlayerUniqueID': 0}
    print(response.result.OutputType)# 4
    print(response.result.SuccessCount) # 1
    print(response.result.OutputMessages) # [{'Success': True, 'Message': 'commands.tp.success.coordinates', 'Parameters': ['OmeGoTest', '100020.73', '320.00', '100364.04']}]
    
    # 以玩家身份执行命令： 框架返回命令执行结果（如果租赁服的 sendcommandfeedback为false，则omega会短暂的将其变为true
    response=api.do_send_player_cmd("/tp @s ~~~",cb=None)
    print(response.result.CommandOrigin) #{'Origin': 5, 'UUID': '7da4befc-08e7-11ed-bcb1-52540039f5d6', 'RequestID': '96045347-a6a3-4114-94c0-1bc4cc561694', 'PlayerUniqueID': 0}
    print(response.result.OutputType)# 4
    print(response.result.SuccessCount) # 1
    print(response.result.OutputMessages) # [{'Success': True, 'Message': 'commands.tp.success.coordinates', 'Parameters': ['OmeGoTest', '100020.73', '320.00', '100364.04']}]
    
    #发送 setting 命令：与前两个不同的是，这里的cmd 虽然也能是 "setblock ..." 但是对于诸如 "tp ..." 等指令并不能有效的执行且这个指令没有返回值，因此，omega框架仅仅会简单的返回一个 ack=True (ack 意为 acknowledge)
    response=api.do_send_wo_cmd("setblock ~~~ air",cb=None)
    print(response.ack) # True
    
    # 从omega框架那里获得一个有非常多的和玩家状态相关信息的数据体
    response=api.do_get_uqholder(cb=None)
    #print(response) # 这里有非常多的和玩家状态相关信息，自己看吧
    
    # 不是简单的玩家列表，除了名字外还有三个id，其中runtimeID只有在机器人看到玩家一次时才能获得
    response=api.do_get_players_list(cb=None)
    for player in response:
        print(player.name) #OmeGoTest
        print(player.runtimeID) #0
        print(player.uuid) #00000000-0000-4000-8000-0000392af26c
        print(player.uniqueID) #-214748364274

    # 刷新并读取omega的所有已知积分板，包括的积分板可能不止在线玩家
    # 不要太过频繁的调用它
    response=api.do_get_scoreboard(cb=None)
    print(response) # {'coin': {'OmeGoTest': 100}, 'time': {'OmeGoTest': 0}}

    # 菜单被唤起时会调用这个函数
    def on_menu_invoked(player_input:PlayerInput):
        print("菜单‘测试’被触发了")
        print(player_input.Name) # 2401PT
        print(player_input.Msg) #['啊吧啊吧', 'abab']
        print(player_input.Type) #1 
        print(player_input.FrameWorkTriggered) #True
        print(player_input.Aux) #{'TextType': 1, 'NeedsTranslation': False, 'SourceName': '2401PT', 'Message': 'omg 测试 啊吧啊吧 abab', 'Parameters': None, 'XUID': '', 'PlatformChatID': '', 'PlayerRuntimeID': ''}
        
        player=player_input.Name
        
        # 获取玩家输入
        param_input=api.do_get_get_player_next_param_input(player,hint="请随便说点什么",cb=None)
        print(param_input.success) #True
        print(param_input.player) #2401PT
        print(param_input.input) #['啊吧吧']
        print(param_input.err) #None
        
        # 在玩家聊天栏里显示一条消息
        response=api.do_send_player_msg(player,"hello",cb=None)
        print(response.ack) # True
        
        # 向玩家显示标题
        response=api.do_set_player_title(player,"标题信息",cb=None)
        print(response.ack) # True
        
        # 向玩家显示副标题
        response=api.do_set_player_subtitle(player,"小标题信息",cb=None)
        print(response.ack) # True
        
        # 设置玩家 actionbar
        response=api.do_set_player_actionbar(player,"actionbar",cb=None)
        print(response.ack) # True
        
        # 获得玩家坐标，可以加一些限制
        response=api.do_get_player_pos(player,limit="@p[name=[player]]",cb=None)
        print(response.success) # True
        print(response.pos) # [1, 117, -2]
        
        # 设置一条玩家相关的数据，omega将代为保存且在所有插件中可用
        response=api.do_set_player_data(player,entry="头衔",data="开发者",cb=None)
        print(response.ack) # True
        
        # 获取一条玩家相关的数据
        response=api.do_get_player_data(player,entry="头衔",cb=None)
        print(response.found) # True
        print(response.data) # 开发者
        
    # 创建一个菜单项，会自动整合到omega的菜单里
    response=api.listen_omega_menu(triggers=["测试","ceshi"],argument_hint="[参数A]",usage="打开测试功能",cb=None,on_menu_invoked=on_menu_invoked)
    print(response.sub_id) #['测试', 'ceshi'][参数A]打开测试功能
    
    def on_text_packet(packet):
        print("接受到了聊天数据包") # 每种包内容都不一样，具体内容自己看
        print("数据包内容为: ",packet) #{'TextType': 1, 'NeedsTranslation': False, 'SourceName': '2401PT', 'Message': '啊吧', 'Parameters': None, 'XUID': '', 'PlatformChatID': '', 'PlayerRuntimeID': ''}
    
    # 订阅某类数据包，比如这个就是聊天的数据包
    response=api.listen_mc_packet(pkt_type="IDText",cb=None,on_new_packet_cb=on_text_packet) # 有哪些数据包请查看开发文档
    print(response.succ) # True
    print(response.err) # None
    
    def on_mc_packet(packet):
        print("接受到了MC数据包") # 每种包内容都不一样，具体内容自己看
        print("数据包类型为:",packet.id) #IDPlayerHotBar
        print("数据包内容为: ",packet) #{'client': 0, 'type': 'mcPkt', 'sub': 'IDPlayerHotBar', 'data': {'SelectedHotBarSlot': 0, 'WindowID': 0, 'SelectHotBarSlot': False}, 'id': 'IDPlayerHotBar'}
    
    # 订阅所有数据包，数据包的类型可以在上面的函数里用 packet.id 看到
    # response=api.listen_any_mc_packet(cb=None,on_new_packet_cb=on_mc_packet) # 会收到很多很多数据，请小心的解除注释！
    print(response.succ) # True
    print(response.err) # None
    
    def on_player_login(player:PlayerInfo):
        print("新玩家登陆了")
        print(player.name) # 2401PT
        print(player.runtimeID) # 0 
        print(player.uniqueID) # -4294967295
        print(player.uuid) #00000000-0000-4000-8000-0000f04c2fec

    response=api.listen_player_login(cb=None,on_player_login_cb=on_player_login)
    print(response.ack) #True

    def on_player_logout(player:PlayerInfo):
        print("玩家退出了")
        print(player.name) # 2401PT
        print(player.runtimeID) # 0 
        print(player.uniqueID) # -4294967295
        print(player.uuid) #00000000-0000-4000-8000-0000f04c2fec

    response=api.listen_player_logout(cb=None,on_player_logout_cb=on_player_logout)
    print(response.ack) #True

    def on_block_update(update:BlockUpdateInfo):
        print("方块更新了") # air -> setblock ~ 150 ~ stone 1
        print(update.pos) # [1205, 150, 1068]
        print(update.origin_block_runtime_id) # 6690
        print(update.origin_block_simple_define) # {'Name': 'air', 'Val': 0}
        print(update.origin_block_full_define) # {'found': True, 'name': 'minecraft:air', 'props': {}}
        print(update.new_block_runtime_id) # 654
        print(update.new_block_simple_define) # {'Name': 'stone', 'Val': 1}
        print(update.new_block_full_define) # {'found': True, 'name': 'minecraft:stone', 'props': {'stone_type': 'granite'}}
        
    # response=api.listen_block_update(cb=None,on_block_update=on_block_update)   // 太频繁了，自己解除注释看吧
    # print(response.ack) #True
    
    # 获得手持物品的 runtime id 到描述的映射
    response=api.do_get_item_mapping(cb=None)
    for runtime_id,item_describe in response.mapping.items():
        print(runtime_id) # -10
        print(item_describe.name) # minecraft:stripped_oak_log
        print(item_describe.maxDamage) #0
        break # 太长了，显示一个演示一下就得了
    
    # 获得所有方块的 runtime id 到描述的映射 (调用这个函数会卡相当久，请务必注意！)
    # 这个函数真的很慢！
    response=api.do_get_block_mapping(cb=None)
    # 获得的信息包括到 标准 props 形式的描述
    for runtime_id,block_props_describe in enumerate(response.blocks):
        print(runtime_id," ",block_props_describe) #0 {'Name': 'minecraft:blue_candle', 'Properties': {'candles': 0, 'lit': 0}, 'Version': 17959425}
        break # 太长了，显示一个演示一下就得了
    for runtime_id,block_simple_describe in enumerate(response.simple_blocks):
        print(runtime_id," ",block_simple_describe) #0   {'Name': 'blue_candle', 'Val': 0}
        break # 太长了，显示一个演示一下就得了
    for java_block_describe,runtime_id in response.java_blocks.items():
        print(runtime_id," ",java_block_describe) #7231   minecraft:acacia_button[face=ceiling,facing=east,powered=false]
        break # 太长了，显示一个演示一下就得了
    
    def delay_exec_func(msg1,msg2):
        print("函数被延迟执行了，延迟了3秒， "+msg1+" "+msg2) # 3秒后 函数被延迟执行了， 啊吧 啊吧吧
    # 在独立的线程延迟执行函数，这里延迟 3 秒，"啊吧","啊吧吧" 对应参数 msg1 msg2
    # 非常不建议用 sleep，特别是在回调中，因为用不好可能会卡住所有的插件，只有在单独的线程里才能sleep
    api.execute_after(delay_exec_func,"啊吧","啊吧吧",delay_time=3)
    
    def func_in_individual_thread(msg1):
        time.sleep(3)
        print("这个函数在独立的线程被执行了, ",msg1)
    # 在独立的线程中执行，"啊吧" 对应参数 msg1
    api.execute_in_individual_thread(func_in_individual_thread,"啊吧")
    
    def repeat_exec_func(msg1,msg2,msg3):
        print("这个函数又被执行了，"+msg1+" "+msg2+" "+msg3)
    # 在独立的线程中循环执行
    api.execute_with_repeat(repeat_exec_func,"啊吧","啊吧吧","啊吧吧吧",repeat_time=5)
    
    # 向群服互通的群发送一条消息
    response=api.do_send_qq_msg(msg=f"hello from omega python plugin",cb=None)
    print(response.ack) #True
    
    # 执行 fb 命令，等同于在终端输入
    response=api.do_send_fb_cmd("set 0 0 0",cb=None)
    print(response.ack) # True

omega.add_plugin(plugin=plugin)

def another_plugin(api:API):
    response=api.do_echo("hello from another_plugin",cb=None)
    print(response.msg) # hello from another_plugin

omega.add_plugin(plugin=another_plugin)

class ClassPlugin(object):
    def __init__(self,name:str="") -> None:
        self.name=name
    
    def some_func(self):
        pass   
    
    def __call__(self,api:API):
        response=api.do_echo("hello from "+self.name,cb=None)
        print(response.msg) # hello from 类形式插件

class_plugin=ClassPlugin(name="类形式插件")
omega.add_plugin(plugin=class_plugin)

omega.run(addr=None) # 如果没有明确的指定 webscoket 端口，将会使用默认端口 "ws://localhost:24011/omega_side" 
print("主进程退出")

#  如果你想单独调试这个插件的话，你可以将这个文件移动到上一层目录（side目录下），然后运行
# 为什么要移动位置？因为真正运行时，这个文件中的代码会被整合到 side 目录下运行

# 如果不想这么麻烦，且你知道下面几句话在讲什么，那么你可以：
# 将工作目录/cwd 设为 上一层的 side 目录，并执行这份文件
# 或者在side 目录底下： python omega_python_plugins/演示全部api的插件.py
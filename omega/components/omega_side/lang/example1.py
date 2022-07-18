import asyncio
from python3_omega_side.core import MainFrame
from python3_omega_side import bootstrap,utils
from python3_omega_side import BasicPlugin

'''如果你需要自带库以外的第三方库，你应该在这里安装'''
# 注意，安装完成前，你必须假设用户设备中没有这个库
# 这里以 websockets 库作为一个例子
utils.install_lib(lib_name="websockets",lib_install_name="websockets")
'''你必须在安装完后才能 import 需要的第三方库'''
import websockets

'''初始化框架'''
# 当 addr 为None 时，由传入的第一个参数决定 omega side 地址（也就是由omega自动管理时）
# 但是，若你需要独立于omega调试插件的话，你完全可以自己设置 addr，默认应该为 addr="ws://localhost:24011/omega_side"
frame=bootstrap(addr=None)

'''插件定义(函数形式):'''
# 一个插件的入口可以仅仅是一个简单的函数
async def example_plugin1_chat_listener(frame:MainFrame):
    # 对于一个函数式的插件，api 可以通过 frame.xxx 的方式获得
    # 你可以不只一次的在这里调用 frame.xxx，总的来说，没有什么限制
    
    async def on_chat_msg(msg):
        utils.print(f"收到聊天数据包 {msg}")
    # Text 数据包的 ID 为 "IDText"， 若希望收到其他数据包，可以相应的修改值
    success=await frame.reg_mc_pkt_by_type("IDText",on_push_cb=on_chat_msg,wait_result=True) 
    # 如果希望收到所有的数据包，则可以用以下函数
    # async def on_any_msg(msg):
    #     utils.print(f"收到 MC 数据包 {msg}",flush=True)
    # success=await frame.reg_any_mc_pkt(on_push_cb=on_any_msg,wait_result=True)  
    if success[0]: 
        utils.print("已经成功订阅数据包")
    else:
        utils.print("数据包订阅失败, 原因"+success[1])
        
    # 以上为同步接受结果的方式，对于所有形为 frame.xxx 的api，全部提供一种更高性能的异步方式
    # 请注意，所有异步api中的 wait_result=False，或忽略，且 cb 应该为 None 或 async def
    # 例1
    # async def on_func_result(result):
    #     print(f"获得api返回: {result}")
    # await frame.reg_mc_pkt_by_type("IDText",cb=on_func_result,on_push_cb=on_chat_msg) 
    # 例2: cb=None 忽略结果
    # await frame.reg_mc_pkt_by_type("IDText",cb=None,on_push_cb=on_chat_msg) 
    # 例3: 最简洁的形式 （连cb一起忽略）
    # await frame.reg_mc_pkt_by_type("IDText",on_push_cb=on_chat_msg) 
    
# 将定义好的插件注入:
frame.add_plugin(example_plugin1_chat_listener)
# 这样也行: frame.add_plugin(example_plugin1_chat_listener(frame))
# 如果你迫不及待了，可以直接开始运行了(取消注释下面的 frame.start())
# 但是注意！ 这个函数只能被调用一次，且应该在所有插件注入完成之后调用
# frame.start()

'''插件定义(类形式):'''
class ExamplePlugin2_Schedule(BasicPlugin):
    '''
    每隔一段时间执行一次任务
    注意，以类为插件形式的时候，需要继承 BasicPlugin
    此时 api 可以通过 self.frame 访问
    2401 只推荐这种插件形式
    '''
    def __init__(self,name,period,cmds) -> None:
        super().__init__()
        self.name=name
        # 每隔 period 秒，执行一次 cmds
        self.period=period
        self.cmds=cmds
    
    async def on_cmd_result(self,cmd,result):
        utils.print(f"指令{cmd}的结果是{result}")
    
    async def on_cmd_result_no_name(self,result):
        utils.print(f"指令的结果是{result}") 
    
    async def run_cmds(self):
        for cmd in self.cmds:
            await self.frame.send_ws_cmd(cmd=cmd,cb=self.on_cmd_result_no_name)
            
            #这样也行:
            # async def on_result(result):
            #     await self.on_cmd_result(cmd=cmd,result=result)
            # await self.frame.send_ws_cmd(cmd=cmd,cb=on_result)
            
            # 这样也行:
            # result=await self.frame.send_ws_cmd(cmd=cmd,wait_result=True)
            # utils.print(result)
        
    async def __call__(self):
        while True:
            await utils.sleep(self.period)
            # 严禁使用 time.sleep
            utils.print(f"插件:{self.name} 执行计划任务")
            await self.run_cmds()
    
        

# 将定义好的插件注入:
frame.add_plugin(ExamplePlugin2_Schedule(name="示例插件A",period=3,cmds=["/list","/timeset day"]))
# 如果你迫不及待了，可以直接开始运行了(取消注释下面的 frame.start())
# frame.start()

class ExamplePlugin3_Echo(object):
    '''如果你不喜欢继承，你也可以这么写，总之只要保证能访问api就行'''
    def __init__(self,name) -> None:
        super().__init__()
        self.name=name
    
    async def __call__(self,frame):
        # 看起来和函数也差不多是吧
        resp=await frame.echo(msg="hello from "+self.name,wait_result=True)
        utils.print(resp)

    async def any_name(self,frame):
        # 如果你不喜欢 __call__
        resp=await frame.echo(msg="hello from "+self.name,wait_result=True)
        utils.print(resp)

# 如果你不喜欢继承，你也可以这么注入一个普通的类
# 可以这样:
# frame.add_plugin(ExamplePlugin3_Echo(name="示例插件B")(frame=frame))
# 这样也ok：
# frame.add_plugin(ExamplePlugin3_Echo(name="示例插件C").any_name(frame=frame))

''' 千万别忘了启动！'''
frame.start()

'''总结'''
# 首先，安装你需要的第三方库（必须假设用户设备上没有）
    # 例： utils.install_lib(lib_name="websockets",lib_install_name="websockets")
    # 你必须在安装完后才能 import 需要的第三方库
    # import websockets

# 初始化omega side 框架
    # frame=bootstrap(addr=None)
    # 如果你需要脱离omega调试或者开发，需要手动启动python程序，你可以：
    # frame=bootstrap(addr="ws://localhost:24011/omega_side")

# 定义插件，这些形式都是允许的：
    # 插件定义(函数形式): 除非你的插件真的很短，不然还是不要这么写
    # async def plugin(frame:MainFrame):
    #     async def cb(msg:str):
    #         utils.print(msg)
    #     await frame.echo("hello",cb=cb)
    # # 注入方式:
    # frame.add_plugin(plugin)
    # # 或:
    # frame.add_plugin(plugin(frame))

    # 插件定义(类形式): 2401 推荐这种插件形式
    # class ExamplePlugin2_Schedule(BasicPlugin):
    #     '''
    #     每隔一段时间执行一次任务
    #     注意，以类为插件形式的时候，需要继承 BasicPlugin
    #     此时 api 可以通过 self.frame 访问
    #     2401 只推荐这种插件形式
    #     '''
    #     def __init__(self,name,period,cmds) -> None:
    #         super().__init__()
    #         self.name=name
    #         # 每隔 period 秒，执行一次 cmds
    #         self.period=period
    #         self.cmds=cmds
        
    #     async def on_cmd_result(self,cmd,result):
    #         utils.print(f"指令{cmd}的结果是{result}")
        
    #     async def on_cmd_result_no_name(self,result):
    #         utils.print(f"指令的结果是{result}") 
        
    #     async def run_cmds(self):
    #         for cmd in self.cmds:
    #             await self.frame.send_ws_cmd(cmd=cmd,cb=self.on_cmd_result_no_name)
                
    #             #这样也行:
    #             # async def on_result(result):
    #             #     await self.on_cmd_result(cmd=cmd,result=result)
    #             # await self.frame.send_ws_cmd(cmd=cmd,cb=on_result)
                
    #             # 这样也行:
    #             # result=await self.frame.send_ws_cmd(cmd=cmd,wait_result=True)
    #             # utils.print(result)
            
    #     async def __call__(self):
    #         while True:
    #             await utils.sleep(self.period)
    #             # 严禁使用 time.sleep
    #             utils.print(f"插件:{self.name} 执行计划任务")
    #             await self.run_cmds()
        
    # # 将插件注入:
    # frame.Plugin(Plugin(name="示例插件A"))

    # 不推荐这种插件形式
    # class Plugin(object):
    #     '''如果你不喜欢继承，你也可以这么写，总之只要保证能访问api就行'''
    #     def __init__(self,name) -> None:
    #         super().__init__()
    #         self.name=name
        
    #     async def __call__(self,frame):
    #         # 看起来和函数也差不多是吧
    #         resp=await frame.echo(msg="hello from "+self.name,wait_result=True)
    #         utils.print(resp)

    #     async def any_name(self,frame):
    #         # 如果你不喜欢 __call__
    #         resp=await frame.echo(msg="hello from "+self.name,wait_result=True)
    #         utils.print(resp)

    # # 如果你不喜欢继承，你也可以这么注入一个普通的类
    # # 可以这样:
    # frame.add_plugin(Plugin(name="示例插件B")(frame=frame))
    # # 这样也ok：
    # frame.add_plugin(Plugin(name="示例插件C").any_name(frame=frame))

# 启动，千万别忘了启动！
# frame.start()
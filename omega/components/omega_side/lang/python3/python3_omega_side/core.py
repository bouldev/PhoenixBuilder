from calendar import c
import json
import os,sys 
import asyncio
from dataclasses import dataclass
from typing import Any, Callable, Coroutine,Awaitable, Tuple, TypeVar,Union,List
from types import MethodType,FunctionType

from .msg import *
from . import utils

@dataclass
class Info:
    script_name:str 
    server_addr:str 
    current_dir:str
    
    conn:any=None

class Client(object):
    def __init__(self,info:Info) -> None:
        self.info=info
        self.conn=None
        self.loop=asyncio.get_event_loop()
        self.id=1
        async def default_on_msg_fn(msg):
            print(msg,flush=True)
        self.on_msg_cb=default_on_msg_fn
    
    async def setup_ws_client(self):
        import websockets
        wsconn= await websockets.connect(self.info.server_addr)
        self.info.conn=wsconn
        self.conn=wsconn
        return wsconn

    async def send_request(self,func:str="",args:dict={})->int:
        msgID=self.id
        self.id+=1
        if self.id==24012:
            self.id=1
        msg=json.dumps({
            "client":msgID,
            "function":func,
            "args":args
        },ensure_ascii=False)
        await self.conn.send(msg)
        return msgID

    async def start_recv_loop(self):
        print("开始监听消息",flush=True)
        while True:
            msg=await self.conn.recv()
            await self.on_msg_cb(msg)
        
    async def run(self,plugin_tasks):
        await self.setup_ws_client()
        print(f"激活所有插件,插件数:{len(plugin_tasks)}",flush=True)
        tasks=[asyncio.create_task(self.start_recv_loop())]
        tasks+=[asyncio.create_task(t) for t in plugin_tasks]
        await asyncio.wait(tasks)
             
    def start(self,plugin_tasks):
        self.loop.run_until_complete(self.run(plugin_tasks))


class ResultWaitor(object):
    def __init__(self) -> None:
        self.event=asyncio.Event()
        self.result=None
        
    async def set_result(self,result):
        self.result=result
        self.event.set()
        
    async def wait_result(self)->any:
        await self.event.wait()
        return self.result

# AsyncCallBack 类型： (异步回调方式接受结果，即wait_result=False 时的 cb 参数的类型定义)
# 这个类型是 OmegaSide 中以异步回调方式接受结果或者推送的函数的类型定义
# 这个函数可以是 None，
# 也可以是一个 async def 开头的函数，或者是成员变量
# 例子:
# async def on_result(result):
#     print(result)
# await frame.xxxx(cb=on_result,wait_result=False)
RT=TypeVar("RT") # RT= result type 返回值类型
AsyncCallBack=Union[Callable[[RT],Awaitable],None]

# APIResult 类型： （同步方式接受结果）
# 这个函数是所有 Omega Side Python API 以同步形式返回结果时的返回值的类型定义
# 这个类型的意思是，以同步形式获取结果时，必须以：
# result = await frame.xxx(wait_result=True)
# 方式获得结果，一般来说，这样的效率会比异步低很多
# 注意，异步形式虽然结果由callback接受，但仍然需要 await
APIResult=Awaitable[Union[None,RT]]

class MainFrame(object):
    def __init__(self,client) -> None:
        self.client=client
        self.plugin_tasks=[]
        self.client.on_msg_cb=self._on_msg
        self.on_resp={}
        self.onAnyMCPkt=[]
        self.onTypedMCPkt={}
        self.onMenuTriggered={}
        self.started=False
        
    async def _on_push_msg(self,push_type,sub_type,data):
        if push_type=="mcPkt":
            data["id"]=sub_type
            if sub_type in self.onTypedMCPkt.keys():
                for cb in self.onTypedMCPkt[sub_type]:
                    if cb is not None:
                        await cb(data)
            for cb in self.onAnyMCPkt:
                if cb is not None:
                    await cb(data)
        elif push_type=="menuTriggered":
            await self.onMenuTriggered[sub_type](data)
    
    async def _on_msg(self,msg):
        msg=json.loads(msg)
        msgID=msg["client"]
        data=msg["data"]
        if msgID!=0:
            violate=msg["violate"]
            if violate:
                raise Exception(f"从omega框架收到了 Violate 数据包: {msg}")
            cb=self.on_resp[msgID]
            if cb[2] is None:
                return
            else:
                await cb[1](data,cb[2])
        else:
            await self._on_push_msg(msg["type"],msg["sub"],data)
        
    async def _send_request(self,func:str="",args:dict={},cb=None,wait_result=False):
        msgID=await self.client.send_request(func=func,args=args)
        if not wait_result:
            self.on_resp[msgID]=cb
        else:
            waitor=ResultWaitor()
            self.on_resp[msgID]=(cb[0],cb[1],waitor.set_result)
            return await asyncio.wait([waitor.wait_result()])
    
    def _add_plugin(self,plugin):
        plugin_task=None
        try:
            if isinstance(plugin,Awaitable):
                # 插件可以直接被 await plugin 时
                self.plugin_tasks.append(plugin)
                return
            elif isinstance(plugin,Callable) and isinstance(getattr(plugin,"__call__"),MethodType):
                # 插件是一个类对象，且有成员函数 async def __call__(self): 时
                plugin_task=plugin()
                assert isinstance(plugin_task,Awaitable),"当插件为一个类的实例时，其必须有成员函数 async def __call__(self): "
                self.plugin_tasks.append(plugin_task)
                return
            elif (not isinstance(plugin,MethodType)) and isinstance(plugin,FunctionType):
                # 插件是一个函数，且定义为 async def __call__(frame): 时
                plugin_task=plugin(self)
                self.plugin_tasks.append(plugin_task)
                return
        except Exception:
            pass
        raise Exception(f"""
插件格式错误，其必须为以下几种形式之一:

插件定义(函数形式): 除非你的插件很短，不然不推荐
async def plugin(frame:MainFrame):
    async def cb(msg:str):
        utils.print(msg)
    await frame.echo("hello",cb=cb)

# 注入方式:
frame.add_plugin(plugin)
# 或:
frame.add_plugin(plugin(frame))

插件定义(类形式): 2401 推荐这种插件形式
class Plugin(BasicPlugin):
    '''
    注意，以类为插件形式的时候，需要继承 BasicPlugin
    此时 api 可以通过 self.frame 访问
    2401 推荐这种插件形式
    '''
    def __init__(self,name) -> None:
        super().__init__()
        self.name=name
    
    async def __call__(self):
        utils.print(await self.frame.echo(msg="hello",wait_result=True))

# 将插件注入:
frame.Plugin(Plugin(name="示例插件A"))

class Plugin(object):
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
frame.add_plugin(Plugin(name="示例插件B")(frame=frame))
# 这样也ok：
frame.add_plugin(Plugin(name="示例插件C").any_name(frame=frame))
""")
        
        
    def add_plugin(self,plugin,*plugins):
        self._add_plugin(plugin)
        if len(plugins)>0:
            for more_plugin in plugins:
                self._add_plugin(more_plugin)
            
    async def echo(self,msg:any,cb:AsyncCallBack[str]=None,wait_result:bool=False)->APIResult[str]:
        return await self._send_request(*encode_echo(msg),
                                 cb=("echo",decode_echo,cb),wait_result=wait_result)
    
    async def reg_mc_pkt_by_type(self,pktID:str,cb:AsyncCallBack[Tuple[bool,str]]=None,on_push_cb:AsyncCallBack[dict]=None,wait_result:bool=False)->APIResult[Tuple[bool,str]]:
        if not pktID in self.onTypedMCPkt.keys():
            self.onTypedMCPkt[pktID]=[]
        self.onTypedMCPkt[pktID].append(on_push_cb)
        return await self._send_request(*encode_reg_mc_pkt_by_type(pktID=pktID),
        cb=("reg_mc_pkt_by_type",decode_reg_mc_pkt_by_type,cb),wait_result=wait_result)
    
    async def reg_any_mc_pkt(self,cb:AsyncCallBack[bool]=None,on_push_cb:AsyncCallBack[dict]=None,wait_result:bool=False)->APIResult[bool]:
        self.onAnyMCPkt.append(on_push_cb)
        return await self._send_request(*encode_reg_any_mc_pkt(),
        cb=("reg_any_mc_pkt",decode_reg_any_mc_pkt,cb),wait_result=wait_result)
    
    async def send_ws_cmd(self,cmd:str,cb:AsyncCallBack[str]=None,wait_result:bool=False)->APIResult[str]:
        return await self._send_request(*encode_send_ws_cmd(cmd),
        cb=("send_ws_cmd",decode_send_ws_cmd,cb),wait_result=wait_result)

    async def send_player_cmd(self,cmd:str,cb:AsyncCallBack[dict]=None,wait_result:bool=False)->APIResult[dict]:
        return await self._send_request(*encode_send_player_cmd(cmd),
        cb=("send_player_cmd",decode_send_player_cmd,cb),wait_result=wait_result)
        
    async def send_wo_cmd(self,cmd:str,cb:AsyncCallBack[bool]=None,wait_result:bool=False)->APIResult[bool]:
        return await self._send_request(*encode_send_wo_cmd(cmd),
        cb=("send_wo_cmd",decode_send_wo_cmd,cb),wait_result=wait_result)
    
    async def get_uqholder(self,cb:AsyncCallBack[dict]=None,wait_result:bool=False)->APIResult[dict]:
        return await self._send_request(*encode_get_uqholder(),
        cb=("get_uqholder",decode_get_uqholder,cb),wait_result=wait_result)

    async def get_players_list(self,cb:AsyncCallBack[List[dict]]=None,wait_result:bool=False)->APIResult[List[dict]]:
        return await self._send_request(*encode_get_players_list(),
        cb=("get_players_list",decode_get_players_list,cb),wait_result=wait_result)

    async def reg_menu(self,triggers:List[str],argument_hint:str,usage:str,cb:AsyncCallBack[dict]=None,on_push_cb:AsyncCallBack[dict]=None,wait_result:bool=False)->APIResult[dict]:
        sub_id=f"{triggers}{argument_hint}{usage}"
        self.onMenuTriggered[sub_id]=on_push_cb
        return await self._send_request(*encode_reg_menu(triggers=triggers,argument_hint=argument_hint,usage=usage,sub_id=sub_id),
        cb=("reg_menu",decode_reg_menu,cb),wait_result=wait_result)

    async def get_player_next_input(self,player:str,hint:str,cb:AsyncCallBack[Tuple[bool,str,str]]=None,wait_result:bool=False)->APIResult[Tuple[bool,str,str]]:
        return await self._send_request(*encode_get_player_next_input(player=player,hint=hint),
        cb=("player.next_input",decode_player_next_input,cb),wait_result=wait_result)

    def start(self,*plugins):
        if self.started:
            raise Exception("你不能多次调用start函数")
        self.start=True
        if len(plugins)>0:
            self.add_plugin(*plugins)
        self.client.start(self.plugin_tasks)

frame=None

def get_mainframe() -> MainFrame:
    global frame
    assert frame is not None, "你必须先调用 bootstrap"
    return frame

def bootstrap(addr=None) -> MainFrame:
    global frame
    if frame is not None:
        raise Exception("你不应该重复调用 bootstrap")
    #  获得基本启动信息
    script_name=sys.argv[0]
    if addr==None:
        side_server_addr=sys.argv[1]
        addr="ws://"+side_server_addr+"/omega_side"
    cwd=os.getcwd()
    print(f"脚本名: {script_name} Omega Side 服务器地址: {addr} 当前目录: {cwd}")
    
    info=Info(script_name=script_name,server_addr=addr,current_dir=cwd)
    
    utils.install_lib(lib_name="websockets",lib_install_name="websockets")
    print(f"omega side server addr: {addr}")
    client=Client(info=info)
    frame=MainFrame(client=client)
    return frame
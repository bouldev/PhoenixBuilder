import time
import json
from threading import Event 
from queue import Queue
from traceback import print_stack
from typing import *
from . import bootstrap
from .bootstrap import omega_args

import websocket
from .protocol import *
from easydict import EasyDict

class ResultWaiter(object):
    def __init__(self) -> None:
        self.value=None
        self.event=Event()    
        
    def wait_result(self):
        self.event.wait()
        return self.value
    
    def __call__(self, value):
        self.value=value
        self.event.set()
        

class API(object):
    def __init__(self,frame) -> None:
        self.frame=frame
        self.client_id=1
        self.on_resp_cbs:Mapping[int,Callable[[EasyDict],None]]={}
        self.on_any_mc_pkt_cbs=[]
        self.on_typed_mc_pkt_cbs={}
        self.on_menu_triggered_cbs={}
        self.normal_callbacks={}
        
    def execute_after(self,func:Callable,*args:List[any],delay_time:int):
        def delay_wrapper():
            time.sleep(delay_time)
            func(*args)
        bootstrap.execute_func_in_thread_with_auto_restart(delay_wrapper)
        
    def execute_with_repeat(self,func:Callable,*args:List[any],repeat_time:int):
        def repeat_wrapper():
            next_execute_time=time.time()+repeat_time
            while True:
                sleep_time=next_execute_time-time.time()
                if sleep_time>0:time.sleep(sleep_time)
                next_execute_time=time.time()+repeat_time
                func(*args)
        bootstrap.execute_func_in_thread_with_auto_restart(repeat_wrapper)
        
    def execute_in_individual_thread(self,func:Callable,*args:List[any]):
        bootstrap.execute_func_in_thread_with_auto_restart(func,*args)
        
    def _send_request(self,request_msg:RequestMsg,cb:Callable[[EasyDict],None]):
        msgID=self.client_id
        self.client_id+=1
        if self.client_id==24012:
            self.client_id=1
        request_msg.client=msgID
        _cb=cb
        if cb is None:
            _cb=ResultWaiter()
        self.on_resp_cbs[msgID]=_cb
        self.frame.send_request(request_msg)
        if cb is None:
            return _cb.wait_result()
        else:
            return
        
    def _add_normal_callback(self,type,sub,cb):
        if type not in self.normal_callbacks.keys():
            self.normal_callbacks[type]={}
        if sub not in self.normal_callbacks[type]:
            self.normal_callbacks[type][sub]=[]
        self.normal_callbacks[type][sub].append(cb)
        
    def _on_resp(self,resp:ResponseMsg):
        msgID=resp.client
        data=resp.data
        cb=self.on_resp_cbs[msgID]
        if cb is not None:
            cb(data)
    
    def _on_push(self,data:PushMsg):
        if data.type=="mcPkt":
            data.id=data.sub
            if data.sub in self.on_typed_mc_pkt_cbs.keys():
                for cb in self.on_typed_mc_pkt_cbs[data.sub]:
                    if cb is not None:
                        cb(data["data"])
            for cb in self.on_any_mc_pkt_cbs:
                if cb is not None:
                    cb(data)
        elif data.type=="menuTriggered":
            bootstrap.AutoRestartThreadContainer(self.on_menu_triggered_cbs[data.sub],data['data']).start()
        else:
            for cb in self.normal_callbacks[data.type][data.sub]:
                cb(data.data)
    
    def do_echo(self,msg:str,cb:Callable[[EchoResp],None])->EchoResp:
        return self._send_request(RequestMsg(function="echo",args={"msg":msg}),cb=cb)
    
    def do_send_ws_cmd(self,cmd:str,cb:Callable[[CmdResp],None])->CmdResp:
        return self._send_request(RequestMsg(function="send_ws_cmd",args={"cmd":cmd}),cb=cb)
    
    def do_send_player_cmd(self,cmd:str,cb:Callable[[CmdResp],None])->CmdResp:
        return self._send_request(RequestMsg(function="send_player_cmd",args={"cmd":cmd}),cb=cb)
    
    def do_send_wo_cmd(self,cmd:str,cb:Callable[[AcknowledgeResp],None])->AcknowledgeResp:
        return self._send_request(RequestMsg(function="send_wo_cmd",args={"cmd":cmd}),cb=cb)

    def do_send_packet(self,packetID:int,jsonStr:str,cb:Callable[[SendPacketAcknowledgeResp],None])->SendPacketAcknowledgeResp:
        return self._send_request(RequestMsg(function="send_packet",args={"packetID": packetID, "jsonStr": jsonStr}),cb=cb)
    
    def do_get_uqholder(self,cb:Callable[[dict],None])->dict:
        return self._send_request(RequestMsg(function="get_uqholder",args={}),cb=cb)
    
    def do_get_new_uqholder(self,cb:Callable[[dict],None])->dict:
        return self._send_request(RequestMsg(function="get_new_uqholder",args={}),cb=cb)
    
    def do_get_players_list(self,cb:Callable[[List[PlayerInfo]],None])->List[PlayerInfo]:
        return self._send_request(RequestMsg(function="get_players_list",args={}),cb=cb)
    
    def do_get_get_player_next_param_input(self,player:str,hint:str,cb:Callable[[PlayerParamInput],None])->PlayerParamInput:
        return self._send_request(RequestMsg(function="player.next_input",args={"player":player,"hint":hint}),cb=cb)
    
    def do_send_player_msg(self,player:str,msg:str,cb:Callable[[AcknowledgeResp],None])->AcknowledgeResp:
        return self._send_request(RequestMsg(function="player.say_to",args={"player":player,"msg":msg}),cb=cb)
    
    def do_set_player_title(self,player:str,msg:str,cb:Callable[[AcknowledgeResp],None])->AcknowledgeResp:
        return self._send_request(RequestMsg(function="player.title_to",args={"player":player,"msg":msg}),cb=cb)
    
    def do_set_player_subtitle(self,player:str,msg:str,cb:Callable[[AcknowledgeResp],None])->AcknowledgeResp:
        return self._send_request(RequestMsg(function="player.subtitle_to",args={"player":player,"msg":msg}),cb=cb)
        
    def do_set_player_actionbar(self,player:str,msg:str,cb:Callable[[AcknowledgeResp],None])->AcknowledgeResp:
        return self._send_request(RequestMsg(function="player.actionbar_to",args={"player":player,"msg":msg}),cb=cb)
    
    def do_get_player_pos(self,player:str,limit:str,cb:Callable[[PlayerPoseResp],None])->PlayerPoseResp:
        return self._send_request(RequestMsg(function="player.pos",args={"player":player,"limit":limit}),cb=cb)
    
    def do_set_player_data(self,player:str,entry:str,data:any,cb:Callable[[AcknowledgeResp],None])->AcknowledgeResp:
        return self._send_request(RequestMsg(function="player.set_data",args={"player":player,"entry":entry,"data":data}),cb=cb)
    
    def do_get_player_data(self,player:str,entry:any,cb:Callable[[PlayerDataResponse],None])->PlayerDataResponse:
        return self._send_request(RequestMsg(function="player.get_data",args={"player":player,"entry":entry}),cb=cb)
    
    def do_get_item_mapping(self,cb:Callable[[ItemMappingResp],None])->ItemMappingResp:
        return self._send_request(RequestMsg(function="query_item_mapping",args={}),cb=cb)
    
    def do_get_block_mapping(self,cb:Callable[[BlockMappingResp],None])->BlockMappingResp:
        return self._send_request(RequestMsg(function="query_block_mapping",args={}),cb=cb)
    
    def do_get_scoreboard(self,cb:Callable[[Dict[str,Dict[str,int]]],None])->Callable[[Dict[str,Dict[str,int]]],None]:
        return self._send_request(RequestMsg(function="query_memory_scoreboard",args={}),cb=cb)
    
    def do_send_fb_cmd(self,cmd:str,cb:Callable[[AcknowledgeResp],None])->AcknowledgeResp:
        return self._send_request(RequestMsg(function="send_fb_cmd",args={"cmd":cmd}),cb=cb)
    
    def do_send_qq_msg(self,msg:str,cb:Callable[[AcknowledgeResp],None])->AcknowledgeResp:
        return self._send_request(RequestMsg(function="send_qq_msg",args={"msg":msg}),cb=cb)
    
    def listen_omega_menu(self,triggers:List[str],argument_hint:str,usage:str,cb:Callable[[RegMenuResp],None],on_menu_invoked=Callable[[PlayerInput],None])->RegMenuResp:
        sub_id=f"{triggers}{argument_hint}{usage}"
        self.on_menu_triggered_cbs[sub_id]=on_menu_invoked
        return self._send_request(RequestMsg(function="reg_menu",args={"triggers":triggers,"argument_hint":argument_hint,"usage":usage,"sub_id":sub_id}),cb=cb)
    
    def listen_mc_packet(self,pkt_type:str,cb:Callable[[ListenPacketAcknowledgeResp],None],on_new_packet_cb:Callable[[dict],None])->ListenPacketAcknowledgeResp:
        if isinstance(pkt_type,int):
            pkt_type=self._send_request(RequestMsg(function="query_packet_name",args={"pktID":pkt_type}),cb=cb).name
            if pkt_type==None:
                raise ValueError(f"{pkt_type} is not valid type")
        if not pkt_type in self.on_typed_mc_pkt_cbs.keys():
            self.on_typed_mc_pkt_cbs[pkt_type]=[]
        self.on_typed_mc_pkt_cbs[pkt_type].append(on_new_packet_cb)
        return self._send_request(RequestMsg(function="reg_mc_packet",args={"pktID":pkt_type}),cb=cb)

    def listen_any_mc_packet(self,cb:Callable[[ListenPacketAcknowledgeResp],None],on_new_packet_cb:Callable[[dict],None])->ListenPacketAcknowledgeResp:
        self.on_any_mc_pkt_cbs.append(on_new_packet_cb)
        return self._send_request(RequestMsg(function="reg_mc_packet",args={"pktID":"all"}),cb=cb)
    
    def listen_player_login(self,cb:Callable[[AcknowledgeResp],None],on_player_login_cb:Callable[[PlayerInfo],None])->AcknowledgeResp:
        self._add_normal_callback("playerLogin",sub="",cb=on_player_login_cb)
        return self._send_request(RequestMsg(function="reg_login",args={}),cb=cb)

    def listen_player_logout(self,cb:Callable[[AcknowledgeResp],None],on_player_logout_cb:Callable[[PlayerInfo],None])->AcknowledgeResp:
        self._add_normal_callback("playerLogout",sub="",cb=on_player_logout_cb)
        return self._send_request(RequestMsg(function="reg_logout",args={}),cb=cb)

    def listen_block_update(self,cb:Callable[[AcknowledgeResp],None],on_block_update:Callable[[PlayerInfo],None])->AcknowledgeResp:
        self._add_normal_callback("blockUpdate",sub="",cb=on_block_update)
        return self._send_request(RequestMsg(function="reg_block_update",args={}),cb=cb)

class MainFrame(object):
    def __init__(self):
        self.ws:websocket.WebSocketApp=None
        self.is_running:bool=False
        self.api:API=None
        self.plugins:List[Callable[[API]]]=[]
        self.running_threads:Mapping[Callable[[API],None],bootstrap.AutoRestartThreadContainer]={}
        self.send_queue=Queue(maxsize=1024)
        self.wsa=None
        
    def _on_open(self,ws):
        self.ws=ws
        self.api=self._get_api()
        def send_loop():
            while True:
                ws.send(self.send_queue.get())
        bootstrap.execute_func_in_thread_with_auto_restart(func=send_loop,auto_restart=False)
        self._run_existing_plugins()
        print("已经打开到 Omega 框架的数据通道")
        
    def _add_plugin(self,plugin:Callable[[API],None]):
        self.plugins.append(plugin)
        if self.api is not None:
            thread_container=bootstrap.AutoRestartThreadContainer(plugin,self.api,
                                                           exit_on_program_terminate=True,
                                                           auto_restart=False,only_restart_on_err=False)
            self.running_threads[plugin]=thread_container
            thread_container.start()
        
    def add_plugin(self,plugin:Callable[[API],None],*args:Callable[[API],None]):
        self._add_plugin(plugin)
        for p in args:
            self._add_plugin(p)
    
    def _run_existing_plugins(self):
        for plugin in self.plugins:
            thread_container=bootstrap.AutoRestartThreadContainer(plugin,self.api,
                                                           exit_on_program_terminate=True,
                                                           auto_restart=False,only_restart_on_err=False)
            self.running_threads[plugin]=thread_container
        for k,v in self.running_threads.items():
            v.start()
    def _on_close(self,*args):
        print("与Omega框架的连接已经断开")
        exit(0)
        
    def _on_error(self,ws, error):
        print_stack()
        print("ws error: ",error)
        exit(0)
    
    def _on_resp(self,resp:ResponseMsg):
        self.api._on_resp(resp=EasyDict(resp))
        
    def _on_push(self,data:PushMsg):
        self.api._on_push(data=EasyDict(data))
    
    def on_message(self,ws,message):
        msg=json.loads(message)
        msgID=msg["client"]
        if msgID!=0:
            violate=msg["violate"]
            if violate:
                raise Exception(f"从omega框架收到了 Violate 数据包: {msg}")
            self._on_resp(msg)
        else:
            self._on_push(msg)
            
    def send_request(self,req:RequestMsg):
        self.send_queue.put(req.to_json())
        # self.ws.send(req.to_json())
            
    def _get_api(self)->API:
        if self.api is None:
            self.api=API(self)
        return self.api
        
    def connect(self,addr:str):
        ws = websocket.WebSocketApp(addr,
                            on_open=self._on_open,
                            on_message=self.on_message,
                            on_error=self._on_error,
                            on_close=self._on_close)
        ws.run_forever()
        return self 
    
    def run(self,addr:str=None):
        if self.is_running:
            return
        self.is_running=True
        if addr is None:
            addr=omega_args.ws_server_addr
        bootstrap.execute_func_in_thread_with_auto_restart(self.connect,addr,exit_on_program_terminate=False)
        # self.connect(addr)
        
frame=MainFrame()


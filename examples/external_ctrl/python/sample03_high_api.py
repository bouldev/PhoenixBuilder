import time

from proxy import forward
from proxy import high_api
from proxy import utils
from proxy import packets as P

# high api, 并非可复用的形式出现，而是以类似模版的方式出现，功能是作为参数传入的
# 所以才叫 high_level_api, 可以更专注的实现功能
# 当 high_level_api 不能实现某个功能时，应该参考源码写另一个模版

conn=forward.connect_to_fb_transfer(host="localhost",port=8000)
db={
    'unkown_packet_type':{},
    'command_resp':{},
    'players_login':{},
    'players_off':{},
}

class Reactor(object):
    def __init__(self,db=None) -> None:
        self.db=db
        self.reaction={
            P.IDLogin:self.on_login,
            P.IDCommandOutput:self.on_command_output,
            P.IDText:self.on_text,
            P.IDMovePlayer:self.on_move_player,
            P.IDSetTime:self.on_set_time,
        }
        self.command_resp=self.db['command_resp']
    
    def on_login(self,msg:P.Login):
        print('user login')
    
    def on_command_output(self,msg:P.CommandOutput):
        uuid_bytes=msg.CommandOrigin.UUID
        if uuid_bytes in self.command_resp.keys():
            reaction=self.command_resp[uuid_bytes]
            del self.command_resp[uuid_bytes]
            if reaction is not None:
                return reaction(msg)
            # else:
            #     print(msg)
        print(msg)
    
    def on_text(self,msg:P.Text):
        if msg.Message=='§e%multiplayer.player.joined':
            msgs=[]
            for name in msg.Parameters:
                localtime = time.asctime( time.localtime(time.time()) )
                self.db['players_login'][name]=localtime
                print(f'user login: {name} {localtime}')
                msg,_=utils.pack_command('tellraw @a {"rawtext" : [{"text":"§6§lOmega System: '+name+'欢迎回来!"}]}')
                msgs.append(msg)
            return msg
        if msg.Message=="菜单":
            msg,_=utils.pack_command('tellraw '+msg.SourceName+' {"rawtext" : [{"text":"§6§l我们有苹果葡萄泡泡糖，请问您要什么呢？"}]}')
            return msg
        print(msg)
        
    def on_move_player(self,msg:P.MovePlayer):
        # print('Player move :',msg)
        pass
    
    def on_set_time(self,msg:P.SetTime):
        # well 这里好像有问题
        # print('set time:',msg.Time)
        pass
    
    def on_not_implemented_msg(self,packet_id,bytes_msg):
        if packet_id not in self.db['unkown_packet_type'].keys():
            self.db['unkown_packet_type'][packet_id]=bytes_msg
            print(f'new unkown packet type {packet_id} {bytes_msg}')
        return None
    
    def on_msg(self,packet_id,msg,sender_subclient,target_subclient):
        cb=self.reaction.get(packet_id)
        if cb is None:
            return None
        return cb(msg)
    
    def __call__(self,bytes_msg,decoded_msg):
        (packet_id,msg)=decoded_msg
        if msg is None:
            return self.on_not_implemented_msg(packet_id,bytes_msg)
        else:
            _msg,sender_subclient,target_subclient=msg
            return self.on_msg(packet_id,_msg,sender_subclient,target_subclient)

class UserInteract(object):
    def __init__(self) -> None:
        super().__init__()
        self.feedbackon,_=utils.pack_command('gamerule sendcommandfeedback true',uuid=None)
        self.feedbackoff,_=utils.pack_command('gamerule sendcommandfeedback false',uuid=None)
        
    def __call__(self):
        time.sleep(0.1)
        command=input("")
        if command.startswith('.'):
            msg,uuid_bytes=utils.pack_command(command[1:],uuid=None)
            db['command_resp'][uuid_bytes]=lambda m:print('cmd resp :',m)
            return [
                # self.feedbackon,
                msg,
                # self.feedbackoff
            ]
        elif command.startswith(':'):
            msg,_=utils.pack_command('tellraw @a {"rawtext" : [{"text":"'+command[1:]+'"}]}')
            return msg 

class KeepAlive(object):
    def __init__(self) -> None:
        super().__init__()
        
    def on_restart(self):
        msg,_=utils.pack_command('tellraw @a {"rawtext" : [{"text":"§6§lOmega System: I Back To Serve!"}]}')
        return msg
    
    def __call__(self):
        '''每隔5秒发送一条指令，防止因为过长时间不发送数据被服务器踢掉'''
        msg,uuid_bytes=utils.pack_command('testforblock ~~~ air',uuid=None)
        time.sleep(15)
        return msg

recv_func=Reactor(db)
workers={
    'keep_alive':KeepAlive(),
    'user_interact':UserInteract(),
}
'''
    recv_func: 接收函数，每当有新消息到来时，该函数被调用，该函数可以有返回也可以无返回，返回值将被转发
    workers: 字典，{名字:workers} ，每个worker将在单独线程中工作，并被循环调用，每次调用可以有返回也可以无返回，返回值将被转发
    respawn: 当连接断开时是否需要重连
'''
high_api.one_sender_multiworkers(conn,recv_func,workers,respawn=True)
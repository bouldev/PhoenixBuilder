from proxy import forward
from proxy import utils
from threading import Thread
from queue import Queue
import time

# 有两个子线程，一个负责停的解析数据，并通过 Queue 将解析结果在线程之间传递
# 另一个子线程的目仅仅负责发送指令
# 当任意子线程死亡时，主线程尝试重连


class Config(object):
    def __init__(self) -> None:
        self.recv_thread_alive=True
        self.working_threads_alive={}
        self.receiver=None
        self.sender=None
config=Config()

def recv_thread_func(recv_queue:Queue):
    while True:
        while config.receiver is None:
            time.sleep(1)
        config.recv_thread_alive=True
        print('recv thread activated!')
        try:
            while True:
                bytes_msg,(packet_id,decoded_msg)=config.receiver()
                if decoded_msg is None:
                    # 还未实现该类型数据的解析(会有很多很多的数据包！)
                    # print(f'unkown decode packet ({packet_id}): ',bytes_msg)
                    continue
                else:
                    # 已经实现类型数据的解析
                    msg,sender_subclient,target_subclient=decoded_msg
                    print(msg)
                    recv_queue.put(msg)
        except Exception as e:
            print('Recv thread terminated!',e)
            config.recv_thread_alive=False
            config.receiver=None
            config.sender=None
            print('Recv thread waiting for restarting...') 
            time.sleep(3)

def working_thread_func(thread_name):
    msg=None
    while True:
        while (config.sender is None) or (not config.recv_thread_alive):
            time.sleep(1)
        config.working_threads_alive[thread_name]=True
        print(f'working thread [{thread_name}] activated!')
        try:
            while True:
                if msg is None:
                    command=input('cmd:')
                    msg,uuid_bytes=utils.pack_ws_command(command,uuid=None)
                    print(uuid_bytes)
                config.sender(msg)
                msg=None
                time.sleep(0.1)
        except Exception as e:
            print(f'Working thread [{thread_name}] terminated!',e)
            config.working_threads_alive[thread_name]=False
            config.receiver=None
            config.sender=None
            print('Working thread waiting for restarting...') 
            time.sleep(3)
            
conn=forward.connect_to_fb_transfer(host="localhost",port=8000)

config.sender=forward.Sender(connection=conn)
config.receiver=forward.Receiver(connection=conn)
recv_queue = Queue(maxsize=10240)
recv_thread = Thread(target=recv_thread_func, args=(recv_queue,))
work_thread = Thread(target=working_thread_func, args=('user_interact',))

recv_thread.daemon = True
recv_thread.start()
work_thread.daemon = True
work_thread.start()   

while True:
    time.sleep(0.1)
    if (not config.recv_thread_alive) or (False in config.working_threads_alive.keys()):
        print('sub process crashed! tring to restart connection...')
        while True:
            time.sleep(3)
            try:
                conn=forward.connect_to_fb_transfer(host="localhost",port=8000)
                config.sender=forward.Sender(connection=conn)
                config.receiver=forward.Receiver(connection=conn)
                break
            except Exception as e:
                print(f'restart error : {e} ... continue retry')

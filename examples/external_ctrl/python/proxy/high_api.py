from socket import socket
from proxy import forward
from proxy import utils
from threading import Thread
from queue import Queue
import time
from typing import Any, Callable, Iterable

# 理论上 epoll 和 asyncio 才是更好的选择，但是这俩api都比较新，暂时先不用了
def one_sender_multiworkers(conn:socket,recv_func:Callable,workers:dict=None,respawn=True):
    if workers is None:
        workers={}
    class Config(object):
        def __init__(self) -> None:
            self.recv_thread_alive=True
            self.working_threads_alive={}
            self.receiver=None
            self.sender=None
    config=Config()
    
    def send_responses(response):
        if response is None:
            pass 
        elif isinstance(response,list):
            while len(response):
                config.sender(response[0])
                response.pop(0)
        else:
            config.sender(response)
        return None
    
    def recv_thread_func(recv_func:Callable):
        if hasattr(recv_func,'on_start'):
            send_responses(recv_func.on_start())
        response=None
        while True:
            while config.receiver is None:
                time.sleep(1)
            config.recv_thread_alive=True
            print('recv thread activated!')
            try:
                if hasattr(recv_func,'on_restart'):
                    send_responses(recv_func.on_restart())
                while True:
                    if response is None:
                        bytes_msg,decoded_msg=config.receiver()
                        response=recv_func(bytes_msg,decoded_msg)
                    response=send_responses(response)
            except Exception as e:
                print('Recv thread terminated!',e)
                config.recv_thread_alive=False
                config.receiver=None
                config.sender=None
                print('Recv thread waiting for restarting...') 
                time.sleep(1)
    def working_thread_func(thread_name:str,worker:Callable):
        if hasattr(worker,'on_start'):
            send_responses(worker.on_start())
        msg=None
        while True:
            while (config.sender is None) or (not config.recv_thread_alive):
                time.sleep(1)
            config.working_threads_alive[thread_name]=True
            print(f'working thread [{thread_name}] reactivated!')
            try:
                if hasattr(worker,'on_restart'):
                    send_responses(worker.on_restart())
                while True:
                    if msg is None:
                        msg=worker()
                    msg=send_responses(msg)
            except Exception as e:
                print(f'Working thread [{thread_name}] terminated!',e)
                config.working_threads_alive[thread_name]=False
                config.receiver=None
                config.sender=None
                print('Working thread waiting for restarting...') 
                time.sleep(1)
    config.sender=forward.Sender(connection=conn)
    config.receiver=forward.Receiver(connection=conn)
    recv_thread = Thread(target=recv_thread_func, args=(recv_func,))
    recv_thread.daemon = True
    recv_thread.start()
    for worker_name,worker in workers.items():
        working_thread = Thread(target=working_thread_func, args=(worker_name,worker))
        working_thread.daemon = True
        working_thread.start()  
        
    while True:
        time.sleep(0.1)
        if (not config.recv_thread_alive) or (False in config.working_threads_alive.keys()):
            if not respawn:
                print('sub process crashed! quitting...') 
                exit(-1)
            else:
                print('sub process crashed! tring to restart connection...')
                while True:
                    time.sleep(1)
                    try:
                        conn=forward.connect_to_fb_transfer(host="localhost",port=8000)
                        config.sender=forward.Sender(connection=conn)
                        config.receiver=forward.Receiver(connection=conn)
                        break
                    except Exception as e:
                        print(f'restart error : {e} ... continue retry')
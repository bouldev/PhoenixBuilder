import socket
import threading
import struct

from numpy.lib.arraysetops import isin
from .buffer_io import BufferDecoder
from .buffer_io import BufferEncoder
from .packets_io import encode
from .packets_io import decode

def connect_to_fb_transfer(host="localhost",port=8000):
    '''python 作为server端，需要某种手段建立和fb的链接'''
    addr = (host,port)
    proxy = socket.socket(socket.AF_INET,socket.SOCK_STREAM)
    print(f'Try connecting to @ {addr} ...')
    proxy.connect(addr)
    print(f'Connect success')
    return proxy

class Sender(object):
    def __init__(self,connection:socket.socket,SenderSubClient:int=0,TargetSubClient:int=0) -> None:
        self.conn=connection 
        self.mutex = threading.Lock()
        self.SenderSubClient=SenderSubClient
        self.TargetSubClient=TargetSubClient
        
    def send_bytes(self,msg:bytes):
        msg_len=len(msg)+4
        full_msg=struct.pack('I',msg_len)+msg
        current_send=0
        self.mutex.acquire()
        while current_send<msg_len:
            bytes_sent=self.conn.send(full_msg[current_send:])
            current_send+=bytes_sent
            if bytes_sent==0:
                raise Exception('send 0 bytes ... connection maybe closed')
        self.mutex.release()
        
    def send_fb_cmd_str(self,msg:str):
        msg=msg.encode(encoding="utf-8")
        msg_len=len(msg)+4
        full_msg=struct.pack('I',msg_len+2**30)+msg
        current_send=0
        self.mutex.acquire()
        while current_send<msg_len:
            bytes_sent=self.conn.send(full_msg[current_send:])
            current_send+=bytes_sent
            if bytes_sent==0:
                raise Exception('send 0 bytes ... connection maybe closed')
        self.mutex.release()
        
    def __call__(self, msg):
        if isinstance(msg,bytes):
            return self.send_bytes(msg)
        elif isinstance(msg,str):
            self.send_fb_cmd_str(msg)
        else:
            msg=encode(msg,self.SenderSubClient,self.TargetSubClient)
            return self.send_bytes(msg)         
    
class Receiver(object):
    def __init__(self,connection:socket.socket):
        '''
            不要在多个线程上运行...即使加互斥锁也不行!
            **总是**应该只有一个专门的线程负责读取
        '''
        self.conn=connection
        self.buffed_bytes=b''
        
    def __call__(self):
        buffed_bytes=self.buffed_bytes
        current_bytes=len(buffed_bytes)
        required_bytes=0
        if current_bytes>=4:
            required_bytes = struct.unpack('I',buffed_bytes[:4])[0]
        while True:
            if required_bytes==0:
                recv_bytes=self.conn.recv(4-current_bytes)
                if recv_bytes==b'':
                    raise Exception('recv 0 bytes ... connection maybe closed')
                buffed_bytes+=recv_bytes
                current_bytes=len(buffed_bytes)
                if current_bytes>=4:
                    required_bytes = struct.unpack('I',buffed_bytes[:4])[0]
            if current_bytes<required_bytes:
                recv_bytes=self.conn.recv(required_bytes-current_bytes)
                if recv_bytes==b'':
                    raise Exception('recv 0 bytes ... connection maybe closed')
                buffed_bytes+=recv_bytes
                current_bytes=len(buffed_bytes)
            if current_bytes>=required_bytes:
                msg=buffed_bytes[4:required_bytes]
                self.buffed_bytes=buffed_bytes[required_bytes:]
                return msg,decode(msg)
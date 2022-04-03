import ctypes
from typing import List

class intGoSlice(ctypes.Structure):
    _fields_ = [("data", ctypes.POINTER(ctypes.c_longlong)),
                ("len", ctypes.c_longlong),
                ("cap", ctypes.c_longlong)]

class byteGoSlice(ctypes.Structure):
    _fields_ = [("data", ctypes.POINTER(ctypes.c_char)),
                ("len", ctypes.c_longlong),
                ("cap", ctypes.c_longlong)]

class GoLibUtils(object):
    intGo = ctypes.c_longlong
    floatGo = ctypes.c_double
    stringGo = ctypes.c_char_p
    boolGo = ctypes.c_bool

    @staticmethod
    def load_lib(lib_path:str):
        return ctypes.cdll.LoadLibrary(lib_path)
    
    @staticmethod
    def setup_func(func, arg_types=None, res_type=None):
        if arg_types is not None:
            func.argtypes = arg_types
        if res_type is not None:
            func.restype = res_type
        return func
    
    @staticmethod
    def to_go_string(string:str):
        return ctypes.c_char_p(bytes(string, encoding='UTF-8'))

    @staticmethod   
    def to_py_string(string:bytes):
        return string.decode('UTF-8')
    
    @staticmethod
    def to_go_int(int_:int):
        return ctypes.c_longlong(int_)

    @staticmethod
    def to_py_int(int_):
        return int_
    
    @classmethod
    def list_to_slice(cls,ls: list, data_type) -> ctypes.Structure:
        ls=[data_type(x)for x in ls ]
        length = len(ls)
        if data_type is None:
            if length > 0:
                data_type = type(ls[0])
            else:
                raise ValueError('slice has 0 elements')

        kwargs = {
            'data': (data_type * length)(*ls),
            'len': length,
            'cap': length
        }

        if data_type == cls.intGo:
            slc = intGoSlice(**kwargs)
        else:
            raise NotImplementedError(f'data_type {data_type} not supported')
        return slc

    @staticmethod
    def slice_to_list(slc):
        ls = []
        for i in range(slc.len):
            ls.append(slc.data[i])
        return ls
    
    @classmethod
    def to_go_bytes_slice(cls,bs:bytes):
        l=len(bs)
        kwargs = {
            'data': (ctypes.c_char * l)(*bs),
            'len': l,
            'cap': l
        }
        return  byteGoSlice(**kwargs)


class FBClientLibWrapper(GoLibUtils):
    lib = None
    isInit=False

    @classmethod
    def init(cls,lib_path):
        FBClientLibWrapper.lib=cls.load_lib(lib_path)
        cls._ConnectFB=cls.setup_func(
            cls.lib.ConnectFB,
            [cls.stringGo],
            cls.intGo
        )
        cls._ReleaseConnByID=cls.setup_func(
            cls.lib.ReleaseConnByID,
            [cls.intGo]
        )

        cls._RecvFrame=cls.setup_func(
            cls.lib.RecvFrame,
            [cls.intGo],
            cls.stringGo
        )
    
    @classmethod
    def ConnectFB(cls,address):
        connID=cls._ConnectFB(
            cls.to_go_string(address)
        )
        connID=cls.to_py_int(connID)
        if connID==-1:
            raise Exception("Connection Fail")
        return FBClient(connID)
    
    @classmethod
    def ReleaseConnByID(cls,connID):
        cls._ReleaseConnByID(
                cls.to_go_int(connID),
            )

    @classmethod
    def RecvFrame(cls,connID):
        return cls._RecvFrame(
            cls.to_go_int(connID),
        )

class FBClient(object):
    def __init__(self,connID) -> None:
        self.connID=connID

    def Close(self):
        FBClientLibWrapper.ReleaseConnByID(self.connID)

    def RecvFrame(self):
        r= FBClientLibWrapper.RecvFrame(self.connID)
        if r==b'':
            self.Close()
            raise Exception("recv 0 data, connection maybe closed")
        return r 

if __name__=="__main__":
    FBClientLibWrapper.init("./fb_conn.so")
    client=FBClientLibWrapper.ConnectFB("127.0.0.1:5678")
    print(client)
    while True:
        frame=client.RecvFrame()
        print(frame)
    

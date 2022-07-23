from dataclasses import dataclass,field
from http import client
from typing import Dict, List
from dataclasses_json import dataclass_json

@dataclass_json
@dataclass
class RequestMsg:
    client:int=0
    function:str=""
    args:dict=field(default_factory=lambda :{})
    
@dataclass_json
@dataclass
class ResponseMsg:
    client:int=0
    violate:bool=False
    data:dict=field(default_factory=lambda :{})
    
@dataclass_json
@dataclass
class PushMsg:
    client:int=0
    type:str=""
    sub:str=""
    data:dict=field(default_factory=lambda :{})


# frame.do_xxx
@dataclass
class EchoResp:
    msg:str=""

@dataclass 
class CmdOutputMsg:
    Success:bool=False
    Message:str=""
    Parameters:List[str]=None
    DataSet:str=""

@dataclass
class CmdOrigin:
    Origin:int=0
    UUID:str=""
    RequestID:str=""
    PlayerUniqueID:int=0

@dataclass
class CmdResult:
    CommandOrigin:CmdOrigin=None
    OutputType:int=0
    SuccessCount:int=0
    OutputMessages:List[CmdOutputMsg]=None

@dataclass 
class CmdResp:
    result:CmdResult=None
    
@dataclass 
class AcknowledgeResp:
    ack:bool=False
    
@dataclass
class PlayerInfo:
    name:str=""
    runtimeID:int=0
    uuid:str=""
    uniqueID:int=0
    
@dataclass 
class PlayerParamInput:
    success:bool=False
    player:str=""
    input:str=""
    err:str=""
    
@dataclass
class RegMenuResp:
    sub_id:str=""
    
@dataclass
class PlayerInput:
    Name:any=None
    Msg:List[str]=None
    Type:int=0
    FrameWorkTriggered:bool=False
    Aux:any=None
    
@dataclass
class ListenPacketAcknowledgeResp:
    succ:bool=False
    err:str=""
    
@dataclass
class PlayerPoseResp:
    success:bool=False
    pos:List[int]=None
    
@dataclass
class PlayerDataResponse:
    found:bool=False
    data:any=None

@dataclass 
class SimpleBlockDefine:
    Name:str=""
    Val:int=0

@dataclass
class FullBlockDefine:
    name:str=""
    props:dict=None
    found:bool=False

@dataclass
class BlockUpdateInfo:
    pos:List[int]=None
    origin_block_runtime_id:int=0
    origin_block_simple_define:dict=None
    origin_block_full_define:dict=None
    new_block_runtime_id:int=0
    new_block_simple_define:dict=None
    new_block_full_define:dict=None

@dataclass 
class ItemDescribe:
    name:str=""
    maxDamage:int=0

@dataclass 
class ItemMappingResp:
    mapping:Dict[str,ItemDescribe]
    
@dataclass
class BlockMappingResp:
    blocks:List[dict]=None
    simple_blocks:List[dict]=None
    java_blocks:Dict[str,int]=None
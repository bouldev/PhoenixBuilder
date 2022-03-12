from uuid import uuid1,UUID
from .packets import *

def pack_ws_command(command:str,uuid:UUID=None):
    '''返回的uuid_bytes代表附加于这条指令的uuid，bytes，大端序'''
    if uuid is None:
        uuid=uuid1()
        uuid_bytes=uuid.bytes
    if isinstance(uuid,UUID):
        uuid_bytes=uuid.bytes
        
    # 搞不懂fb究竟想干嘛
    request_id="96045347-a6a3-4114-94c0-1bc4cc561694"
    
    origin=CommandOrigin(
        Origin=CommandOriginAutomationPlayer,
        UUID=uuid_bytes,
        RequestID=request_id,
        PlayerUniqueID=0
    )
    commandRequest=CommandRequest(
        CommandLine=command,
        CommandOrigin=origin,
        Internal=False,
        UnLimited=False
    )
    return commandRequest,uuid_bytes

def pack_command(command:str,uuid:UUID=None):
    '''返回的uuid_bytes代表附加于这条指令的uuid，bytes，大端序'''
    if uuid is None:
        uuid=uuid1()
        uuid_bytes=uuid.bytes
    if isinstance(uuid,UUID):
        uuid_bytes=uuid.bytes
        
    # 搞不懂fb究竟想干嘛
    request_id="96045347-a6a3-4114-94c0-1bc4cc561694"
    
    origin=CommandOrigin(
        Origin=CommandOriginPlayer,
        UUID=uuid_bytes,
        RequestID=request_id,
        PlayerUniqueID=0
    )
    commandRequest=CommandRequest(
        CommandLine=command,
        CommandOrigin=origin,
        Internal=False,
        UnLimited=False
    )
    return commandRequest,uuid_bytes

def pack_wo_command(command:str):
    return SettingsCommand(command,SuppressOutput=True)

def send_chat(text:str,source='Omega System',XUID=''):
    return Text(
        TextType=TextTypeChat,
        NeedsTranslation=False,
        SourceName=source,
        Message=text,
        XUID=''
    )
    
from time import daylight
from typing import Any, Dict, List
from uuid import UUID
from dataclasses import dataclass,field

# only a very small part of packets...
# go to fb/minecraft/protocol/packet/pool.go for full list

IDLogin=1
IDText=9 
IDSetTime=10
IDMovePlayer=19
IDMobEquipment=31 
IDPlayerList=63
IDClientBoundMapItemData=67
IDGameRulesChanged=72
IDCommandRequest=77
IDCommandBlockUpdate=78
IDCommandOutput=79
IDStructureTemplateDataRequest=123
IDSettingsCommand=140


@dataclass
class BlockPos:
    x:int=0
    y:int=0
    z:int=0
@dataclass
class Vec3:
    x:float=0
    y:float=0
    z:float=0

@dataclass
class ItemType:
	NetworkID:int=0
	MetadataValue:int=0
 
@dataclass
class ItemStack(ItemType):
    Count:int=0
    NBTData:Any=None
    CanBePlacedOn:List[str]=field(default_factory=lambda:[])
    CanBreak:List[str]=field(default_factory=lambda:[])

# Login 
@dataclass
class Login:
    ClientProtocol:int
    ConnectionRequest:bytes
    
    
# Text 
TextTypeChat, TextTypeWhisper, TextTypeAnnouncement=1,7,8
TextTypeRaw, TextTypeTip, TextTypeSystem, TextTypeObject, TextTypeObjectWhisper=0,5,6,9,10
TextTypeTranslation, TextTypePopup, TextTypeJukeboxPopup=2,3,4

@dataclass
class Text:
    TextType:int=0
    NeedsTranslation:bool=False
    SourceName:str=''
    Message:str =''
    Parameters:str =''
    XUID:str=''
    PlatformChatID:str=''
# IDSetTime
@dataclass 
class SetTime:
    Time:int

# MovePlayer
MoveModeTeleport=2
class MovePlayer:
    EntityRuntimeID:int
    Position:tuple
    Pitch:float
    Yaw:float
    HeadYaw:float
    Mode:int 
    OnGround:bool
    RiddenEntityRuntimeID:int 
    TeleportCause:int 
    TeleportSourceEntityType:int
    Counter:int
# IDMobEquipment=31 
@dataclass
class MobEquipment:
    EntityRuntimeID:int=0
    NewItem:ItemStack=field(default_factory=ItemStack)
    InventorySlot:int=0
    HotBarSlot:int=0
    WindowID:int=0
# IDPlayerList 63
@dataclass 
class Skin:
    SkinID:str
    # SkinResourcePatch:bytes
    # SkinImageWidth:int 
    # SkinImageHeight:int
    # SkinData:bytes
@dataclass
class PlayerListEntry:
    UUID:bytes
    EntityUniqueID:int
    Username:str 
    XUID:str 
    PlatformChatID:str 
    BuildPlatform:int
    PlatformChatID:str 
    BuildPlatform:int 
    Skin:Skin
    Teacher:bool
    Host:bool
@dataclass
class PlayerList:
    ActionType:bytes=None
    Entries:List[PlayerListEntry]=None

# CommandOrigin
CommandOriginPlayer=0
CommandOriginDevConsole=3
CommandOriginTest=4
CommandOriginAutomationPlayer=5

@dataclass
class CommandOrigin:
    Origin:int=0
    UUID:bytes=b''
    RequestID:str=''
    PlayerUniqueID:int=0
    
# CommandOutputMessage
@dataclass
class CommandOutputMessage:
    Success:bool=False
    Message:str =''
    Parameters:List[str]=None

# IDGameRulesChanged 72
@dataclass
class GameRulesChanged:
    GameRules:Dict[str,Any]=None

# CommandRequest 77
@dataclass
class CommandRequest:
    CommandLine:str=''
    CommandOrigin:CommandOrigin=None
    Internal:bool=False
    UnLimited:bool=False

# IDCommandBlockUpdate 78
CommandBlockImpulse = 0
CommandBlockRepeat =1
CommandBlockChain =2

CB_FACE_DOWN = 0
CB_FACE_UP = 1
# z--
CB_FACE_NORTH = 2
CB_FACE_ZNN = 2
# z++
CB_FACE_SOUTH = 3
CB_FACE_ZPP = 3
# x--
CB_FACE_WEST = 4
CB_FACE_XNN = 4
# x++
CB_FACE_EAST = 5
CB_FACE_XPP = 5

@dataclass
class CommandBlockUpdate:
    Block:bool=True
    Position:BlockPos=field(default_factory=BlockPos)
    Mode:int=CommandBlockImpulse
    NeedsRedstone:bool=True
    Conditional:bool=False
    MinecartEntityRuntimeID:int=0
    Command:str=""
    LastOutput:str=""
    Name:str=""
    ShouldTrackOutput:bool=True
    TickDelay:int=0
    ExecuteOnFirstTick:bool=True

# CommandOutput 79
@dataclass
class CommandOutput:
    CommandOrigin:CommandOrigin=None
    OutputType:int=0
    SuccessCount:int=0
    OutputMessages:List[CommandOutputMessage]=None
    UnknownString:str=''

# IDStructureTemplateDataRequest 123
StructureTemplateRequestExportFromSave=1
StructureTemplateRequestExportFromLoad=2
StructureTemplateRequestQuerySavedStructure=3


@dataclass
class StructureSettings:
    PaletteName:str="default"
    IgnoreEntities:bool=True
    IgnoreBlocks:bool=False
    Size:BlockPos=field(default_factory=BlockPos)
    Offset:BlockPos=field(default_factory=BlockPos)
    LastEditingPlayerUniqueID:int=0
    Rotation:int=0
    Mirror:int=0
    Integrity:float=0
    Seed:int=0
    Pivot:Vec3=field(default_factory=Vec3)

@dataclass
class StructureTemplateDataRequest:
    StructureName:str="tmp_exp"
    Position:BlockPos=field(default_factory=BlockPos)
    Settings:StructureSettings=field(default_factory=StructureSettings)
    RequestType:int=0

# SettingsCommand
@dataclass
class SettingsCommand:
    CommandLine:str
    SuppressOutput:bool
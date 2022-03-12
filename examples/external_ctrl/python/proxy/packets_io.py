from .packets import *
from .buffer_io import BufferDecoder,BufferEncoder
# only a very small part of packets...
# go to fb/minecraft/protocol/packet/pool.go for full list

def encode_u_block_position(e:BufferEncoder,i:BlockPos):
    e.write_var_int32(i.x)
    e.write_var_uint32(i.y)
    e.write_var_int32(i.z)
    return e


# Login 
def decode_login(d:BufferDecoder):
    return Login(d.read_var_uint32(),d.read_tail())

# Text 
def decode_text(d:BufferDecoder):
    o=Text()
    TextType,NeedsTranslation= int(d.read_byte()),d.read_boolen()
    o.TextType=TextType
    o.NeedsTranslation=NeedsTranslation
    if TextType in [TextTypeChat, TextTypeWhisper, TextTypeAnnouncement]:
        o.SourceName=d.read_str()
        o.Message=d.read_str()
    elif TextType in [TextTypeRaw, TextTypeTip, TextTypeSystem, TextTypeObject, TextTypeObjectWhisper]:
        o.Message=d.read_str()
    elif TextType in [TextTypeTranslation, TextTypePopup, TextTypeJukeboxPopup]:
        o.Message=d.read_str()
        length=d.read_var_uint32()
        o.Parameters=[d.read_str() for _ in range(length)]
    o.XUID=d.read_str()
    o.PlatformChatID=d.read_str()
    return o

def encode_text(e:BufferEncoder,i:Text):
    e.write_byte(i.TextType)
    e.write_boolen(i.NeedsTranslation)
    if i.TextType in [TextTypeChat, TextTypeWhisper, TextTypeAnnouncement]:
        e.write_str(i.SourceName)
        e.write_str(i.Message)
    elif i.TextType in [TextTypeRaw, TextTypeTip, TextTypeSystem, TextTypeObject, TextTypeObjectWhisper]:
        e.write_str(i.Message)
    elif i.TextType in [TextTypeTranslation, TextTypePopup, TextTypeJukeboxPopup]:
        e.write_str(i.Message)
        e.write_var_uint32(len(i.Parameters))
        for p in i.Parameters:
            e.write_str(p)
    e.write_str(i.XUID)
    e.write_str(i.PlatformChatID)
    if i.TextType == TextTypeChat:
        e.write_byte(2)
        e.write_str('PlayerId')
        e.write_str('-12345678')
    return e
        
# IDSetTime
def decode_set_time(d:BufferDecoder):
    # well 这里好像有问题
    o=SetTime(d.read_var_int32())
    return o
# MovePlayer
def decode_move_player(d:BufferDecoder):
    o=MovePlayer()
    o.EntityRuntimeID=d.read_var_uint64()
    o.Position=d.read_vec3()
    o.Pitch=d.read_float32()
    o.Yaw=d.read_float32()
    o.HeadYaw=d.read_float32()
    o.Mode=d.read_byte()
    o.OnGround=d.read_byte()
    o.RiddenEntityRuntimeID=d.read_var_uint64()
    if o.Mode==MoveModeTeleport:
        o.TeleportCause=d.read_int32()
        o.TeleportSourceEntityType=d.read_int32()
    o.Counter=d.read_var_uint64()
    return o

# IDMobEquipment=31 
def decode_item(d:BufferDecoder):
    o=ItemStack()
    o.NetworkID=d.read_var_int32()
    if o.NetworkID==0:
        o.MetadataValue=0
        o.Count=0
        o.CanBePlacedOn=None
        o.CanBreak=None
        return o 
    auxValue=d.read_var_int32()
    o.MetadataValue=(auxValue>>8)&255
    o.Count=auxValue&255
    userDataMarker=d.read_int16()
    if userDataMarker==-1:
        userDataVersion=d.read_uint8()
        if userDataVersion==1:
            o.NBTData=d.read_nbt()
        else:
            raise Exception(f"unexpected item user data version {userDataVersion}")
    elif userDataMarker!=0:
        if userDataMarker<0:
            raise Exception(f"invalid NBT length")
        o.NBTData=d.read_nbt(userDataMarker)
    count=d.read_var_int32()
    if count<0:
        raise Exception('NegativeCountError{Type: "item can be placed on"}')
    if count>1024:
        raise Exception(f"item can be placed on")
    o.CanBePlacedOn=[]
    for i in range(count):
        o.CanBePlacedOn.append(d.read_str())
    
    count=d.read_var_int32()
    if count<0:
        raise Exception('NegativeCountError{Type: "item can break"}')
    if count>1024:
        raise Exception(f"item can break")
    o.CanBreak=[]
    for i in range(count):
        o.CanBreak.append(d.read_str())
    return o
def decode_mob_equipment(d:BufferDecoder):
    o=MobEquipment()
    o.EntityRuntimeID=d.read_var_uint64()
    o.NewItem=decode_item(d)
    o.InventorySlot=d.read_byte()
    o.HotBarSlot=d.read_byte()
    o.WindowID=d.read_byte()
    return o

# IDCommandBlockUpdate
def encode_command_block_update(e:BufferEncoder,i:CommandBlockUpdate):
    e.write_boolen(i.Block)
    if i.Block:
        encode_u_block_position(e,i.Position)
        e.write_var_uint32(i.Mode)
        e.write_boolen(i.NeedsRedstone)
        e.write_boolen(i.Conditional)
    else:
        e.write_var_uint64(i.MinecartEntityRuntimeID)
    e.write_str(i.Command)
    e.write_str(i.LastOutput)
    e.write_str(i.Name)
    e.write_boolen(i.ShouldTrackOutput)
    e.write_var_int32(i.TickDelay)
    e.write_boolen(i.ExecuteOnFirstTick)
    return e

# CommandOrigin
def decode_command_origin_data(d:BufferDecoder):
    o=CommandOrigin()
    o.Origin=d.read_var_uint32()
    o.UUID=d.read_UUID()
    o.RequestID=d.read_str()
    if o.Origin in [CommandOriginDevConsole,CommandOriginTest]: 
        o.PlayerUniqueID=d.read_var_uint32()
    return o

def encode_command_origin_data(e:BufferEncoder,i:CommandOrigin):
    e.write_var_uint32(i.Origin)
    e.write_UUID_bytes(i.UUID)
    e.write_str(i.RequestID)
    if i.Origin == CommandOriginDevConsole or i.Origin == CommandOriginTest:
        e.write_var_uint32(i.PlayerUniqueID)

# CommandOutputMessage
def decode_command_message(d:BufferDecoder):
    o=CommandOutputMessage()
    o.Success=d.read_boolen()
    o.Message=d.read_str()
    count=d.read_var_uint32()
    o.Parameters=[d.read_str() for _ in range(count)]
    return o

# CommandOutput
def decode_command_output(d:BufferDecoder):
    o=CommandOutput()
    o.CommandOrigin=decode_command_origin_data(d)
    o.OutputType=int(d.read_byte())
    o.SuccessCount=d.read_var_uint32()
    count=d.read_var_uint32()
    o.OutputMessages=[decode_command_message(d) for _ in range(count)]
    if o.OutputType==4:
        o.UnknownString=d.read_str()
    return o

# GameRulesChanged 72
def decode_gamerule_changed(d:BufferDecoder):
    o=GameRulesChanged()
    o.GameRules={}
    count=d.read_var_uint32()
    for _ in range(count):
        name=d.read_str()
        type=d.read_var_uint32()
        if type==1:
            val=d.read_boolen()
        elif type==2:
            val=d.read_var_uint32()
        elif type==3:
            val=d.read_float32()
        o.GameRules[name]=val
    return o

# CommandRequest 77
def encode_command_request(e:BufferEncoder,i:CommandRequest):
    e.write_str(i.CommandLine)
    encode_command_origin_data(e,i.CommandOrigin)
    e.write_boolen(i.Internal)
    e.write_boolen(i.UnLimited)
    return e

# IDStructureTemplateDataRequest 132
def encode_vec3(e:BufferEncoder,i:Vec3):
    e.write_var_int32(i.x)
    e.write_var_int32(i.y)
    e.write_var_int32(i.z)
    

def encode_structure_settings(e:BufferEncoder,i:StructureSettings):
    e.write_str(i.PaletteName)
    e.write_byte(i.IgnoreEntities)
    e.write_byte(i.IgnoreBlocks)
    encode_u_block_position(e,i.Size)
    encode_u_block_position(e,i.Offset)
    e.write_var_int64(i.LastEditingPlayerUniqueID)
    e.write_byte(i.Rotation)
    e.write_byte(i.Mirror)
    e.write_float32(i.Integrity)
    e.write_uint32(i.Seed)
    encode_vec3(e,i.Pivot)
    

def encode_structure_template_data_request(e:BufferEncoder,i:StructureTemplateDataRequest):
    e.write_str(i.StructureName)
    encode_u_block_position(e,i.Position)
    encode_structure_settings(e,i.Settings)
    e.write_byte(i.RequestType)
    return e

# SettingsCommand
def encode_settings_command(e:BufferEncoder,i:SettingsCommand):
    e.write_str(i.CommandLine)
    e.write_boolen(i.SuppressOutput)
    return e

packet_encode_pool={
    IDCommandRequest:encode_command_request,
    CommandRequest:(IDCommandRequest,encode_command_request),
    IDSettingsCommand:encode_settings_command,
    SettingsCommand:(IDSettingsCommand,encode_settings_command),
    IDText:encode_text,
    Text:(IDText,encode_text),
    IDStructureTemplateDataRequest:encode_structure_template_data_request,
    StructureTemplateDataRequest:(IDStructureTemplateDataRequest,encode_structure_template_data_request),
    IDCommandBlockUpdate:encode_command_block_update,
    CommandBlockUpdate:(IDCommandBlockUpdate,encode_command_block_update)
}

packet_decode_pool={
    IDLogin:decode_login,
    IDText:decode_text,
    IDCommandOutput:decode_command_output, 
    IDMovePlayer:decode_move_player,   
    IDMobEquipment:decode_mob_equipment,
    IDSetTime:decode_set_time,
    IDGameRulesChanged:decode_gamerule_changed,
}

def decode(packet:bytes):
    '''将字节形式的收到的数据包解析为特定mc类型'''
    d=BufferDecoder(packet)
    value=d.read_var_uint32()
    packet_id = value & 0x3FF
    sender_subclient=(value >> 10) & 0x3
    target_subclient=(value >> 12) & 0x3
    decode_func=packet_decode_pool.get(packet_id)
    if decode_func is None:
        # print(f'decode func not implemented: packet type id: {packet_id}')
        # print(f'forward: fb -> packet(ID={packet_id}) -> drop')
        return packet_id,None
    else:
        return packet_id,(decode_func(d),sender_subclient,target_subclient)

def encode(packet,SenderSubClient:int=0,TargetSubClient:int=0):
    '''特定mc类型的数据包编码为字节形式'''
    type_id,encode_func=packet_encode_pool[type(packet)]
    e=BufferEncoder()
    e.write_var_uint32(type_id|(SenderSubClient<<10)|(TargetSubClient<<12))
    e=encode_func(e,packet)
    return e.bytes
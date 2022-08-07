from typing import *

def encode_echo(msg:str):
    return "echo",{"msg":msg}

async def decode_echo(data:dict,cb):
    await cb(data["msg"])


def encode_reg_mc_pkt_by_type(pktID:str):
    return "regMCPkt",{"pktID":pktID}

async def decode_reg_mc_pkt_by_type(data:dict,cb):
    await cb((data["succ"],data["err"]))

def encode_reg_any_mc_pkt():
    return "regMCPkt",{"pktID":"all"}

async def decode_reg_any_mc_pkt(data:dict,cb):
    await cb(data["succ"])
    
def encode_send_ws_cmd(cmd:str):
    return "send_ws_cmd",{"cmd":cmd}

async def decode_send_ws_cmd(data:dict,cb):
    await cb(data["result"])
    
def encode_send_player_cmd(cmd:str):
    return "send_player_cmd",{"cmd":cmd}

async def decode_send_player_cmd(data:dict,cb):
    await cb(data["result"])
    
def encode_send_wo_cmd(cmd:str):
    return "send_wo_cmd",{"cmd":cmd}

async def decode_send_wo_cmd(data:dict,cb):
    await cb(data["ack"])

def encode_get_uqholder():
    return "get_uqholder",{}

async def decode_get_uqholder(data:dict,cb):
    await cb(data)

def encode_get_players_list():
    return "get_players_list",{}

async def decode_get_players_list(data:dict,cb):
    await cb(data)

def encode_reg_menu(triggers:List[str],argument_hint:str,usage:str,sub_id:str):
    return "reg_menu",{"triggers":triggers,"argument_hint":argument_hint,"usage":usage,"sub_id":sub_id}

async def decode_reg_menu(data:dict,cb):
    await cb(data["sub_id"])

def encode_get_player_next_input(player:str,hint:str):
    return "player.next_input",{"player":player,"hint":hint}

async def decode_player_next_input(data:dict,cb):
    succ=data["success"]
    player=data["player"]
    inp=data["input"] if succ else None
    err=data["err"] if not succ else None
    await cb((succ,player,inp,err))
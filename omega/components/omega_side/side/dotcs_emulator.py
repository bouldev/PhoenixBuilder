# -*- coding: UTF-8 -*-
# Omega DotCS Emulator By 2401PT
# Omega DotCS Emulator 是为了在 Omega 中高效运行 DotCS 社区版 插件的代码连接器和框架
print("\033[32m正在启动 Omega DotCS Emulator By 2401PT\033[0m")
print("\033[32mOmega DotCS Emulator 是用于在 Omega 中高效运行 DotCS 社区版插件的代码连接器和框架\033[0m")


import os,sys 
from collections import defaultdict
from typing import *
import time 

start_time=time.time()
class PluginCode(object):
    def __init__(self,code:List[str]) -> None:
        self.typed_code:Dict[str,List[List[str]]]=defaultdict(list)
        self.partition(code)
        self.head_part=[]
        self.body_part=[]
        self.tail_part=[]
        
    def add_head_line(self,line):
        self.head_part.append(line)
        
    def add_body_line(self,line):
        self.body_part.append(line)
        
    def add_tail_line(self,line):
        self.tail_part.append(line)
        
    def partition(self,code:List[str]):
        mark="# PLUGIN TYPE: "
        frag_name="Description"
        code_frag:List[str]=[]
        blank_line=False
        for line in code:
            if line.startswith(mark):
                self.typed_code[frag_name].append(code_frag)
                frag_name=line[len(mark):].strip()
                code_frag=[]
            else:
                if line.startswith("\t"):
                    formate_line=""
                    for i,c in enumerate(line):
                        if c=="\t":
                            formate_line+="    "
                        else:
                            formate_line+=line[i:]
                    line=formate_line
                if line.strip()!="":
                    code_frag.append(line)
                    blank_line=False
                elif not blank_line and len(code_frag)>0:
                    code_frag.append("\n")
                    blank_line=True
        if len(code_frag)>0 and code_frag[-1]=="\n":
            code_frag=code_frag[:-1]
        self.typed_code[frag_name].append(code_frag)
        
    def generate_linked_code(self,rule)->Mapping[str,str]:
        for partition_type,code_frags in self.typed_code.items():
            for frag_i,code_frag in enumerate(code_frags):
                rule(partition_type)(self,frag_i,code_frag)
        return {"head":''.join(self.head_part),"body":''.join(self.body_part),"tail":''.join(self.tail_part)}
                
class Linker(object):
    def __init__(self) -> None:
        self.plugin_codes:Dict[str,PluginCode]={}
        self.auto_gen_func_count=0
        self.global_var_str="    global allplayers,all_players_dict,msgList,rev,robotname,timesErr,msgRecved,entityRuntimeID2playerName,XUID2playerName,msgLastRecvTime,itemNetworkID2NameDict,itemNetworkID2NameEngDict,needToGetMainhandItem,needToGetArmorItem,needToGetArmorItem,needToGetMainhandAndArmorItem,targetMainhandAndArmor,itemMainhandAndArmor,targetArmor,targetMainhand\n"
        
    def add_dotcs_python_file(self,code:str,file_name:str):
        self.plugin_codes[file_name] =PluginCode(code)
        
    def generate_linked_code(self)->str:
        linked_code={"head":"","body":"","tail":""}
        for file_name,plugin_code in self.plugin_codes.items():
            print("正在重新整合/连接插件: "+file_name)
            plugin_linked_code=plugin_code.generate_linked_code(self.generate_rule)
            for position,code in linked_code.items():
                frag=plugin_linked_code[position]
                linked_code[position]=code+(f"# From plugin file {file_name}\n"+frag+"\n" if frag!="" else "")
        import_part=["from  omega_side.alter.dotcs_env import *",
                     "from  omega_side.alter import dotcs_env as omega_dotcs_emulator",
                     "import os",
                     "import socket",
                     "import urllib,requests",
                     "import traceback, socket, datetime, time, json, random, sys, urllib, urllib.parse, _thread as thread, platform, sqlite3, threading, struct, hashlib, shutil, base64, ctypes, collections, types",
                     "\n"]
        import_part="\n".join(import_part)
        return import_part+"\n# HEAD PART\n"+linked_code["head"]+"\n# BODY PART\n"+"\n"+linked_code["body"]+"\n"+"\n# TAIL PART\n"+linked_code["tail"]
                
    def generate_rule(self,rule:str):
        if rule=="Description": return self.generate_description
        if rule =="def": return lambda code,frag_i,code_frag: code.add_head_line("".join(code_frag))
        if rule =="init": return lambda code,frag_i,code_frag: code.add_head_line("".join(code_frag))
        if rule=="player message":return self.generate_player_message
        if rule =="player prejoin":return self.generate_player_prejoin
        if rule =="player join": return self.generate_player_join
        if rule =="player leave": return self.generate_player_leave
        if rule =="player death": return self.generate_player_death
        if rule.startswith("repeat"):
            repeat_time=int(rule.strip().removeprefix("repeat").removesuffix("s"))
            return lambda code,frag_i,code_frag:self.generate_repeat_func(code,frag_i,code_frag,repeat_time)
        if rule.startswith("packet on another thread"):
            packet_id=int(rule.strip().removeprefix("packet on another thread").removesuffix("s"))
            return lambda code,frag_i,code_frag:self.generate_on_packet_func(code,frag_i,code_frag,packet_id)
        print(rule)
        return lambda code,frag_i,code_frag:None

    def generate_description(self,code:PluginCode,frag_i,code_frag:List[str]):
        code.add_body_line(f"# Description {frag_i}")
        for frag in code_frag:
            if frag!="Description: unknown" and frag!="\n":
                code.add_body_line("\n#"+frag)
        code.add_body_line("\n")
        
    def generate_player_message(self,code:PluginCode,frag_i,code_frag:List[str]):
        func_name=f"on_player_message_{frag_i}_{self.auto_gen_func_count}"
        self.auto_gen_func_count+=1
        code.add_body_line(f"def {func_name}(textType,playername,msg):\n"+self.global_var_str)    
        for frag in code_frag:
            code.add_body_line("    "+frag)
        code.add_tail_line(f"omega_dotcs_emulator.listen_player_message({func_name})\n")      

    def generate_player_prejoin(self,code:PluginCode,frag_i,code_frag:List[str]):
        func_name=f"on_player_prejoin_{frag_i}_{self.auto_gen_func_count}"
        self.auto_gen_func_count+=1
        code.add_body_line(f"def {func_name}(textType,playername,msg):\n"+self.global_var_str)
        for frag in code_frag:
            code.add_body_line("    "+frag)
        code.add_tail_line(f"omega_dotcs_emulator.listen_player_prejoin({func_name})\n")   

    def generate_player_join(self,code:PluginCode,frag_i,code_frag:List[str]):
        func_name=f"on_player_join_{frag_i}_{self.auto_gen_func_count}"
        self.auto_gen_func_count+=1
        code.add_body_line(f"def {func_name}(textType,playername,msg):\n"+self.global_var_str)
        for frag in code_frag:
            code.add_body_line("    "+frag)
        code.add_tail_line(f"omega_dotcs_emulator.listen_player_join({func_name})\n")   

    def generate_player_leave(self,code:PluginCode,frag_i,code_frag:List[str]):
        func_name=f"on_player_leave_{frag_i}_{self.auto_gen_func_count}"
        self.auto_gen_func_count+=1
        code.add_body_line(f"def {func_name}(textType,playername,msg):\n"+self.global_var_str)
        for frag in code_frag:
            code.add_body_line("    "+frag)
        code.add_tail_line(f"omega_dotcs_emulator.listen_player_leave({func_name})\n")   

    def generate_player_death(self,code:PluginCode,frag_i,code_frag:List[str]):
        func_name=f"on_player_death_{frag_i}_{self.auto_gen_func_count}"
        self.auto_gen_func_count+=1
        code.add_body_line(f"def {func_name}(textType,playername,msg,killer):\n"+self.global_var_str)
        for frag in code_frag:
            code.add_body_line("    "+frag)
        code.add_tail_line(f"omega_dotcs_emulator.listen_player_death({func_name})\n")   

    def generate_repeat_func(self,code:PluginCode,frag_i,code_frag:List[str],repeat_time):
        func_name=f"repeat_func_by_{repeat_time}s_{frag_i}_{self.auto_gen_func_count}"
        self.auto_gen_func_count+=1
        code.add_body_line(f"def {func_name}():\n"+self.global_var_str)
        for frag in code_frag:
            code.add_body_line("    "+frag)
        code.add_tail_line(f"omega_dotcs_emulator.repeat_exec({func_name},{repeat_time})\n")
        
    def generate_on_packet_func(self,code:PluginCode,frag_i,code_frag:List[str],packet_id):
        func_name=f"listen_packet_{packet_id}s_{frag_i}_{self.auto_gen_func_count}"
        self.auto_gen_func_count+=1
        code.add_body_line(f"def {func_name}(jsonPkt):\n"+self.global_var_str)
        for frag in code_frag:
            code.add_body_line("    "+frag)
        code.add_tail_line(f"omega_dotcs_emulator.listen_packet({func_name},{packet_id})\n")

print("\033[32m"+"开始重新整合/连接 DotCS 社区版插件的代码 为 Omega 框架兼容的 Python 代码 以提高运行速度\033[0m")
linker=Linker()
dotcs_plugins_dir="dotcs_plugins"
for file_name in os.listdir(dotcs_plugins_dir):
    if not (file_name.endswith(".py")):
        continue
    file_full_path=os.path.join(dotcs_plugins_dir,file_name)
    with open(file_full_path,"r",encoding="utf-8") as f:
        linker.add_dotcs_python_file(f.readlines(),file_name=file_name)
linked_code=linker.generate_linked_code()
with open("linked_dotcs_plugin.py","w",encoding="utf-8") as f:
    f.write("# -*- coding: UTF-8 -*-\n")
    f.write("orig_print=print\ndef alter_print(*args,**kwargs):\n\tkwargs['flush']=True\n\torig_print(*args,**kwargs)\nprint=orig_print\n")
    f.write(linked_code)
print("\033[32m"+"开始启动 DotCS 社区版插件 ( DotCS 社区版作者为 7912)\033[0m")
import linked_dotcs_plugin
print("\033[32m"+f"重新整合/连接+启动完成，共计耗时 {time.time()-start_time:.2f}s\033[0m")
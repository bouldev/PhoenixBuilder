# -*- coding: UTF-8 -*-
import os,sys 
from collections import defaultdict
from typing import *
import time 
from omega_side.python3_omega_sync import frame as omega
# Omega Plugin Loader By 2401PT
# 每个 Omega 的标准 python 插件当然都可以以一个单独进程的方式运行
# 但是把所有的 Omega Python 插件合并到一个线程中，资源消耗更少
# 但是，潜在的问题是，如果一个插件写的有问题（比如在回调中不合理的使用了sleep），可能导致所有的插件都受到影响

start_time=time.time()
print("\033[32m正在启动 Omega Plugin Loader by 2401PT\033[0m")
# 可以异步的，连接Omega框架的同时链接文件
omega.run(addr=None)

class Linker(object):
    # omega 的标准 python
    def __init__(self,plugin_dir:str) -> None:
        self.plugin_dir=plugin_dir
        self.merged_code_body:List[str]=[]
        self.merged_code_head:List[str]=[]
        self.merged_code_tail:List[str]=[]
        self.plugin_counter=0
        
    def scan_all_plugins(self):
        for file_name in os.listdir(self.plugin_dir):
            if not file_name.endswith(".py"):
                continue
            with open(os.path.join(self.plugin_dir,file_name),"r",encoding="utf-8") as f:
                code=f.readlines()
                self.add_plugin(code,file_name)

    def clean_up_codes(self,code:List[str])->List[str]:
        cleaned_codes=[]
        last_line_is_blank=False
        for line in code:
            # 有多行空行的话只空一行, 且开头不得为空行(其实开头也不可能是空行)
            line=line.rstrip()
            if "    global " in line:
                # 移除糟糕的global语法，并提示
                print(f"因为含有global声明，行{line}已经被移除，为了不把运行环境搞成垃圾场，请避免使用 global 语法")
                line=""
            if line=="":
                if last_line_is_blank and len(cleaned_codes)>0:
                    continue
                else:
                    last_line_is_blank=True
                    cleaned_codes.append("\n")
                    continue
            last_line_is_blank=False
            #   如果是 tab 表示缩进，则改用空格缩进
            if line.startswith("\t"): 
                formate_line=""
                for i,c in enumerate(line):
                    if c=="\t":
                        formate_line+="    "
                    else:
                        formate_line+=line[i:]
                line=formate_line
            cleaned_codes.append(line+"\n")
        return cleaned_codes
    
    def add_plugin(self,code:List[str],plugin_name:str):
        if len(code)==0 or not ("插件" in code[0]):
            return
        enable_flag=False
        for enable_mark in ("on","ON","On","开","启用","Kai","kai","KAI","qi","QI","Qi"):
            if enable_mark in code[0]:
                enable_flag=True
                break
        if not enable_flag:
            return
        print(f"正在合并: {plugin_name}")
        if self.plugin_counter>0:
            self.merged_code_body.append("\n")
        self.plugin_counter+=1
        self.merged_code_body.append(f"# From Plugin: {plugin_name}\n")
        code=self.clean_up_codes(code)
        # 为每个插件文件中的代码都创建独立的运行条件
        self.merged_code_body.append(f"def plugin_file_{self.plugin_counter}():\n")
        for l in code:
            if "import" in l and "*" in l:
                self.merged_code_head.append(l.strip()+"\n")
            else:
                self.merged_code_body.append("    "+l)
        self.merged_code_tail.append(f"plugin_file_{self.plugin_counter}()\n")
    
    def dump_code(self)->str:
        code="".join(self.merged_code_head+["\n"]+self.merged_code_body+["\n"]+self.merged_code_tail)
        return code

print("\033[32m正在将插件链接到单一进程中\033[0m")
linker=Linker("omega_python_plugins")
linker.scan_all_plugins()
with open("linked_python_plugin.py","w",encoding="utf-8") as f:
    f.write("# -*- coding: UTF-8 -*-\n")
    f.write("orig_print=print\ndef alter_print(*args,**kwargs):\n\tkwargs['flush']=True\n\torig_print(*args,**kwargs)\nprint=orig_print\n\n")
    f.write(linker.dump_code())
    
import linked_python_plugin
print("\033[32m"+f"标准 Omega Python 插件启动完成，共计耗时 {time.time()-start_time:.2f}s\033[0m")
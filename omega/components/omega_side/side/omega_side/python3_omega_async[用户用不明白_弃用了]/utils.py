import asyncio
import shutil
import sys
import os
import subprocess
import typing

def get_python_exec()-> str:
    python_exec=sys.executable
    return python_exec

def run_cmd_sync(cmd:typing.List[str]):
    p = subprocess.Popen(cmd, shell=False, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,env=os.environ)
    while p.poll() is None:
        line = p.stdout.readline()
        line = line.strip()
        if line:
            print("\t",line.decode(encoding='utf-8'))  
    return p.returncode == 0

def install_lib(lib_name:str,lib_install_name:str):
    target_dir=os.path.join(os.getcwd(),"python_lib")
    os.makedirs(target_dir,exist_ok=True)
    sys.path.append(target_dir) 
    import importlib
    try:
        module=importlib.import_module(lib_name)
        # print(f"库 {lib_name} 已安装 {module}")
        return True
    except Exception as e:
        # print(e)
        pass 
    print(f"开始安装库: {lib_name}")
    mirror_site="https://mirrors.bfsu.edu.cn/pypi/web/simple"
    if run_cmd_sync([get_python_exec(),"-m","pip","install","-i",mirror_site,f"--target={target_dir}",lib_install_name]):
        print(f"库 {lib_name} 安装成功")
    else:
        raise Exception(f"库 {lib_name} 安装失败")

async def sleep(time):
    await asyncio.sleep(time)
    
async def run_multiple_tasks_parallel(tasks:typing.List[typing.Awaitable]):
    await asyncio.gather(*tasks)

# 有时候，不加 flush 可能导致print不出来，所以我们要重新定义这个print
std_print=print
def flush_print(*args,**kwargs):
    kwargs["flush"]=True
    std_print(*args,**kwargs)
print=flush_print

def run_in_background(task:typing.Awaitable) ->typing.Awaitable:
    return asyncio.create_task(task)
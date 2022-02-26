import os
import shutil
current_folder="fyne-gui"
tmp_workspace="tmp_workspace"

assert os.path.split(os.getcwd())[-1] == current_folder, f"should run under {current_folder}"

if os.path.isdir(tmp_workspace):
    shutil.rmtree(tmp_workspace)
os.makedirs(tmp_workspace, exist_ok=True)
for p in os.listdir("."):
    # 逐个将所有的文件复制到tmp_workspace目录下(因为编译工具的缺陷)
    if p.startswith(".") or p==tmp_workspace:
        continue
    if os.path.isdir(p):
        shutil.copytree(p, os.path.join(tmp_workspace, p))
    else:
        shutil.copyfile(p, os.path.join(tmp_workspace, p))
tmp_fastbuilder=os.path.join(tmp_workspace,"fastbuilder")
os.makedirs(tmp_fastbuilder, exist_ok=False)
for p in os.listdir(".."):
    # 逐个将所有的文件复制到tmp_workspace目录下(因为编译工具的缺陷)
    if p.startswith(".") or p==current_folder:
        continue
    fullPath=os.path.join("..", p)
    if os.path.isdir(fullPath):
        shutil.copytree(fullPath, os.path.join(tmp_fastbuilder, p))
    else:
        shutil.copyfile(fullPath, os.path.join(tmp_fastbuilder, p))
data=""
with open(os.path.join(tmp_workspace,"go.mod"),"r") as f:
    data=f.read()
    data=data.replace("replace phoenixbuilder => ../","replace phoenixbuilder => ./fastbuilder")
with open(os.path.join(tmp_workspace,"go.mod"),"w") as f:
    f.write(data)
for root, dirs, files in os.walk(tmp_workspace):
    for f in files:
        if f.endswith(".c") or f.endswith(".h"):
            os.remove(os.path.join(root, f))
        if f.endswith(".go"):
            rewrite=""
            remove=False
            with open(os.path.join(root, f),"r") as fp:
                data=fp.readline()
                if "!fyne_gui" in data:
                    remove=True
                if "fyne_gui" in data:
                    rewrite="\n".join(fp.readlines())
            if remove:
                os.remove(os.path.join(root, f))
            elif rewrite!="":
                with open(os.path.join(root, f),"w") as fp:
                    fp.write(rewrite)
                
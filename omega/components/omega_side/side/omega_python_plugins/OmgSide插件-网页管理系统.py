# 插件: 关

import os,sys
from omega_side.python3_omega_sync.bootstrap import install_lib
if "Windows" in sys.platform:
    os.system("cd omega_storage/side/interpreters/python/bin/ && python.exe -m pip config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple")
else:
    os.system("cd omega_storage/side/interpreters/python/bin/ && python -m pip config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple")
install_lib("flask",mirror_site = "https://pypi.tuna.tsinghua.edu.cn/simple");install_lib("requests",mirror_site = "https://pypi.tuna.tsinghua.edu.cn/simple")
from omega_side.python3_omega_sync import API
from omega_side.python3_omega_sync import frame as omega
from omega_side.python3_omega_sync.protocol import *
import json,requests
from flask import Flask,render_template


def plugin_main(api:API):
    def plugin_web():
        def plugin_list_启用():
            list = []
            lst=os.listdir("./omega_python_plugins")
            for filename in lst:
                if filename.endswith('.py'):
                    with open("./omega_python_plugins/"+filename, 'r', encoding='utf-8') as f:
                        lines = f.readlines()
                        first = lines[0]
                    if first[6] == "开":
                        list.append(filename)
            else:
                return list
        def plugin_list_禁用():
            try:
                list = []
                lst=os.listdir("./omega_python_plugins")
                for filename in lst:
                    if filename.endswith('.py'):
                        with open("./omega_python_plugins/"+filename, 'r', encoding='utf-8') as f:
                            lines = f.readlines()
                            first = lines[0]
                        if first[6] == "关":
                            list.append(filename)
                else:
                    return list
            except:
                pass
        app = Flask(__name__,template_folder="./omega_python_plugins/OmgSide插件-网页管理系统HTML")
        token = "xEF5GZvAlhToXo6WM91pe8K8WYVG9GwvYIZs5VmIMsM5D8vZa1"
        # 默认令牌 xEF5GZvAlhToXo6WM91pe8K8WYVG9GwvYIZs5VmIMsM5D8vZa1
        # 解决浏览器输出乱码问题
        app.config['JSON_AS_ASCII'] = False
        @app.route('/omage/login')
        def login():
            return render_template('login.html')
        @app.route('/omage/main')
        def main():
            return render_template('main.html',token=token,plugin1=plugin_list_启用(),plugin2=plugin_list_禁用())
        @app.route('/omage/command')
        def command():
            return render_template('command.html',token=token)
        @app.route("/omage/api/login/<tk>")
        def api_login(tk):
            if tk == token:
                return 'True'
            else:
                return 'False'
        @app.route('/omage/exit')
        def exit():
            return '''
        <!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
        </head>    
        <body onload="et()">
            <script>
                function et(){
                    sessionStorage.removeItem('token')
                    window.location.href = './login';
                }
            </script>
        </body>
            '''
        @app.route("/omage/api/commandstart/<command>")
        def commandstart(command):
            response=api.do_send_ws_cmd(command,cb=None)
            return response
        app.run(host='127.0.0.1',port=5000)
    plugin_web()
omega.add_plugin(plugin=plugin_main)
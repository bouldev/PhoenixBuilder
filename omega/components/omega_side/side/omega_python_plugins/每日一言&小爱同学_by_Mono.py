# 插件: 关



from threading import Thread
from omega_side.python3_omega_sync import API
from omega_side.python3_omega_sync import frame as omega
from omega_side.python3_omega_sync.bootstrap import install_lib
from omega_side.python3_omega_sync.protocol import *
import requests
class version_api:
    author     =    "Mono"
    version    =    "2.0 for omegaside"
def Mono_plugin_requests(api:API):
    def word(player_input:PlayerInput):
        player=player_input.Name
        url = "https://v1.hitokoto.cn"
        response = requests.get(url)
        data = response.json()
        list_values = [i for i in data.values()]
        if "hitokoto" in data:
            api.do_send_ws_cmd(f"/playsound note.harp {player}",cb=None)
            Everydaysay=list_values[2]
            api.do_send_player_msg(player,"§a每日一言§e>> §b%s"% Everydaysay ,cb=None)
            api.do_send_player_msg(player,'§a来自§e%s§b-§e%s§r'%(list_values[4],list_values[5]),cb=None)
    def _xiaoaichat(chat,target)->None:
        msg = chat.replace('#', '').replace('＃', '').replace('&', '').replace('＆', '')
        try:#密钥获取方式:https://apibug.cn/doc/xiaoai.html
            page = requests.get(f'https://apibug.cn/api/xiaoai/?msg={msg}&apiKey=这里输入你的api密钥')
        except Exception as err:
            api.do_send_player_msg(target, '§c获取错误.',cb=None)
        result = eval(str(page.text))
        if result['code'] == 200: #这里请根据你的api的返回格式来判断获取状态，自行更改
            returns = result['text']
            if returns != '':
                api.do_send_player_msg(target,f'<§l§b小爱同学§r> §6@{target}§r\n§l§e{returns}',cb=None)
            else:
                api.do_send_player_msg(target,f'<§l§b小爱同学§r> §6@{target}§r\n§l§e你在说什么，我好像不明白。',cb=None)
        else:
            api.do_send_player_msg(target, '§l§4ERROR§r>> §c获取聊天失败.',cb=None)
    def xiaoai(player_input:PlayerInput):
        Thread(target=_xiaoaichat,args=["".join(player_input.Msg),player_input.Name]).start()
    api.listen_omega_menu(triggers=["word","一言"],argument_hint="",usage="获取每日一言",cb=None,on_menu_invoked=word)
    api.listen_omega_menu(triggers=["小爱","xiaoai","xa"],argument_hint="",usage="和小爱聊天",cb=None,on_menu_invoked=xiaoai)
omega.add_plugin(plugin=Mono_plugin_requests)

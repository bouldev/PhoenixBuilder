# 插件: 关
# 需要使用的请把这个"关"改为"开"
from omega_side.python3_omega_sync import API
from omega_side.python3_omega_sync import frame as omega
from omega_side.python3_omega_sync.bootstrap import install_lib
from omega_side.python3_omega_sync.protocol import *
import smtplib
from email.mime.text import MIMEText
from email.header import Header
import requests
class version_mail:
    author     =    "Mono"
    version    =    "2.0 for omegaside"

def sendmail(playermsg:str,playername:str) -> bool:
    try:
        smtp                    = smtplib.SMTP_SSL("smtp.126.com")
        smtp.login(user         = "你的邮箱号", password = "密码/授权码")
        message                 = MIMEText(playermsg, 'plain', 'utf-8')
        message['From']         = Header("你的邮箱地址", 'utf-8')  # 发件人的昵称
        message['To']           = Header("对方邮箱地址", 'utf-8')  # 收件人的昵称
        message['Subject']      = Header(f'来自{playername}的反馈', 'utf-8')  # 定义主题内容
        smtp.sendmail(from_addr = "你的邮箱地址", to_addrs = "对方邮箱地址", msg = message.as_string())
        smtplib.SMTP_SSL("smtp.126.com").quit()
        return True
    except:
        return False
def Mono_plugin_mail(api:API):
    def mailMenu(player_input:PlayerInput):
        player=player_input.Name
        wait_input = api.do_get_get_player_next_param_input(player,hint=f"§7{'-'*20}\n§f> 请输入发送的内容\n> 长度限制:50字以内\n§7{'-'*20}",cb=None)
        wait_text  = "".join(wait_input.input)
        if len(wait_text) <= 50 and wait_text != "***":
            if 1:
                wait_input_1 = api.do_get_get_player_next_param_input(player,hint="§f> 确认发送输入1 取消输入0",cb=None)
                wait_text_1  = "".join(wait_input_1.input)
                if wait_text_1 in ["1","y","Y","yes","Yes"]:
                    api.do_send_player_msg(player,"§f> §a邮件正在发送.",cb=None)
                    result = sendmail(wait_text,player)
                    
                    if result == True:
                        api.do_send_player_msg(player,"§f> §a邮件发送成功.",cb=None)
                    else:
                        api.do_send_player_msg(player,"§f> §c邮件发送失败.",cb=None)
                else:
                    api.do_send_player_msg(player,"§f> §c邮件取消发送.",cb=None)
        elif len(wait_text) > 50:
            api.do_send_player_msg(player,"§f> §c字数过长.",cb=None)
        else:
            api.do_send_player_msg(player,"§f> §c取消发送.",cb=None)
    api.listen_omega_menu(triggers=["mail","email"],argument_hint="",usage="发送反馈邮件",cb=None,on_menu_invoked=mailMenu)
omega.add_plugin(plugin=Mono_plugin_mail)

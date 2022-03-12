import proxy
from proxy import forward

# 最简单的例子，仅解析收到的数据，看看游戏里发生了什么

conn=forward.connect_to_fb_transfer(host="localhost",port=8000)
sender=forward.Sender(connection=conn)
receiver=forward.Receiver(connection=conn)

while True:
    bytes_msg,(packet_id,decoded_msg)=receiver()
    if decoded_msg is None:
        # 还未实现该类型数据的解析(会有很多很多的数据包！)
        # print(f'unkown decode packet ({packet_id}): ',bytes_msg)
        continue
    else:
        # 已经实现类型数据的解析
        packet_id,msg,sender_subclient,target_subclient=decoded_msg
        print(msg)
import uuid
import proxy
from proxy import forward
from proxy import utils
from threading import Thread
from queue import Queue
import time

# 有两个子线程，一个负责停的解析数据，并通过 Queue 将解析结果在线程之间传递
# 另一个子线程的目仅仅负责发送指令
# 主线程每隔一段时间检查连接状态

conn = forward.connect_to_fb_transfer(host="localhost", port=8000)
sender = forward.Sender(connection=conn)
receiver = forward.Receiver(connection=conn)

statues = {
    'recv_thread_alive': True,
    'working_thread_alive': True
}

recv_queue = Queue()


def recv_thread_func(receiver: forward.Receiver, recv_queue: Queue):
    try:
        while True:
            bytes_msg, (packet_id, decoded_msg) = receiver()
            if decoded_msg is None:
                # 还未实现该类型数据的解析(会有很多很多的数据包！)
                # print(f'unkown decode packet ({packet_id}): ',bytes_msg)
                continue
            else:
                # 已经实现类型数据的解析
                msg, sender_subclient, target_subclient = decoded_msg
                print(msg)
                recv_queue.put(msg)
    except Exception as e:
        print('Recv thread terminated!')
        statues['recv_thread_alive'] = False
        raise e


recv_thread = Thread(target=recv_thread_func, args=(receiver, recv_queue))


def working_thread_func(sender: forward.Sender, recv_queue: Queue):
    try:
        while True:
            command = input('cmd:')
            msg, uuid_bytes = utils.pack_ws_command(command, uuid=None)
            sender(msg)
            time.sleep(0.1)
    except Exception as e:
        print('Working thread terminated!')
        statues['working_thread_alive'] = False
        raise e


work_thread = Thread(target=working_thread_func, args=(sender, recv_queue))

recv_thread.daemon = True
recv_thread.start()
work_thread.daemon = True
work_thread.start()

while True:
    time.sleep(0.1)
    if False in statues.values():
        print('sub process crashed! programme terminating...')
        exit(-1)

import time
import random
from CodeWar.Utils import CodeWar

# 注意 A B 项 添一项即可, C 项必填

# 创建房间需要设置 (A)
playernum = 2 # 玩家人数
row = 20 # 最大不超过100
col = 20 # 最大不超过100
barback = 10 # 兵营
portal  = 20 # 据点
barrier = 30 # 障碍物

# 加入房间需要设置 (B)
roomtoken = '' 

# 账号及服务器设置 (C)
chrome = '' # 对于Windows用户可能需要填写Chrome安装路径.../chrome.exe
username = 'test'
password = 'test'
email = 'test'
url = '127.0.0.1'
port = '52333'


if __name__ == '__main__':
    # 以下不需要修改
    my = CodeWar(url=url, port=port, 
        username=username, password=password, email=email, 
        chrome=chrome,
        roomtoken=roomtoken, 
        playernum=playernum, 
        row=row, col=col, 
        barback=barback, portal=portal, barrier=barrier)
    # 以上不需要修改
    
    # a.register() # 注册
    while not my.run():
        time.sleep(3)

    # 获取初始位置(基地)
    a.isStart()
    x, y = a.getBase()
    # 获取地图大小
    row, col = a.getMapSize()

    while True:
        # 游戏已结束
        if not a.isStart():
            a.getScoreBoard()
            a.leave()
            break
        
        ### 以下为策略主体部分 ###
        '''
            your coding
        '''
        # Params: (x, y)
        # 不传参数默认 x = base.x  y = base.y
        # Return: (query_res, query_status)
        # query_res: {
        #     m1: [][],
        #     m2: [][]
        # }
        # query_status: 0 => 失败  1 => 成功
        query_res, query_status = a.query(x, y)
        time.sleep(1)
        # Params: (x, y, radio, direction)
        # radio:
        #   1 -> all 调动 (x, y)  所有兵力
        #   2 -> half 调动 (x, y)  1/2的兵力
        #   3 -> quarter 调动 (x, y)  1/4的兵力
        # direction:
        #   1 -> (x, y) => (x - 1, y)
        #   2 -> (x, y) => (x, y + 1)
        #   3 -> (x, y) => (x + 1, y)
        #   4 -> (x, y) => (x, y - 1)
        # 
        # Return: (action_length, move_status)
        # action_length 操作队列长度
        # move_status: 0 => 失败  1 => 成功
        print('---[moveTo]',x ,y, a.id)
        direction = random.randint(1,4)
        radio = random.randint(1,3)
        action_length, move_status = a.move(x,y,radio,direction)
        time.sleep(1)
        ### 以上为策略猪蹄部分 ###

    a.logout()
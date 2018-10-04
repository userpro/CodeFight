import time
import random
from CodeWar.Utils import CodeWar

playernum = 1 # 玩家人数(创建房间需要)
row = 30 # (创建房间需要 最大不超过100)
col = 30 # (创建房间需要 最大不超过100)
roomtoken = '' # 加入房间需要

chrome = '' # 对于Windows用户可能需要填写Chrome安装路径.../chrome.exe
username = 'hipro'
password = 'okiamhi'
email = 'test@test'
url = '127.0.0.1'
port = '52333'


if __name__ == '__main__':
    a = CodeWar(url, port, username, password, email, chrome)
    # a.register() # 注册
    a.run(roomtoken=roomtoken, playernum=playernum, row=row, col=col)

    # 游戏主循环 move操作延时不要低于100ms 操作会累积
    # 获取初始位置(基地)
    a.isStart()
    x, y = a.getBase()
    # 获取地图大小
    row, col = a.getMapSize()
    print(row, col)

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
        dir = random.randint(1,4)
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
        action_length, move_status = a.move(x,y,1,dir)
        if dir == 1: x = x - 1
        elif dir == 2: y = y + 1
        elif dir == 3: x = x + 1
        elif dir == 4: y = y - 1
        if x <= 0: 
            x = 0
        if x > row: 
            x = row - 1
        if y <= 0: 
            y = 0
        if y > col: 
            y = col - 1
        time.sleep(1)
        ### 以上为策略猪蹄部分 ###
    
    a.logout()



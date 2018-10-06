import time
import random
from CodeWar.Utils import CodeWar

playernum = 1 # 玩家人数(创建房间需要)
row = 10 # (创建房间需要 最大不超过100)
col = 10 # (创建房间需要 最大不超过100)
roomtoken = '' # 加入房间需要

chrome = '' # 对于Windows用户可能需要填写Chrome安装路径.../chrome.exe
username = 'hipro'
password = 'okiamhi'
email = 'test@test'
url = '127.0.0.1'
port = '52333'


if __name__ == '__main__':
    my = CodeWar(url, port, username, password, email, chrome)
    # a.register() # 注册
    while not my.run(roomtoken=roomtoken, playernum=playernum, row=row, col=col):
        time.sleep(3)

    # 游戏主循环 move操作延时不要低于100ms 操作会累积
    # 获取初始位置(基地)
    my.isStart()
    x, y = my.getBase()
    # 获取地图大小
    row, col = my.getMapSize()
    print(row, col)

    direction = 1
    MapUnit = my.MapUnit
    while True:
        # 游戏已结束
        if not my.isStart():
            my.getScoreBoard()
            my.leave()
            break
        
        ### 以下为策略主体部分 ###
        '''
            your coding
        '''
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
        print('---[moveTo]',x ,y, direction)
        action_length, move_status = my.move(x,y,1,direction)
        # 操作序列过长 
        if action_length > 3:
            time.sleep(3)

        time.sleep(0.5)
        # Params: (x, y)
        # 不传参数默认 x = base.x  y = base.y
        # Return: (query_res, query_status)
        # query_res: {
        #     m1: [5][5],
        #     m2: [5][5]
        # }
        # 视野为以(x, y)为中心的5*5矩阵
        # query_status: 0 => 失败  1 => 成功
        query_res, query_status = my.query(x, y)
        
        if query_status == 1:
            tmpx = tmpy = 0
            mx_army = 0 # 兵力最多
            m1 = query_res['m1']
            m2 = query_res['m2']
            for i in range(5):
                for j in range(5):
                    if m1[i][j] != -1 and MapUnit.getUserId(m2[i][j]) == my.id:
                        # 调用兵力最多的 Cell
                        if mx_army < m1[i][j]:
                            mx_army = m1[i][j]
                            tmpx = i
                            tmpy = j

            # 找一个不属于自己的方向
            if tmpx-1>=0 and m1[tmpx-1][tmpy]!=-1 and not MapUnit.isBarrier(m2[tmpx-1][tmpy]) and MapUnit.getUserId(m2[tmpx-1][tmpy]) != my.id:
                direction = 1
            elif tmpy+1<5 and m1[tmpx][tmpy+1]!=-1 and not MapUnit.isBarrier(m2[tmpx][tmpy+1]) and MapUnit.getUserId(m2[tmpx][tmpy+1]) != my.id:
                direction = 2
            elif tmpx+1<5 and m1[tmpx+1][tmpy]!=-1 and not MapUnit.isBarrier(m2[tmpx+1][tmpy]) and MapUnit.getUserId(m2[tmpx+1][tmpy]) != my.id:
                direction = 3
            elif tmpy-1>=0 and m1[tmpx][tmpy-1]!=-1 and not MapUnit.isBarrier(m2[tmpx][tmpy-1]) and MapUnit.getUserId(m2[tmpx][tmpy-1]) != my.id:
                direction = 4
            else:
                direction = random.randint(1,4)

            x = tmpx + x - 2
            y = tmpy + y - 2


        time.sleep(1)
        ## 以上为策略猪蹄部分 ###
    
    my.logout()



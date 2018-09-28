import time
import random
import webbrowser
from CodeWar.Utils import CodeWar

playernum = 1 # 玩家人数(创建房间需要)
row = random.randint(30,90) # (创建房间需要)
col = random.randint(30,90) # (创建房间需要)
roomtoken = '' # 加入房间需要

chromePath = '' # 对于Windows用户可能需要填写Chrome安装路径.../chrome.exe
username = 'hipro'
password = 'okiamhi'
email = 'test@test'
url = '127.0.0.1'
port = '8080'


def view():
    # 浏览器打开view页面 <测试用>
    openUrl = a.url+'/view/'+a.roomtoken
    if chromePath != '':
        webbrowser.register('chrome', None, webbrowser.BackgroundBrowser(chromePath))

    try:
        webbrowser.get('chrome').open(openUrl, new=0, autoraise=True)
    except Exception as e:
        webbrowser.open(openUrl, new=0, autoraise=True)


if __name__ == '__main__':
    a = CodeWar(url, port, username, password, email)
    # a.register() # 注册
    a.login()
    
    if roomtoken != "":
        # 加入房间
        a.join(roomtoken=roomtoken)
    else:
        # 创建房间
        a.join(playernum=playernum, row=row, col=col)
    
    # 检测游戏是否开始
    while not a.isStart():
        time.sleep(3)

    view() # 展示web页面 (非必须)

    # 游戏主循环 操作延时不要低于100ms 操作会累积
    # 获取初始位置(基地)
    x, y = a.getBase()    
    dir = 1
    while True:
        # 游戏已结束
        if not a.isStart():
            a.getScoreBoard()
            a.leave()
            break
        # 策略
        res = a.query()
        time.sleep(3)
        if dir == 4: 
            dir = 1
        else: 
            dir = dir + 1
        a.move(x,y,1,dir)
        time.sleep(2)
        res = a.query()
        # a.view()
        time.sleep(2)
    
    a.logout()


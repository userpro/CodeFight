import urllib.parse
import requests
import json
import time
import random

class CodeWar(object):
    """docstring for CodeWar"""
    def __init__(self, url, port, username, password, email):
        super(CodeWar, self).__init__()
        self.url = 'http://' + str(url) + ':' + str(port)
        self.username = username
        self.password = password
        self.email = email
        self.usertoken = self.roomtoken = ''
        self.x = self.y = -1
        self.row = self.col = 0

    # --* 注册 *--
    def register(self):
        data = {
            'username': self.username,
            'password': self.password,
            'email': self.email
        }
        payload = urllib.parse.urlencode(data
            )
        _register = requests.post(self.url + "/user", params=payload).json()
        
        print('[register] ', _register)
        if _register['status'] != 1:
            print('<Error>[register] ', _register['message'])

    # --* 登录 *--
    def login(self):
        _login = requests.get(self.url + "/user?username=" + self.username + "&password=" + self.password).json()

        print('[login] ', _login)
        if _login['status'] == 1:
            self.usertoken = _login['usertoken']
        elif _login['status'] == 2:
            print('[login] ', _login['message'])
        else:
            print('<Error>[login] ', _login['message'])

    # --* 加入或者创建房间 *--
    def join(self, roomtoken='', playnum=1, row=30, col=30):
        data = {
            'usertoken': self.usertoken,
            'playnum': 1,
            'row': row,
            'col': col
        }
        if roomtoken != '': 
            data['roomtoken'] = roomtoken
            del data['row']
            del data['col']

        payload = urllib.parse.urlencode(data)

        _join = requests.post(self.url + "/room", params=payload).json()
        
        print('[join] ', _join)
        if _join['status'] == 1:
            self.roomtoken = _join['roomtoken']
            self.row = _join['row']
            self.col = _join['col']
        else:
            print('<Error>[join] ', _join['message'])

    # --* 获取游戏是否开始以及基地坐标 *--
    def isStart(self):
        _isStart = requests.get(self.url + "/room/start?usertoken=" + self.usertoken + "&roomtoken=" + self.roomtoken).json()
        
        print('[isStart] ', _isStart)
        if _isStart['status'] == 1:
            self.x = _isStart['x']
            self.y = _isStart['y']
            return True

        return False

    # --* 返回基地坐标 *--
    def getBase(self):
        return (self.x, self.y)

    # --* 移动 *--
    def move(self, x, y, radio, direction):
        payload = {
            'usertoken': self.usertoken,
            'roomtoken': self.roomtoken,
            'radio': radio,
            'direction': direction,
            'loc': {
                'x': x, 'y':y
            }
        }
        headers = {'Content-Type': 'application/json'}
        _move = requests.put(self.url + "/room", headers=headers, data=json.dumps(payload)).json()

        print('[move] ', _move)


    # --* 查询 *--
    def query(self, x=-1, y=-1):
        if x == -1:
            x = self.x
        if y == -1:
            y = self.y
        payload = {
            'usertoken': self.usertoken,
            'roomtoken': self.roomtoken,
            'loc': {
                'x': x,
                'y': y
            }
        }
        headers = {'Content-Type': 'application/json'}
        _query = requests.get(self.url + "/room", headers=headers, data=json.dumps(payload)).json()
        
        print('[query] ', _query)
        if _query['status'] != 1:
            print('<Error>[_query] ', _query['message'])
        else:
            return _query['eyeshot']

    # --* 查询用户状态(是否游戏 上一局输赢) *--
    def getStatus(self):
        _status = requests.get(self.url + "/user/status?usertoken=" + self.usertoken).json()

        print('[getStatus] ', _status)
        if _status['status'] != 1:
            print('<Error>[getStatus]')
        else:
            return _status['userstatus']

    # --* 退出游戏 *--
    def logout(self, usertoken=''):
        if usertoken == '':
            usertoken = self.usertoken
        _logout = requests.delete(self.url + "/user?usertoken="+usertoken)

    def view(self):
        _view = requests.get(self.url+"/view/"+self.roomtoken).json()
        print('[view] ', _view)


if __name__ == '__main__':
    username = 'hipro'
    password = 'okiamhi'
    email = 'test@test'
    url = '127.0.0.1'
    port = '8080'
    # username1 = 'test'
    # password1 = 'test'
    # email1 = 'test'

    a = CodeWar(url, port, username, password, email)
    # a.register()
    a.login()
    a.join(playnum=1, row=random.randint(10,30), col=random.randint(10,30))
    # 检测游戏是否开始
    while not a.isStart():
        time.sleep(1)

    # 获取初始位置(基地)
    x, y = a.getBase()
    dir = 1
    while True:
        # 游戏已结束
        if not a.isStart():
            print(a.getStatus())
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


import urllib.parse
import requests
import json

class CodeWar(object):
    """docstring for War"""
    def __init__(self, url, port, username, password, email):
        super(CodeWar, self).__init__()
        self.url = 'http://' + str(url) + ':' + str(port)
        self.username = username
        self.password = password
        self.email = email
        self.playernum = 0
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
        _register = requests.post(self.url + "/user", params=payload, timeout=5).json()
        
        print('[register] ', _register)
        if _register['status'] != 1:
            print('<Error>[register] ', _register['message'])
        # 返回 1 => 注册成功
        return _register['status']

    # --* 登录 *--
    def login(self):
        _login = requests.get(self.url + "/user?username=" + self.username + "&password=" + self.password, timeout=5).json()

        print('[login] ', _login)
        if _login['status'] == 1:
            self.usertoken = _login['usertoken']
        elif _login['status'] == 2:
            self.usertoken = _login['usertoken']
            self.roomtoken = _login['roomtoken']
        else:
            print('<Error>[login] ', _login['message'])
        # 返回状态 1=>登录成功  2=>断线重连成功
        return _login['status']

    # --* 加入或者创建房间 *--
    def join(self, roomtoken='', playernum=1, row=30, col=30):
        data = {
            'usertoken': self.usertoken,
            'playernum': playernum,
            'row': row,
            'col': col
        }
        if roomtoken != '': 
            self.roomtoken = roomtoken
            data['roomtoken'] = roomtoken
            del data['row']
            del data['col']
            del data['playernum']

        payload = urllib.parse.urlencode(data)

        _join = requests.post(self.url + "/room", params=payload, timeout=5).json()
        
        print('[join] ', _join)
        if _join['status'] == 1:
            self.roomtoken = _join['roomtoken']
            self.row = _join['row']
            self.col = _join['col']
            self.playernum = _join['playernum']
        else:
            print('<Error>[join] ', _join['message'])

        # 返回 1 => join 成功
        return _join['status']

    # --* 获取游戏是否开始以及基地坐标 *--
    def isStart(self):
        _isStart = requests.get(self.url + "/room/start?usertoken=" + self.usertoken + "&roomtoken=" + self.roomtoken, timeout=5).json()
        
        print('[isStart] ', _isStart)
        if _isStart['status'] == 1:
            self.x = _isStart['x']
            self.y = _isStart['y']
            
        # 返回 1 => start  0 => not start
        return _isStart['status']

    # --* 返回基地坐标 *--
    def getBase(self):
        return (self.x, self.y)

    # --* 返回地图大小 *--
    def getMapSize(self):
        return (self.row, self.col)

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
        _move = requests.put(self.url + "/room", headers=headers, data=json.dumps(payload), timeout=5).json()

        print('[move] ', _move)
        # 返回 1 => 成功 
        # PS: 这个返回值没啥用
        return _move['status']


    # --* 查询 *--
    def query(self, x=-1, y=-1):
        if x == -1: x = self.x
        if y == -1: y = self.y
        
        payload = {
            'usertoken': self.usertoken,
            'roomtoken': self.roomtoken,
            'loc': {
                'x': x,
                'y': y
            }
        }

        headers = {'Content-Type': 'application/json'}
        _query = requests.get(self.url + "/room", headers=headers, data=json.dumps(payload), timeout=5).json()
        
        print('[query] ', _query)

        # 返回值 (result, status)
        # status 1 => 成功
        if _query['status'] != 1:
            print('<Error>[_query] ', _query['message'])
            return ({}, _query['status'])
        
        return (_query['eyeshot'], _query['status'])

    # --* 查询得分状态(少一点 耗性能) *--
    def getScoreBoard(self):
        _scoreboard = requests.get(self.url + "/room/scoreboard?roomtoken=" + self.roomtoken, timeout=5).json()

        print('[getScoreBoard] ', _scoreboard)
        
        # 返回值 (result, status)
        # status 1 => 成功
        if _scoreboard['status'] != 1:
            print('<Error>[getScoreBoard]')
            return ({}, _scoreboard['status'])
        
        return (_scoreboard['scoreboard'], _scoreboard['status'])

    # --* 离开房间 *--
    def leave(self):
        _leave = requests.delete(self.url + "/room?usertoken=" + self.usertoken + "&roomtoken=" + self.roomtoken, timeout=5).json()
        self.roomtoken = ''

        print('[leave]', _leave)

    # --* 退出游戏 *--
    def logout(self, usertoken=''):
        if usertoken == '':
            usertoken = self.usertoken
        _logout = requests.delete(self.url + "/user?usertoken="+usertoken, timeout=5)
        self.usertoken = ''

        print('[logout]')




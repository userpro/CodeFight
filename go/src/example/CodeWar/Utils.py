import urllib.parse
import requests
import json
import time
import webbrowser

class MapUnit(object):
    """解析 Cell 工具类 用于解析地图的 m2 图层"""
    def __init__(self):
        super(MapUnit, self).__init__()
        # 地形 Cell Type
        self.SPACE   = 0x00 # 000- ----
        self.BASE    = 0x20 # 001- ----
        self.BARBACK = 0x40 # 010- ----
        self.PORTAL  = 0x60 # 011- ----
        self.BARRIER = 0x80 # 100- ----
        # 阵营 User Id
        self.SYSTEM  = 0x00 # ---- -000
        self.VISITOR = 0x01 # ---- -001
        # mask
        self.USER_MASK= 0x1f # 0001 1111
        self.TYPE_MASK= 0xe0 # 1110 0000

    def isSpace(self, cell):
        return cell & self.TYPE_MASK == self.SPACE

    def isBase(self, cell):
        return cell & self.TYPE_MASK == self.BASE 

    def isBarback(self, cell):
        return cell & self.TYPE_MASK == self.BARBACK 

    def isPortal(self, cell):
        return cell & self.TYPE_MASK == self.PORTAL 

    def isBarrier(self, cell):
        return cell & self.TYPE_MASK == self.BARRIER 

    def isSystem(self, cell):
        return cell & self.USER_MASK == self.SYSTEM 

    # --* 设置 Cell 类型 *--
    def setCellType(self, cell, typ):
        return cell & self.USER_MASK | typ 

    # --* 设置 Cell 所有者id *--
    def setCellId(self, cell, uid):
        return cell & TYPE_MASK | uid 

    # --* 获取 Cell 所有者id *--
    def getUserId(self, cell):
        return cell & self.USER_MASK 


class CodeWar(object):
    """玩家操作类"""
    def __init__(self, url, port, username, password, email, chrome='', roomtoken='', playernum=1, row=30, col=30, barback=10, portal=20, barrier=30):
        super(CodeWar, self).__init__()
        self.MapUnit = MapUnit()
        self.chrome = chrome # chrome path

        self.url = 'http://' + str(url) + ':' + str(port)
        self.username = username
        self.password = password
        self.email = email

        self.id = 0
        self.usertoken = ''
        self.roomtoken = roomtoken
        self.x = self.y = -1
        self.playernum = playernum
        self.row = row
        self.col = col
        self.barback = barback
        self.portal  = portal
        self.barrier = barrier

        self.log = "" # 保留错误信息


    def info(self):
        print('\n')
        print('username : ', self.username)
        print('usertoken: ', self.usertoken)
        print('roomtoken: ', self.roomtoken)
        print('playernum: ', self.playernum)
        print('row: ', self.row, '\ncol: ', self.col)
        print('startX: ', self.x, '\nstartY: ',self.y)
        print('\n')
        print(self.barback, self.portal, self.barrier)

    # --* 返回log信息 *--
    def log(self):
        print(self.log)
        return self.log

    # --* 检验point合法性 *--
    def check(self, x, y):
        return x>=0 and y>=0 and x<self.row and y<self.col

    # --* 注册 *--
    # 返回值: status
    # status 1 => 注册成功  0 => 注册失败
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
            self.log = _register['message']
            print('<Error>[register] ', _register['message'])
        
        return _register['status']

    # --* 登录 *--
    # 返回值: status
    # status: 1 => 登录成功  2 =>断线重连成功
    def login(self):
        _login = requests.get(self.url + "/user?username=" + self.username + "&password=" + self.password, timeout=5).json()

        print('[login] ', _login)
        
        if _login['status'] == 1:
            self.usertoken = _login['usertoken']
        elif _login['status'] == 2:
            self.id = _login['id']
            self.usertoken = _login['usertoken']
            self.roomtoken = _login['roomtoken']
        else:
            self.log = _login['message']
            print('<Error>[login] ', _login['message'])
        
        return _login['status']

    # --* 是否已经在room中 *--
    def isInRoom(self):
        return self.roomtoken != ''

    # --* 加入或者创建房间 *--
    # 返回值: status
    # status 1 => join 成功  0 => 失败
    def join(self):
        data = { }
        if self.roomtoken == '': 
            data = {
                'usertoken': self.usertoken,
                'playernum': self.playernum,
                'row': self.row,
                'col': self.col,
                'barback': self.barback,
                'portal':  self.portal,
                'barrier': self.barrier
            }
        else:
            data = {
                'usertoken': self.usertoken,
                'roomtoken': self.roomtoken
            }

        payload = urllib.parse.urlencode(data)

        _join = requests.post(self.url + "/room", params=payload, timeout=5).json()
        
        print('[join] ', _join)
        
        if _join['status'] == 1:
            self.id = _join['id']
            self.roomtoken = _join['roomtoken']
            self.row = _join['row']
            self.col = _join['col']
            self.playernum = _join['playernum']
        else:
            self.log = _join['message']
            print('<Error>[join] ', _join['message'])

        return _join['status']

    # --* 获取游戏是否开始以及基地坐标 *--
    # 返回值: status
    # status 1 => start  0 => not start
    def isStart(self):
        _isStart = requests.get(self.url + "/room/start?usertoken=" + self.usertoken + "&roomtoken=" + self.roomtoken, timeout=5).json()
        
        # print('[isStart] ', _isStart)
        
        if _isStart['status'] == 1:
            self.x = _isStart['x']
            self.y = _isStart['y']
            self.row = _isStart['row']
            self.col = _isStart['col']
        
        return _isStart['status']

    # --* 返回基地坐标 *--
    def getBase(self):
        return (self.x, self.y)

    # --* 返回地图大小 *--
    def getMapSize(self):
        return (self.row, self.col)

    # --* 移动 *--
    # 返回值: status
    # status 1 => 成功  0 => 失败
    # PS: 这个返回值没啥用
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

        # print('[move] ', _move)
        if _move['status'] == 1:
            return (_move['length'], _move['status'])
        return _move['status']


    # --* 查询 *--
    # 返回值 (result, status)
    # status: 1 => 成功  0 => 失败
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
        
        # print('[query] ', _query)

        if _query['status'] != 1:
            self.log = _query['message']
            print('<Error>[_query] ', _query['message'])
            return ({}, _query['status'])
        
        return (_query['eyeshot'], _query['status'])

    # --* 查询得分状态(少一点 耗性能) *--
    # 返回值 (result, status)
    # status: 1 => 成功  0 => 失败
    def getScoreBoard(self):
        _scoreboard = requests.get(self.url + "/room/scoreboard?roomtoken=" + self.roomtoken, timeout=5).json()

        # print('[getScoreBoard] ', _scoreboard)
        
        if _scoreboard['status'] != 1:
            self.log = '<Error>[getScoreBoard]'
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
        self.roomtoken = ''

        print('[logout]')

    # --* 浏览器打开view页面 观看对战情况 *--
    def view(self):
        openUrl = self.url+'/view/'+self.roomtoken
        if self.chrome != '':
            webbrowser.register('chrome', None, webbrowser.BackgroundBrowser(self.chrome))

        try:
            webbrowser.get('chrome').open(openUrl, new=0, autoraise=True)
        except Exception as e:
            webbrowser.open(openUrl, new=0, autoraise=True)    


    # run
    def run(self):
        # 登录
        if not self.login():
            self.log()
            return False
        
        self.join()

        cnt = 0
        # 检测游戏是否开始
        while not self.isStart():
            time.sleep(3)
            cnt = cnt + 1
            if cnt == 40: # 两分钟
                print("超时未开始")
                self.leave()
                return False

        self.view() # 展示web页面 (非必须)
        return True
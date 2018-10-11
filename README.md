# CodeFight

### 注意

因文档更新并不十分及时, 若文档与代码有出入, 以代码为准. (咕咕咕咕...

### 游戏介绍

简单的代码对战游戏, 受到  [generals.io]() 以及知乎上 Color Fight 启发 .  Go fight! 

游戏中兵力即 Cell 中的数值, 两个不同阵营的兵力交战时, 结果为相减 .

游戏中有 兵营(barback)、据点(portal)、障碍物(barrier)、基地(base) 四种建筑物

兵营: 间隔较小轮数增加在其中的兵力

据点: 提升据点内兵力的实力 (即在其中的兵力相当于乘以一个因子, 出了据点即失效)

障碍物: 无法通过与摧毁之地

基地: 每个玩家仅一个, 被摧毁即失败

### 如何测试/运行

##### 依赖

```html
都为最新版
go:(可能需要翻墙)
	go get github.com/labstack/echo
	go get github.com/go-sql-driver/mysql
	go get golang.org/x/net/websocket
python:
	pip3 install requests
 	(如果想用Chrome打开前端展示页面, Windows可能需要设置代码里 gamePlayer1.py:11 chromepath变量的值)
```

##### 数据库

在 `go/src` 目录下新建文件 `config.txt` 内容格式为 `dbuser;dbpass;dbtablename;web_server_port`  (⚠️: 第四项是 web 服务器的监听端口)

```html
数据库结构:
	表名 userinfo
	id	 		    int(11) 		AUTO_INCREMENT
	username		text	 		utf8mb4_unicode_ci
	password		text	 		utf8mb4_unicode_ci
	email			text	 		utf8mb4_unicode_ci
	status			int(11)  		0->未审核 1->审核通过
	CreateAt		datetime 		CURRENT_TIMESTAMP
	UpdateAt		datetime		CURRENT_TIMESTAMP		on update CURRENT_TIMESTAMP
```

##### 命令

```shell
git clone https://github.com/userpro/CodeFight.git
cd CodeFight/go/src
go run main.go
python3 gamePlayer1.py
# 如果需要启动多个 player 测试, 修改 cleaner.py:6 的 playernum 数量和 cleaner2.py:14 的roomtoken(需要先启动 cleaner.py 获得)
```

### 如何编写自己的Bot

参考 `go/src/example/CodeWar/Utils.py` 简单将API封装了一个 class , 简单的主体框架可参考 `go/src/example/template.py` , 也可以参考本人写的一个智障障bot `go/src/example/cleaner.py (cleaner2.py是为了演示对战)`

更多游戏环境参数的设置在 `go/src/fight/config.go` 中

### Map

地图分两层 `m1`, `m2`  每个格子就叫Cell吧. (坐标均从左上角开始 包括返回值)

`m1`存放Cell数值

`m2` 存放Cell类型 (低8位有效)

对于`m2`:

```go
/* 
地形 Cell Type
每个Cell Type唯一 不会重叠
*/
_space_   = 0x00 // 000- ----  空地
_base_    = 0x20 // 001- ----  基地(一个玩家仅一个, 即初始位置)  
_barback_ = 0x40 // 010- ----  军营(一个地图有随机个, 需要抢占, 每回合可加一个单位兵力)
_portal_  = 0x60 // 011- ----  据点(一个地图有随机个, 需要抢占, 兵力在里面可提高防御力)
_barrier_ = 0x80 // 100- ----  障碍(不可经过, 不可占领)

/* 阵营 User Id */
_system_  = 0x00 // ---0 0000  地图默认id
_visitor_ = 0x01 // ---0 0001  玩家默认id(加入游戏后, 玩家id会从此开始往后递增分配, 每个玩家id唯一)

_user_mask_ = 0x1f // 0001 1111
_type_mask_ = 0xe0 // 1110 0000

```

### HTTP 接口

保证每个返回值都一定会有 message 和 status 两项

无特殊说明 status 返回 1 是成功, 0 是失败. (失败原因在 message 里)

注意: 优先判断 status , 在 status 为 0 时, 其他 json key 不一定存在 ( message 一定存在)

##### API 一览

```go
e.POST(  "/user", register)
e.GET(   "/user", login   )
e.DELETE("/user", logout  )

e.GET(   "/room", query   )
e.POST(  "/room", join    )
e.PUT(   "/room", move    )
e.DELETE("/room", leave   )
e.GET("/room/start",      isStart)
e.GET("/room/scoreboard", getScoreBoard)
```

##### 玩家相关

###### register

| 名称       | 说明                                                        |
| ---------- | ----------------------------------------------------------- |
| 功能       | 注册( register )                                            |
| 请求方法   | POST                                                        |
| URL        | /user                                                       |
| 参数       | username=XXX&password=XXX&email=XXX                         |
| 示例       | /user?username=hi&password=pro&email=abc@d.com              |
| 返回值     | { <br />    "message": "XXX",<br />     "status": 1<br /> } |
| 返回值说明 | status 为 1 仅表示信息注册成功, 需要审核通过才可以登录.     |

###### login

| 名称       | 说明                                                         |
| ---------- | ------------------------------------------------------------ |
| 功能       | 登录 ( login )                                               |
| 请求方法   | GET                                                          |
| URL        | /user                                                        |
| 参数       | username=XXX&password=XXX                                    |
| 示例       | /user?username=hi&password=pro                               |
| 返回值     | { <br />     "usertoken": "XXXX",<br />     "message": "XXX",<br />     "status": 1 <br /> } |
| 返回值说明 | 仅 status 为 1 时, 才会有 usertoken 项, 表示登录成功<br />当 status 为 2 时, 表示已经登录过, 为在线状态 |

###### logout

| 名称     | 说明                              |
| -------- | --------------------------------- |
| 功能     | 登出 (logout)                     |
| 请求方法 | DELETE                            |
| URL      | /user                             |
| 参数     | usertoken=XXX                     |
| 示例     | /user?usertoken=XXX               |
| 返回值   | 无 ( HTTP状态码为 NoContent 204 ) |

##### 房间相关

###### join

| 名称       | 说明                                                         |
| ---------- | ------------------------------------------------------------ |
| 功能       | 加入或者创建房间 ( join )                                    |
| 请求方法   | POST                                                         |
| URL        | /room                                                        |
| 参数       | 加入房间:<br />roomtoken=XXX<br />创建房间:<br />playernum=X&row=X&col=X |
| 示例       | 加入房间:<br />/room?roomtoken=XXXX<br />创建房间:<br />/room?playernum=X&row=X&col=X |
| 返回值     | {<br />    "id": X,<br />    "usertoken": "XXX",<br />    "roomtoken": "XXX",<br />    "playernum": "XXX",<br />    "row": XX,<br />    "col":XX,<br />    "barback": XX,<br />    "portal": XX,<br />    "barrier":XX,  <br />    "message": "XXX",<br />    "status": X<br />} |
| 返回值说明 | 返回所加入房间的信息, <br />id为本局游戏你的id, <br />barback 兵营数量, portal 据点数量, barrier 障碍物数量 |

###### leave


| 名称     | 说明                                                     |
| -------- | -------------------------------------------------------- |
| 功能     | 离开房间 (leave)                                         |
| 请求方法 | DELETE                                                   |
| URL      | /room                                                    |
| 参数     | usertoken=XXX                                            |
| 示例     | /room?usertoken=XXX                                      |
| 返回值   | {<br />    "message": "XXX",<br />    "status": 1<br />} |

###### query

| 名称       | 说明                                                         |
| ---------- | ------------------------------------------------------------ |
| 功能       | 查询游戏中对应 Cell 周围的信息  (query)                      |
| 请求方法   | GET                                                          |
| URL        | /room                                                        |
| 参数       | {<br />    "usertoken": "XXX",<br />    "roomtoken": "XXX",<br />    "loc": {<br />        "x": 1,<br />        "y": 3<br />    }<br />} |
| 返回值     | {<br />    "eyeshot": {<br />        "m1": \[n\]\[n\],<br />        "m2": \[n\]\[n\]<br />    },<br />    "status": 1<br />} |
| 返回值说明 | m1 m2 分别指代地图的两个图层, size 一致, <br />标示了以 (x, y) 为中心的一个矩阵的视野,<br />具体矩阵的大小以实际返回为准, 可能会有调整,<br />对于超出地图的部分的值在m1中以 -1 填充 |

###### move

| 名称       | 说明                                                         |
| ---------- | ------------------------------------------------------------ |
| 功能       | 移动地图上指定 Cell 的兵力 (move)                            |
| 请求方法   | PUT                                                          |
| URL        | /room                                                        |
| 参数       | {<br />    "usertoken": "XXX",<br />    "roomtoken": "XXX",<br />    "radio": 1,<br />    "direction": 1,<br />    "loc": {<br />        "x": 1,<br />        "y": 2<br />    }<br />} |
| 参数说明   | radio:<br />    1 -> all 调动 (x, y)  所有兵力<br />    2 -> half 调动 (x, y)  1/2的兵力<br />    3 -> quarter 调动 (x, y)  1/4的兵力<br />direction:<br />    1 -> (x, y) => (x - 1, y)<br />    2 -> (x, y) => (x, y + 1)<br />    3 -> (x, y) => (x + 1, y)<br />    4 -> (x, y) => (x, y - 1)<br />注意: loc的 (x, y) 必须是属于你的 Cell 才能查询 |
| 返回值     | {<br />    "length": 1,<br />    "status": 1<br />}          |
| 返回值说明 | length 是目前操作序列的长度                                  |

###### isStart

| 名称       | 说明                                                         |
| ---------- | ------------------------------------------------------------ |
| 功能       | 查询游戏是否开始 (isStart)                                   |
| 请求方法   | GET                                                          |
| URL        | /room/start                                                  |
| 参数       | usertoken=XXX&roomtoken=XXX                                  |
| 示例       | /room/start?usertoken=XXX&roomtoken=XXX                      |
| 返回值     | {<br />    "x": 1,<br />    "y": 2,<br />    "row": 30,<br />    "col": 30,<br />    "status": 1<br />} |
| 返回值说明 | x, y 为你初始坐标, 也是你的Base坐标<br />row, col 即 map 尺寸 |

###### getScoreBoard

| 名称       | 说明                                                         |
| ---------- | ------------------------------------------------------------ |
| 功能       | 获取当前房间内所有玩家的得分详情 <br />(getScoreBoard)       |
| 请求方法   | GET                                                          |
| URL        | /room/scoreboard                                             |
| 参数       | roomtoken=XXX                                                |
| 示例       | /room/scoreboard?roomtoken=XXX                               |
| 返回值     | {<br />    "scoreboard": {<br />        "username1": score1,<br />        "username2": score2,<br />        ....<br />    },<br />    "status": 1<br />} |
| 返回值说明 | scoreboard 是一个用户名和其得分的键值对<br />(注意: 请不要频繁调用, 会拖慢游戏进程) |

### Webscoket

路径: go/src/public/view.html

仅前端展示页面的数据接口, 可定制自己的前端展示页面 (直接替换view.html, 不要改动文件名).

流程如下:

1. 客户端 send => roomtoken到服务器 (注意 token 是通过模版标签`{{.}}`嵌入到页面的)
2. 客户端 receive <=游戏基本信息A( json格式数据)
3. 客户端 receive <= 每次操作B( json格式数据)

##### JSON格式

A:

```js
{
    type: 1, // 拉取地图信息 只在最开始出现一次
    value: {
        time: 100, // 游戏总时长 (秒)
        col: 30,
        row: 30,
        m1: [default_row][default_col], // default_row default_col是地图总共大小
        m2: [default_row][default_col], // 但每个游戏实际使用地图的尺寸是之前row col所指定的
        roomtoken: "7ac45ec7b7046c529ac650366700ec50", // roomtoken emmmm 顺便发的
        userinfo: [ // 一个obj数组
            {
                id: 2, // userid 对应地图m2
                name: "hipy", // username
                score: 2, // 每个user占有的格子数量 前端维护接下来的变化
                status: 1 // -1: 离线 0: 等待  1: 游戏中
            }
        ]
    }
}
```

B: (有三种)

```js
{
    type: 0 // 游戏结束
}
or
{
    type: 2, // 指定点改变
    value: {
        loc: [ // 一个obj数组
            {
                x: 1,
                y: 2,
                m1: 23, // m1[x][y]=23
                m2: 34  // m2[x][y]=34
            },
            {
                x: 2,
                y: 6,
                m1: 3,
                m2: 42
            },
        ]
    }
}
or
{
    type: 3 // 全图m1不为0的Cell加1 (除了barrier)
}
or
{
    type: 4 // 全图特殊cell加1 (军营, 基地)
}
```




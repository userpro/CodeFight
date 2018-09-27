# CodeFight
Go fight!

### HTTP 接口

文档后补 (咕咕咕咕...

### 如何测试/运行

##### 依赖

```html
都为最新版
go:
	github.com/labstack/echo
	github.com/go-sql-driver/mysql
python:
	requests
 	(如果想用Chrome打开查看, Windows可能需要设置代码里 gameUser1.py:7 chromepath变量的值)
```

##### 数据库

```html
修改 go/src/net/net.go:406-408 为本地数据库信息
数据库结构:
	表名 userinfo
	id(主键)		   int(11) 			AUTO_INCREMENT
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
python3 example.py
```



### Map

地图分两层 `m1,m2`  每个格子就叫Cell吧

m1存放Cell数值

m2 存放Cell类型 (低8位有效)

对于m2:

```go
/* 
地形 Cell Type
每个Cell Type唯一 不会重叠
*/
_space_   = 0x00 // 000- ----  空地
_base_    = 0x20 // 001- ----  基地(一个玩家仅一个)  
_barback_ = 0x40 // 010- ----  军营(一个地图有随机个, 需要抢占, 每回合可加一个单位兵力)
_portal_  = 0x60 // 011- ----  据点(一个地图有随机个, 需要抢占, 兵力在里面可提高防御力)
_barrier_ = 0x80 // 100- ----  障碍(不可经过, 不可占领)

/* 阵营 User Id */
_system_  = 0x00 // ---- -000  地图默认id
_visitor_ = 0x01 // ---- -001  玩家默认id(加入游戏后, 玩家id会从此开始往后递增分配, 每个玩家id唯一)

_user_mask_ = 0x1f // 0001 1111
_type_mask_ = 0xe0 // 1110 0000
```



### Webscoket

流程如下:

1. 客户端 send => roomtoken到服务器
2. 客户端 receive <=游戏基本信息A( json格式数据)
3. 客户端 receive <= 每次操作B( json格式数据)

##### JSON格式

A:

```js
{
    type: 1, // 拉取地图信息 只在最开始出现一次
    value: {
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




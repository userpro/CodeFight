package net

import (
    // "fmt"
    "log"
    "os"
    "sync"
    "time"
    "net/http"
    "strconv"
    "strings"
    "html/template"

    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "github.com/labstack/echo"
    // "github.com/labstack/echo/middleware"

    fight "../fight"
    eventQ "../event"
)

var (
    eventQMap sync.Map // map[roomToken]*eventQ
    netToken  sync.Map // [userToken]*netUserInfo
    netOnline sync.Map // [username]bool
    WSChannelMap sync.Map // map[roomtoken]*WSChannel

    dbw       DbWorker
    netLogger *log.Logger
    dbLogger  *log.Logger
)

// 登录后 未游戏时长超时
func loginTimeOut(utk string) {
    for {
        <- time.After(fight.Default_login_timeout)
        rtk := fight.GetRoom(utk)
        if rtk != "" {
            _, _, gameStart := fight.IsStart(utk, rtk)
            if gameStart { continue }
        }
        v, ok := netToken.Load(utk)
        if !ok { return }
        netOnline.Delete(v.(*netUserInfo).Uname)
        netToken.Delete(utk)
        fight.Logout(utk)
        netLogger.Println("loginTimeOutout")
        return
    }
}

// Form Data
// Return Type: JSON
// curl -d "username=hipro&password=okiamhi&email=34@qq.com" http://127.0.0.1:8080/user
func register(c echo.Context) error {
    uname := c.FormValue("username")
    if uname == "" { return c.JSON(http.StatusOK, &RespInfo{ Message:"Failed! Username can't empty.", Status:0}) }
    pwd   := c.FormValue("password")
    if pwd == "" { return c.JSON(http.StatusOK, &RespInfo{ Message:"Failed! Password can't empty.", Status:0}) }
    email := c.FormValue("email")
    if email == "" { return c.JSON(http.StatusOK, &RespInfo{ Message:"Failed! Email can't empty.", Status:0}) }
    ok := dbw.QueryData(uname, pwd)
    if ok {
        return c.JSON(http.StatusOK, &RespInfo{ Message:"Failed! Username have exist.", Status:0})
    }
    dbw.InsertData(uname, pwd, email)
    return c.JSON(http.StatusOK, &RespInfo{ Message:"Register OK. Waiting for Review.", Status:1})
}

// QueryParam
// Return Type: JSON
// curl http://127.0.0.1:8080/user\?username\=hipro\&password\=okiamhi
// Return Result: 0=>"failed"  1=>"ok"  2=>"have logined"
func login(c echo.Context) error {
    /* 检查uname */
    uname := c.QueryParam("username")
    if uname == "" { return c.JSON(http.StatusOK, &RespInfo{ Message:"Failed! Username can't empty.", Status:0}) }
    /* 检查pwd */
    pwd   := c.QueryParam("password")
    if pwd == "" { return c.JSON(http.StatusOK, &RespInfo{ Message:"Failed! Password can't empty.", Status:0}) }
    /* 判断是否已经在线 */
    _, ok := netOnline.Load(uname)
    if ok {
        return c.JSON(http.StatusOK, &RespInfo{ Message: "Login OK! You have login before.", Status:2})
    }

    ok = dbw.QueryData(uname, pwd)
    if ok {
        if dbw.UserInfo.Status == 0 {
            return c.JSON(http.StatusOK, &RespInfo{ Message:"Login Failed! Waiting for Review.", Status:0})
        }
        // newUname, newPwd, newEmail, uname, pwd
        dbw.UpdateData(uname, pwd, dbw.UserInfo.Email, uname, pwd)
        utk := fight.GenToken(uname + pwd) // usertoken
        // 维护token信息
        netToken.Store(utk, &netUserInfo {
            Uname: uname,
            UserToken: utk,
        })
        // netLogger.Println("[login] netToken: ", netToken[usertoken])
        // 维护在线状态
        netOnline.Store(uname, true)
        fight.Login(utk)
        // 设置登录失效超时
        go loginTimeOut(utk)
        return c.JSON(http.StatusOK, &netUserRet{
            UserToken: utk,
            RespInfo: RespInfo {
                Message: "Login OK.",
                Status: 1,
            },
        })
    }
    return c.JSON(http.StatusOK, &RespInfo{ Message:"Login Failed! Username or password wrong.", Status:0})
}

// QueryParam
// Return Type: String
// curl -v -X DELETE http://127.0.0.1:8080/user\?usertoken=...
func logout(c echo.Context) error {
    utk := c.QueryParam("usertoken")
    if utk == "" { return c.NoContent(http.StatusNoContent) }
    // netLogger.Println("[logout] uname: ", netToken[utk])
    u, ok := netToken.Load(utk)
    if !ok { return c.NoContent(http.StatusNoContent) }
    /* 如果还没退出Room 则先移除事件队列中的相关项 */
    rtk := fight.GetRoom(utk)
    if rtk != "" { 
        _evq, ok := eventQMap.Load(rtk)
        evq := _evq.(*eventQ.EventQueue)
        if ok { evq.Remove(utk) }
        eventQMap.Delete(rtk)
    }

    fight.Logout(utk)
    netOnline.Delete(u.(*netUserInfo).Uname)
    netToken.Delete(utk)
    return c.NoContent(http.StatusNoContent)
}

// Form Data
// Return Type: JSON
// curl -d "usertoken=...&playnum=..." http://127.0.0.1:8080/room
func join(c echo.Context) error {
    /* 检查usertoken */
    utk := c.FormValue("usertoken")
    if utk == "" { return c.JSON(http.StatusOK, &RespInfo{ Message:"Failed! UserToken can't empty.", Status:0}) }
    ntk, ok := netToken.Load(utk)
    if !ok { return c.JSON(http.StatusUnauthorized, &RespInfo{ Message:"Failed! You haven't login!", Status:0 }) }

    /* 检查roomtoken */
    rtk := c.FormValue("roomtoken")
    if rtk == "" {
        /* 没有rtk 则新建room */
        playnum, err := strconv.Atoi(c.FormValue("playnum"))
        if err != nil { playnum = fight.Default_player_num }
        row, err := strconv.Atoi(c.FormValue("row"))
        if err != nil { row = fight.Default_row }
        col, err := strconv.Atoi(c.FormValue("col"))
        if err != nil { col = fight.Default_col }
        msg ,ok := fight.NewRoom(utk, playnum, row, col)
        if ok {
            rtk = msg
            id, ok := fight.GetUserId(utk)
            if !ok { return c.JSON(http.StatusUnauthorized, &RespInfo{ Message:"Failed! You haven't login!", Status:0 }) }

            nntk := ntk.(*netUserInfo)
            nntk.Id = id
            // 如果 room 人数已满 则开始
            // netLogger.Println(fight.IsStart(utk, rtk))
            _, _, gameStart := fight.IsStart(utk, rtk)
            if gameStart {
                netLogger.Println("[Join] Started!")
                teq := eventQ.New()
                eventQMap.Store(rtk, teq) // 绑定事件循环到roomtoken
                // 加入websocket
                wsc := fight.WSNew()
                WSChannelMap.Store(rtk, wsc)
                // 启动游戏 开一个协程等待游戏结束
                go func(rtk string, teq *eventQ.EventQueue, wsc *fight.WSChannel) {
                    ch := fight.Run(rtk, teq, wsc)
                    // 当游戏结束时 回收该房间eventQueue
                    <- ch
                    eventQMap.Delete(rtk)
                    WSChannelMap.Delete(rtk)
                }(rtk, teq, wsc)
            }
            return c.JSON(http.StatusOK, &netUserRet{
                UserToken: utk,
                RoomToken: rtk,
                RoomOptRet:  RoomOptRet {
                    Row: row,
                    Col: col,
                    Playnum: playnum,
                },
                RespInfo: RespInfo {
                    Message: "Join Room OK!",
                    Status: 1,
                },
            }) 
        }
        // Join Failed
        return c.JSON(http.StatusOK, &RespInfo{ Message:msg, Status:0 })
    }

    /* 有rtk 直接加入 */
    msg, ok := fight.Join(utk, rtk)
    if ok {
        // 如果 room 人数已满 则开始
        _, _, gameStart := fight.IsStart(utk, rtk)
        if gameStart {
            // netLogger.Println("[Join] Started!")
            teq := eventQ.New()
            eventQMap.Store(rtk, teq) // 绑定事件循环
            // websocket channel绑定
            _wsc, wscok := WSChannelMap.Load(rtk)
            if !wscok { return c.JSON(http.StatusOK, &RespInfo{ Message:msg, Status:0 }) }
            wsc := _wsc.(*fight.WSChannel)
            go fight.Run(rtk, teq, wsc)
        }

        s := strings.Split(msg, ",")
        id, _ := strconv.Atoi(s[0])
        playnum, _ := strconv.Atoi(s[1])
        row, _ := strconv.Atoi(s[2])
        col, _ := strconv.Atoi(s[3])
        nntk := ntk.(*netUserInfo)
        nntk.Id = byte(id);
        return c.JSON(http.StatusOK, &netUserRet{
            UserToken: utk,
            RoomToken: rtk,
            RoomOptRet: RoomOptRet{
                Row: row,
                Col: col,
                Playnum: playnum,
            },
            RespInfo: RespInfo{ Status:1 },
        })
    }
    // Join Failed
    return c.JSON(http.StatusOK, &RespInfo{ Message: msg, Status:0})
}

// GET
// JSON
/* test-cmd
    curl -H 'Content-Type: application/json' -X GET -d \ 
    '{"usertoken":"", "roomtoken":"", "loc":{"x":24,"y":12}}' \
    http://127.0.0.1:8080/room
*/
func query(c echo.Context) error {
    u := &netQueryReq{}
    err := c.Bind(u)
    if err != nil { return err }
    /* 检查usertoken */
    utk := u.UserToken
    if utk == "" { return c.JSON(http.StatusOK, &RespInfo{ Message:"Failed! UserToken can't empty.", Status:0}) }
    /* 检查是否在线 */
    _, ok := netToken.Load(utk)
    if !ok { return c.JSON(http.StatusUnauthorized, &RespInfo{ Message:"Failed! You haven't login!", Status:0 }) }

    /* 检查roomtoken */
    rtk := u.RoomToken
    if rtk == "" { return c.JSON(http.StatusOK, &RespInfo{ Message:"Failed! RoomToken can't empty.", Status:0}) }

    /* 获取到eye */
    eye, msg, ok := fight.GetEyeShot(u.UserToken, u.RoomToken, u.Loc)
    if !ok { return c.JSON(http.StatusUnauthorized, &RespInfo{ Message:msg, Status:0 }) }

    return c.JSON(http.StatusOK, &netQueryRet{ 
        Eye:eye,
        RespInfo: RespInfo{ Status:1 },
    })
}


// PUT
// JSON
// Request {usertoken, roomtoken, point, radio, direction}
/* test-cmd
    curl -H 'Content-Type: application/json' -X PUT \
    -d '{"usertoken":"", "roomtoken":"", "radio":1, "direction":1, "loc":{"x":1,"y":1}}' \
    http://127.0.0.1:8080/room 
*/
func move(c echo.Context) error {
    u := &netMoveReq{}
    err := c.Bind(u)
    if err != nil { return err }
    /* 检查usertoken */
    utk := u.UserToken
    if utk == "" { return c.JSON(http.StatusOK, &RespInfo{ Message:"Failed! UserToken can't empty.", Status:0}) }
    /* 检查是否在线 */
    _, ok := netToken.Load(utk)
    if !ok { return c.JSON(http.StatusUnauthorized, &RespInfo{ Message:"Failed! You haven't login!", Status:0 }) }

    /* 检查roomtoken */
    rtk := u.RoomToken
    if rtk == "" { return c.JSON(http.StatusOK, &RespInfo{ Message:"Failed! RoomToken can't empty.", Status:0}) }
    
    /* 获取到roomtoken的eventQueue */
    _evq, ok := eventQMap.Load(rtk)
    if !ok { return c.JSON(http.StatusUnauthorized, &RespInfo{ Message:"Failed! Not that room." }) }
    evq := _evq.(*eventQ.EventQueue)

    /* 添加event到eventQueue中 */
    // netLogger.Println("[net-Move] ", u)
    moveEve := fight.ActionMove{
        Radio: u.Radio,
        Direction: u.Direction,
        Loc: u.Loc,
    }
    acEve := fight.ActionEvent{
        Token: utk,
        Typ: fight.Action_move_,
        Ac: moveEve,
    }
    actionCnt := evq.Push(utk, acEve)

    return c.JSON(http.StatusOK, &netMoveRet{ 
        Length: actionCnt,
        RespInfo: RespInfo{Status:1}, 
    })
}

// QueryParam
// curl -v -X DELETE http://127.0.0.1:8080/room?usertoken=...&roomtoken=...
func leave(c echo.Context) error {
    utk := c.QueryParam("usertoken")
    if utk == "" { return c.JSON(http.StatusOK, &RespInfo{ Message:"Failed! Usertoken can't empty.", Status:0}) }
    rtk := c.QueryParam("roomtoken")
    if rtk == "" { return c.JSON(http.StatusOK, &RespInfo{ Message:"Failed! Roomtoken can't empty.", Status:0}) }

    msg, ok := fight.LeaveRoom(utk, rtk)
    if !ok { return c.JSON(http.StatusOK, &RespInfo{ Message:msg, Status:0 }) }
    return c.JSON(http.StatusOK, &RespInfo{ Message:msg, Status:1 })
}

// GET
// QueryParam
// curl http://127.0.0.1:8080/room/start?usertoken=...&roomtoken=...
func isStart(c echo.Context) error {
    utk := c.QueryParam("usertoken")
    /* 检查usertoken */
    if utk == "" { return c.JSON(http.StatusOK, &RespInfo{ Message:"Failed! UserToken can't empty.", Status:0}) }
    /* 检查是否在线 */
    _, ok := netToken.Load(utk)
    if !ok { return c.JSON(http.StatusUnauthorized, &RespInfo{ Message:"Failed! You haven't login!", Status:0 }) }

    /* 检查roomtoken */
    rtk := c.QueryParam("roomtoken")
    if rtk == "" { return c.JSON(http.StatusOK, &RespInfo{ Message:"Failed! RoomToken can't empty.", Status:0}) }

    x, y, ok := fight.IsStart(utk, rtk)
    if !ok { return c.JSON(http.StatusOK, &RespInfo{ Status:0 }) }
    return c.JSON(http.StatusOK, &netStartRet{
        X:x, Y:y, 
        RespInfo: RespInfo{ Status: 1 },
    })
}

// GET
// QueryParam
// Return Status 
/*
    -1 => offline   0 => wait
    1  => playing   2 => win    3 => lose
*/
func getStatus(c echo.Context) error {
    utk := c.QueryParam("usertoken")
    /* 检查usertoken */
    if utk == "" { return c.JSON(http.StatusOK, &RespInfo{ Message:"Failed! UserToken can't empty.", Status:0}) }
    /* 检查是否在线 */
    _, ok := netToken.Load(utk)
    if !ok { return c.JSON(http.StatusUnauthorized, &RespInfo{ Message:"Failed! You haven't login!", Status:0 }) }

    return c.JSON(http.StatusOK, &netUserStatusRet{ 
        Result:fight.GetStatus(utk),
        RespInfo: RespInfo { Status: 1 },
    })
}


func Run() {
    var err error
    dbw = DbWorker{
        Dsn: "root:123@tcp(localhost:3306)/CodeWar?charset=utf8mb4&parseTime=true&loc=Local",
    }
    dbw.Db, err = sql.Open("mysql", dbw.Dsn)
    if err != nil {
        panic(err)
        return
    }
    defer dbw.Db.Close()

    /* html/template render */
    templateRenderer := &Template{
    templates: template.Must(template.ParseGlob("public/*.html"))}

    e := echo.New()
    e.Renderer = templateRenderer
    // e.Use(middleware.Logger())
    // e.Use(middleware.Recover())
    e.POST(  "/user", register)
    e.GET(   "/user", login   )
    e.DELETE("/user", logout  )
    e.GET("/user/status", getStatus)

    e.GET(   "/room", query   )
    e.POST(  "/room", join    )
    e.PUT(   "/room", move    )
    e.DELETE("/room", leave   )
    e.GET("/room/start", isStart)

    e.GET("/view/:roomtoken", view)
    // e.Static("/", "public")
    e.GET("/ws", wsocketView)

    e.GET("/", func(c echo.Context) error {
        return c.Render(http.StatusOK, "index.html", "")
    })
    e.Logger.Fatal(e.Start(":8080"))
}

func init() {
    netLogger = log.New(os.Stdout, "[net] ", log.Ldate | log.Ltime | log.Lshortfile)
    dbLogger = log.New(os.Stdout, "[DB] ", log.Ldate | log.Ltime | log.Lshortfile)
}

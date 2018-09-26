package fight

import (
    // "fmt"
    "os"
    "log"
    "time"
    "sync"
    "strconv"

    eventQ "../event"
)

var (
    fUserList sync.Map // map[userToken]fUser
    fOptsList sync.Map // map[roomToken]fOpts
    fMapList  sync.Map // map[roomToken]*fMap
    WSChannelMap sync.Map // [token]*chan WSAction
    fightLogger  *log.Logger
)

func NewRoom(userToken string, opts ...int) (string, bool) {
    user, ok := getUser(userToken)
    if !ok { return "You haven't login.", false }
    if user.roomToken != "" { return user.roomToken, true }

    roomToken := GenToken(userToken)

    // 可定制参数: 游戏人数 战场size
    var tPlayerNum, trow, tcol int
    if len(opts) == 1 {
        tPlayerNum = opts[0]
        trow = Default_row
        tcol = Default_col
    } else if len(opts) == 3 {
        tPlayerNum = opts[0]
        trow = opts[1]
        tcol = opts[2]
    } else {
        tPlayerNum = Default_player_num
        trow = Default_row
        tcol = Default_col
    }

    if tPlayerNum > Default_player_num || trow > Default_row || tcol > Default_col {
        return "Size or player number out of limit!", false
    }

    opt := &fOpts {
        accId: _visitor_ + 1,
        playerNum: tPlayerNum,
        row: trow,
        col: tcol,
        playing: false,
        userToken: make(map[byte]*fUser),
        roomToken: roomToken,
    }

    // 注册 room
    fOptsList.Store(roomToken, opt)

    // 注册 user
    fUserList.Store(userToken, &fUser {
        id: _visitor_,
        userToken: userToken,
        score : 0,
        energy: 0,
    })

    // 房间创建者 自动加入
    _, ok = Join(userToken, roomToken)
    if !ok {
        fOptsList.Delete(roomToken)
        fUserList.Delete(userToken)
        return "New Room Failed!", false
    }
    // fightLogger.Println(getOpts(roomToken))

    return roomToken, true
}

func LeaveRoom(userToken, roomToken string) (string, bool) {
    user, ok1 := getUser(userToken)
    if !ok1 { return "You haven't login.", false }
    opt, ok2 := getOpts(roomToken)
    if !ok2 { return "Not found room.", false}
    delete(opt.userToken, user.id)
    // 一个 room 没有 player 的时候, 回收 room.
    if len(opt.userToken) == 0 {
        fightLogger.Println("[LeaveRoom] Destroy! ")
        fMapList.Delete(roomToken)
        fOptsList.Delete(roomToken)
    }
    // 更新 User 信息
    user.id = _visitor_
    user.roomToken = ""
    user.score = 0
    user.energy = 0
    // user.status = US_wait_ // 不修改 在join的时候会自动修改
    return "Leave.", true
}

func Join(userToken, roomToken string) (string, bool) {
    user, ok1 := getUser(userToken)
    if !ok1 { return "You haven't login.", false }
    if user.roomToken == roomToken { return "You have in it.", true }
    opt, ok2 := getOpts(roomToken)
    if !ok2 { return "Not found room.", false }
    if opt.playing == true { return "The room is playing.", false }
    if len(opt.userToken) >= opt.playerNum { return "The room is full.", false }

    user.id = opt.accId
    user.roomToken = roomToken
    user.score = 0; user.energy = 0

    opt.userToken[opt.accId] = user
    opt.accId++

    /* 人数满足 游戏开始 */
    if len(opt.userToken) == opt.playerNum { 
        ok := opt.generator() 
        if !ok { fightLogger.Println("[Join] generator Failed.") }
        opt.playing = true
        for _, v := range opt.userToken {
            v.status = US_playing_
            v.score = 1 // 初始时 只有一个base
        }
    }
    // id, playerNum, row, col
    return strconv.Itoa(int(user.id)) + "," + strconv.Itoa(opt.playerNum) + "," + strconv.Itoa(opt.row) + "," + strconv.Itoa(opt.col), true
}

func Login(userToken string) (string, bool) {
    fUserList.Store(userToken, &fUser{
        id: _visitor_,
        userToken: userToken,
        score : 0,
        energy: 0,
        status: US_wait_,
    })
    return userToken, true
}

func Logout(userToken string) {
    user, ok := getUser(userToken)
    if !ok { return }
    if user.roomToken != "" {
        msg, ok := LeaveRoom(userToken, user.roomToken)
        if !ok { fightLogger.Println("[Logout] ", msg) }
    }
    fUserList.Delete(userToken)
}

func GetUserId(userToken string) (byte, bool) {
    user, ok := getUser(userToken)
    if !ok { return _system_, false }
    return user.id, true
}

func GetRoom(userToken string) string {
    user, ok := getUser(userToken)
    if !ok { return "" }
    return user.roomToken
}

func GetGameInfo(roomToken string) *WSAction {
    opt, ok := getOpts(roomToken)
    if !ok { return nil }
    mm, ok := getMap(roomToken)
    if !ok { return nil }
    u := &WSMapInfoRet {
        Row: opt.row,
        Col: opt.col,
        RoomToken: opt.roomToken,
        M1:  &mm.m1,
        M2:  &mm.m2,
    }
    for k, v := range opt.userToken {
        u.UserInfo = append(u.UserInfo, &WSUserInfoRet{
            Uid: k,
            Utoken: v.userToken,
            Score: v.score,
            Energy: v.energy,
            Status: v.status,
        })
    }
    return &WSAction{
        Typ: WSAction_mapinfo,
        Value: u,
    }
    // return &mm.m1, &mm.m2
}

func GetStatus(userToken string) int {
    user, ok := getUser(userToken)
    if !ok { return US_offline_ }
    return user.status
}

func GetEyeShot(userToken, roomToken string, loc Point) (*EyeShot, string, bool) {
    if !checkLoc(roomToken, loc) { return &EyeShot{}, "Invalid Point!", false }
    user, ok1 := getUser(userToken);
    if !ok1 { return &EyeShot{}, "You haven't login!", false }
    opt, ok2 := getOpts(user.roomToken)
    if !ok2 { return &EyeShot{}, "Not found room!", false }
    if !isPlaying(opt) { return &EyeShot{}, "Game is not start!", false }
    mm, ok3 := getMap(user.roomToken)
    if !ok3 { return &EyeShot{}, "Not found map!", false }

    // fightLogger.Println("[GetEyeShot] x:",loc.X, " y:",loc.Y)
    // fightLogger.Println("[GetEyeShot] userid: ",user.id, " cellid: ",getUserId(mm.m2[loc.X][loc.Y]))

    if user.id != getUserId(mm.m2[loc.X][loc.Y]) { return &EyeShot{}, "Not belong to you!", false }

    v := &EyeShot{}
    ltx := loc.X - default_eye_level
    lty := loc.Y - default_eye_level
    lth := default_eye_level * 2 + 1
    for i:=0; i<lth; i++ {
        for j:=0; j<lth; j++ {
            cx := ltx + i
            cy := lty + j
            if (cx < 0 || cy < 0 || cx >= opt.row || cy >= opt.col) {
                v.M1[i][j]=-1;
            } else {
                v.M1[i][j]=mm.m1[cx][cy]
                v.M2[i][j]=mm.m2[cx][cy]
            }
        }
    }
    return v, "OK", true
}

func Move(userToken, roomToken string, direction, radio int, loc Point) (*fPoint, *fPoint, bool) {
    if !checkLoc(roomToken, loc) { 
        fightLogger.Println("[Move] Invalid Point!")
        return nil,nil,false 
    }
    user, ok1 := getUser(userToken)
    if !ok1 { 
        fightLogger.Println("[Move] You haven't login!")
        return nil,nil,false 
    }
    opt, ok2 := getOpts(roomToken)
    if !ok2 { 
        fightLogger.Println("[Move] Not found room!")
        return nil,nil,false 
    }
    if !isPlaying(opt) { 
        fightLogger.Println("[Move] Game is not playing!")
        return nil,nil,false 
    }
    mm, ok3 := getMap(roomToken)
    if !ok3 { 
        fightLogger.Println("[Move] Not found map!")
        return nil,nil,false 
    }

    x := loc.X; y := loc.Y
    cell1 := mm.m2[x][y]
    uid  := getUserId(cell1)
    if uid != user.id { 
        fightLogger.Println("[Move] Not belong to you!")
        return nil,nil,false 
    }
    if isBarrier(cell1) { 
        fightLogger.Println("[Move] Can't move from!")
        return nil,nil,false 
    }
    if mm.m1[x][y] <= 1 { 
        fightLogger.Println("[Move] Don't have enough army!")
        return nil,nil,false 
    }

    finalRadio := getFinalRadio(radio)

    var nx, ny int
    switch direction {
        case _top_:     nx = x - 1; ny = y
        case _right_:   nx = x;     ny = y + 1
        case _bottom_:  nx = x + 1; ny = y
        case _left_:    nx = x;     ny = y - 1
        default: 
            fightLogger.Println("[Move] Invalid direction!")
            return nil,nil,false
    }
    if !checkLoc(roomToken, Point{nx,ny}) { 
        fightLogger.Println("[Move] Invalid next Point!")
        return nil,nil,false 
    }
    // 将要进入的cell的type
    cell2 := mm.m2[nx][ny]
    if isBarrier(cell2) { 
        fightLogger.Println("[Move] Can't move to!")
        return nil,nil,false 
    }

    // fightLogger.Println("[move] x:",x," y:",y," nx:",nx," ny:",ny)

    // 军队调出{x, y}
    var t1 int
    switch finalRadio {
        case _all_:     t1 = mm.m1[x][y] - 1
        case _half_:    t1 = mm.m1[x][y] / 2
        case _quarter_: t1 = mm.m1[nx][ny] / 4
    }
    
    mm.m1[x][y] -= t1
    /* 领地内调动 */
    uid = getUserId(cell2)
    if user.id == uid {
        mm.m1[nx][ny] += t1
    } else {
        /* 敌对领域 attack */
        t2 := mm.m1[nx][ny]
        // portal 防御提升
        if isPortal(cell2) { t2 = int(float32(t2) * default_portal_factor) }

        // 成功占领
        if t1 >= t2 {
            t1 -= t2
            if t1 == 0 { t1 = 1 }

            mm.m1[nx][ny] = t1
            mm.m2[nx][ny] = setCellId(cell2, user.id)

            // 更新自己score
            user.score++;
            // 如果是_system_不需要判断之后
            if isSystem(cell2) {
                return &fPoint{x:x,y:y,w:mm.m1[x][y]},&fPoint{x:nx,y:ny,w:mm.m1[nx][ny]},true
            }
            // 更新敌人score
            opt.userToken[uid].score--

            if isBase(cell2) { // base
                uid := getUserId(cell2)
                utk := opt.userToken[uid].userToken
                usr, _ := getUser(utk)
                someOneGameOver(usr, opt, mm)
                mm.removeBase(Point{nx,ny})
            } else if isPortal(cell2) { // portal
                mm.m2[nx][ny] = setCellType(cell2, _space_)
                mm.removePortal(Point{nx,ny})
            }
        } else {
            // 消耗
            mm.m1[nx][ny] -= int(float32(t1) / default_portal_factor)
        }
    }
    // fightLogger.Println("[move] m1[xy]: ",mm.m1[x][y], " m1[nxny]: ",mm.m1[nx][ny])
    return &fPoint{x:x,y:y,w:mm.m1[x][y]},&fPoint{x:nx,y:ny,w:mm.m1[nx][ny]},true
}


func doIt(roomToken string, actions []eventQ.EventEle, wsActions *WSAction) {
    if len(actions) <= 0 { return }
    var wsChange WSChange // websocket change数据
    for _, v := range(actions) {
        ac := v.Value.(ActionEvent) // v.Token => usertoken
        switch ac.Typ {
        case Action_move_:
            // fightLogger.Println("[doIt] Move ", v.Token)
            aMove := ac.Ac.(ActionMove)
            p1, p2, moveok := Move(v.Token, roomToken, aMove.Direction, getFinalRadio(aMove.Radio), aMove.Loc)
            if !moveok { continue }
            
            wsChange.Loc = append(wsChange.Loc, &WSPoint{ X: p1.x, Y:p1.y, W:p1.w })
            wsChange.Loc = append(wsChange.Loc, &WSPoint{ X: p2.x, Y:p2.y, W:p2.w })

        // case Action_magic_:
        //     fightLogger.Println("[doIt] Magic ", v.Token)
        default:
        }
    }
    wsActions.Typ = WSAction_normal_change
    wsActions.Value = wsChange
}


// Return Result: 0=>"failed"  1=>"end"
func Run(roomToken string, eq *eventQ.EventQueue, ws *WSChannel) chan bool {
    opt, ok1 := getOpts(roomToken)
    if !ok1 { return nil }
    mm, ok2  := getMap(roomToken)
    if !ok2 { return nil }

    // 设置eventQueue
    var tokens []string
    for _, v := range(opt.userToken) {
        tokens = append(tokens, v.userToken)
    }
    eq.Initialize(tokens)
    // fightLogger.Println("[Run] EventQueue: ", eq)

    // 游戏总时长
    playTimer := time.NewTicker(default_play_timeout)
    // 每轮action时间间隔
    actionTimer := time.NewTicker(default_action_interval)

    gameEnd := make(chan bool)

    go func(roomToken string, playTimer, actionTimer *time.Ticker, eq *eventQ.EventQueue, ws *WSChannel, gameEnd chan bool) {
        defer playTimer.Stop()
        defer actionTimer.Stop()
        defer ws.WSDestroy()

        loop_cnt        := 0
        global_add_cnt  := 0
        special_add_cnt := 0

        for {
            select {            
            case <-actionTimer.C:
                var wsActions []*WSAction // websocket时间队列
                loop_cnt++ // 统计游戏共经过几轮操作
                global_add_cnt++;
                if global_add_cnt >= default_global_add {
                    autoGlobalAdd(opt, mm)
                    global_add_cnt = 0
                    // 增加websocket
                    wsActions = append(wsActions, &WSAction{
                        Typ: WSAction_global_add,
                    })
                }
                special_add_cnt++;
                if special_add_cnt >= default_special_add {
                    autoSpecialAdd(mm)
                    special_add_cnt = 0
                    // 增加websocket
                    wsActions = append(wsActions, &WSAction{
                        Typ: WSAction_special_add,
                    })
                }
                /* 
                    如果所有玩家logout造成退出房间 
                    logout 会造成 usertoken 不一致 而无法重新加入房间
                */
                if eq.Empty() { gameEnd <- true; return }
                // 处理每轮的操作
                actions := eq.Get()
                if len(actions) > 0 {
                    var wsNormalAction WSAction // 普通的change
                    doIt(roomToken, actions, &wsNormalAction)
                    wsActions = append(wsActions, &wsNormalAction)
                }

                if !ws.WSEmpty() {
                    for _, v := range wsActions {
                        ws.WSBroadcast(v)
                    }
                }

            default: /* nothing */
            }

            /* 时间到结束 */
            select {
            case <-playTimer.C:
                fightLogger.Println("[Run] Game End!")
                var winner *fUser
                winner = nil
                mxScore := 0
                for _, v := range(opt.userToken) {
                    if v.score > mxScore {
                        mxScore = v.score
                        if winner != nil { winner.status = US_lose_ }
                        winner = v
                    }
                    LeaveRoom(v.userToken, roomToken)
                }
                winner.status = US_win_
                // 清空事件队列
                eq.Clear()
                gameEnd <- true
                return

            default: /* nothing */
            }
        }
    }(roomToken, playTimer, actionTimer, eq, ws, gameEnd)
   
    return gameEnd
}

func init() {
    fightLogger = log.New(os.Stdout, "[fight] ", log.Ldate | log.Ltime | log.Lshortfile)
}
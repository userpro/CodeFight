package fight

import (
    "os"
    "log"
    "time"
    "sync"

    eventQ "../event"
)

var (
    fUserList      sync.Map // map[userToken]*fUser
    fQueryCountMap sync.Map // map[usertoken]*fQueryCounter

    fOptsList      sync.Map // map[roomToken]*fOpts
    fScoreBoardMap sync.Map // map[roomtoken](*map[string]int)
    WSChannelMap   sync.Map // [roomtoken]*chan WSAction
    
    fightLogger    *log.Logger
)

func NewRoom(userToken string, params ...int) (interface{}, int, bool) {
    _, ok := getUser(userToken)
    if !ok { return "You haven't login.", Game_failed_response_, false }

    // 可定制参数: 游戏人数 战场size
    var tPlayerNum, trow, tcol int
    if len(params) == 1 {
        tPlayerNum = params[0]
        trow = Default_row
        tcol = Default_col
    } else if len(params) == 3 {
        tPlayerNum = params[0]
        trow = params[1]
        tcol = params[2]
    } else {
        tPlayerNum = Default_player_num
        trow = Default_row
        tcol = Default_col
    }

    if tPlayerNum > Default_player_num || trow > Default_row || tcol > Default_col {
        return "Size or player number out of limit!", Game_failed_response_, false
    }

    roomToken := GenToken(userToken)
    opt := &fOpts {
        accId: _visitor_ + 1,
        playerNum: tPlayerNum,
        row: trow,
        col: tcol,
        playing: false,
        userInfo:  make(map[byte]*fUser),
        roomToken: roomToken,
    }

    // 注册 room
    fOptsList.Store(roomToken, opt)

    // 房间创建者 自动加入
    joindata, joinstatus, joinok := Join(userToken, roomToken)
    if !joinok {
        fOptsList.Delete(roomToken)
        fUserList.Delete(userToken)
        return "New Room Failed!", Game_failed_response_, false
    }
    // fightLogger.Println(getOpts(roomToken))

    return joindata, joinstatus, true
}

func LeaveRoom(userToken, roomToken string) (string, bool) {
    user, ok1 := getUser(userToken)
    if !ok1 { return "You haven't login.", false }
    opt, ok2 := getOpts(roomToken)
    if !ok2 { return "Not found room.", false}
    opt.leave(user)
    return "Leave.", true
}

func Join(userToken, roomToken string) (interface{}, int, bool) {
    user, ok1 := getUser(userToken)
    if !ok1 { return nil, US_offline_, false }
    opt, ok2 := getOpts(roomToken)
    if !ok2 { return nil, RM_not_found_, false }
    
    status, joinok := opt.join(user)
    if joinok {
        return &JoinRet{
            Uid: user.id,
            PlayerNum: opt.playerNum,
            Row: opt.row,
            Col: opt.col,
            RoomToken: roomToken,
        },  status, true
    }

    return nil, status, false
}

func Login(userToken, userName string) (string, bool) {
    fUserList.Store(userToken, &fUser{
        id: _visitor_,
        userName:  userName,
        userToken: userToken,
        score : 0,
        energy: 0,
        status: US_online_,
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

func GetRoom(userToken string) string {
    user, ok := getUser(userToken)
    if !ok { return "" }
    return user.roomToken
}

func GetScoreBoard(roomToken string) (*map[string]int, int, bool) {
    opt, ok := getOpts(roomToken)
    if !ok { 
        sb, sbok := fScoreBoardMap.Load(roomToken)
        if sbok == true { return sb.(*map[string]int), Game_success_response_, true }
        return nil, RM_not_found_, false 
    }
    sb := make(map[string]int)
    for _, v := range opt.userInfo {
        sb[v.userName] = v.score;
    }
    return &sb, Game_success_response_, true
}

func GetEyeShot(userToken, roomToken string, loc Point) (*EyeShot, string, bool) {
    user, ok1 := getUser(userToken);
    if !ok1 { return nil, "Not login!", false }
    opt, ok2 := getOpts(user.roomToken)
    if !ok2 { return nil, "Not found room!", false }
    if !isPlaying(opt) { return nil, "Game is not start!", false }

    _qry, queryok := fQueryCountMap.Load(userToken)
    qry := _qry.(*fQueryCounter)
    if !queryok { return nil, "No query access!", false }
    if !qry.dec() { return nil, "Query count limited!", false }
    
    v, status, eyeok := opt.eyeshot(user, &fPoint{ x:loc.X, y:loc.Y })

    if eyeok == true {
        return v, "OK", true
    } else if status == Game_invalid_point_ {
        return nil, "Invalid Point!", false
    }
    return nil, "Unknown Error!", false
}

func Move(userToken, roomToken string, direction, radio int, loc Point) (*fPoint, *fPoint, bool) {
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

    src, dest, moveok := opt.move(user, direction, radio, &fPoint{ x:loc.X, y:loc.Y })
    if moveok {
        return src, dest, true
    }
    return nil, nil, false
}

// 重置每轮的查询次数限制
func resetQueryCount(tokens []string) {
    for _, v := range tokens {
        _qry, _ := fQueryCountMap.Load(v)
        qry := _qry.(*fQueryCounter)
        qry.reset()
    }
}

// 清除
func clearQueryCount(tokens []string) {
    for _, v := range tokens {
        fQueryCountMap.Delete(v)
    }
}

// 执行该轮的actionEvent
func doIt(roomToken string, actions []eventQ.EventEle, wsActions *WSAction) {
    if len(actions) <= 0 { return }
    var wsChange WSChange // websocket change数据
    for _, v := range(actions) {
        ac := v.Value.(ActionEvent) // v.Token => usertoken
        switch ac.Typ {
        case Action_move_:
            // fightLogger.Println("[doIt] Move ", v.Token)
            aMove := ac.Ac.(ActionMove)
            p1, p2, moveok := Move(v.Token, roomToken, aMove.Direction, aMove.Radio, aMove.Loc)
            if !moveok { continue }
            
            wsChange.Loc = append(wsChange.Loc, &WSPoint{ X: p1.x, Y:p1.y, M1:p1.m1, M2:p1.m2 })
            wsChange.Loc = append(wsChange.Loc, &WSPoint{ X: p2.x, Y:p2.y, M1:p2.m1, M2:p2.m2 })

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
    fightLogger.Println("Run")
    opt, ok1 := getOpts(roomToken)
    if !ok1 { return nil }
    mm := opt.m

    // 设置eventQueue 和 fQueryCount
    var tokens []string
    for _, v := range(opt.userInfo) {
        tokens = append(tokens, v.userToken)
        // 每轮查询最大值
        fQueryCountMap.Store(v.userToken, &fQueryCounter{ count: default_max_query })
    }
    eq.Initialize(tokens)
    // fightLogger.Println("[Run] EventQueue: ", eq)

    // 游戏总时长
    playTimer := time.NewTicker(default_play_timeout)
    // 每轮action时间间隔
    actionTimer := time.NewTicker(default_action_interval)

    gameEnd := make(chan bool)

    go func() {
        defer playTimer.Stop()
        defer actionTimer.Stop()
        defer ws.WSDestroy()

        loop_cnt        := 0
        global_add_cnt  := 0
        special_add_cnt := 0

        for {
            select {            
            case <-actionTimer.C:
                opt.mu.Lock()
                
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
                if !eq.Empty() { 
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

                    resetQueryCount(tokens) // 重设每轮最大查询次数
                
                } else { /* 已经没有玩家在房间内 */
                    opt.mu.Unlock()
                    // 清除查询次数限制
                    clearQueryCount(tokens)
                    gameEnd <- true
                    return
                }
                opt.mu.Unlock()

            default: /* nothing */
            }

            /* 时间到结束 */
            select {
            case <-playTimer.C:
                fightLogger.Println("[Run] Game End!")
                /* 保留最终榜单一定时间 */
                sb, _, _ := GetScoreBoard(roomToken)
                fScoreBoardMap.Store(roomToken, sb)
                go func(rtk string) {
                    <- time.After(ScoreBoardKeepTime)
                    fScoreBoardMap.Delete(rtk)
                }(roomToken)
                /* websocket发送游戏结束 */
                ws.WSBroadcast(&WSAction{ Typ: WSAction_game_end })
                /* 退出所有玩家 */
                for _, v := range(opt.userInfo) {
                    v.status = US_waiting_
                    LeaveRoom(v.userToken, roomToken)
                }
                // 清空事件队列
                eq.Clear()
                // 清除查询次数限制
                clearQueryCount(tokens)
                gameEnd <- true
                return

            default: /* nothing */
            }
        }
    }()
   
    return gameEnd
}

func WSGetGameInfo(roomToken string) *WSAction {
    opt, ok := getOpts(roomToken)
    if !ok { return nil }
    mm := opt.m
    u := &WSMapInfoRet {
        GameTime: default_play_timeout.Seconds(), // 游戏总时长
        Row: opt.row,
        Col: opt.col,
        RoomToken: opt.roomToken,
        M1:  &mm.m1,
        M2:  &mm.m2,
    }
    for k, v := range opt.userInfo {
        u.UserInfo = append(u.UserInfo, &WSUserInfoRet{
            Uid: k,
            Uname: v.userName,
            Score: v.score,
            Energy: v.energy,
            Status: v.status,
        })
        // fightLogger.Println(v.userName)
    }
    return &WSAction{
        Typ: WSAction_mapinfo,
        Value: u,
    }
}

func init() {
    flogFile, flogFileErr := os.OpenFile("fightlog.txt", os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0644)
    if flogFileErr != nil {
        panic(flogFileErr)
        return
    }
    fightLogger = log.New(flogFile, "[fight] ", log.Ldate | log.Ltime | log.Lshortfile)
}
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

func NewRoom(userToken string, tPlayerNum int, trow int, tcol int, tbarback int, tportal int, tbarrier int) (interface{}, int, bool) {
    _, ok := getUser(userToken)
    if !ok { return "You haven't login.", Game_failed_response_, false }

    // 可定制参数: 游戏人数 战场size
    if tPlayerNum == -1 { tPlayerNum = default_player_num }
    if trow == -1 { trow = default_row }
    if tcol == -1 { tcol = default_col }
    if tbarback == -1 { tbarback = default_barback }
    if tportal  == -1 { tportal  = default_portal  }
    if tbarrier == -1 { tbarrier = default_barrier }

    // 合法性校验
    if tPlayerNum > default_max_player_num || tPlayerNum <= 0 || trow > default_max_row || tcol > default_max_col || trow < default_min_row || tcol < default_min_col {
        return "Out of limit! Check room config!", Game_failed_response_, false
    }

    if tPlayerNum + tbarback + tportal + tbarrier > trow * tcol { return "Building too many!", Game_failed_response_, false }

    fightLogger.Println("tbarback: ",tbarback)
    fightLogger.Println("tportal: ",tportal)
    fightLogger.Println("tbarrier: ",tbarrier)

    roomToken := GenToken(userToken)
    opt := &fOpts {
        accId: _visitor_ + 1,
        playerNum: tPlayerNum,
        row: trow,
        col: tcol,
        barbackNum: tbarback,
        portalNum:  tportal,
        barrierNum: tbarrier,
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
            Id:  user.id,
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

/* 获取玩家所在的房间的roomtoken */
func GetRoom(userToken string) string {
    user, ok := getUser(userToken)
    if !ok { return "" }
    return user.roomToken
}

/* 获取得分榜 */
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

/* 
    两种可能 1.游戏人数不够未开始 2.房间不存在
    (int=>x, int=>y, int=>row, int=>col, bool)
*/
func IsStart(userToken, roomToken string) (int, int, int, int, bool) {
    user, ok := getUser(userToken)
    if !ok { return -1, -1, -1, -1, false }
    room, ok := getOpts(roomToken)
    if !ok { return -1, -1, -1, -1, false }
    if room.playing { return user.baseLoc.X, user.baseLoc.Y, room.row, room.col, true }
    return -1, -1, -1, -1, false
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
    } else if status == Game_not_belong_ {
        return nil, "Not belong to you!", false
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

/* 重置每轮的查询次数限制 */
func resetQueryCount(tokens []string) {
    for _, v := range tokens {
        _qry, _ := fQueryCountMap.Load(v)
        qry := _qry.(*fQueryCounter)
        qry.reset()
    }
}

/* 清除查询限制 */
func clearQueryCount(tokens []string) {
    for _, v := range tokens {
        fQueryCountMap.Delete(v)
    }
}

/* 
    执行该轮的actionEvent
    并 push 进 wsaction 队列
*/
func executeIt(roomToken string, actions []eventQ.EventEle, wsActionsOneLoop *WSAction) {
    if len(actions) <= 0 { return }
    var wsChange WSChange // websocket change数据
    /* 遍历执行每个action */
    for _, v := range(actions) {
        ac := v.Value.(ActionEvent) // v.Token => usertoken
        switch ac.Typ {
        case Action_move_:
            // fightLogger.Println("[executeIt] Move ", v.Token)
            aMove := ac.Ac.(ActionMove)
            p1, p2, moveok := Move(v.Token, roomToken, aMove.Direction, aMove.Radio, aMove.Loc)
            if !moveok { continue }
            
            wsChange.Loc = append(wsChange.Loc, &WSPoint{ X: p1.x, Y:p1.y, M1:p1.m1, M2:p1.m2 })
            wsChange.Loc = append(wsChange.Loc, &WSPoint{ X: p2.x, Y:p2.y, M1:p2.m1, M2:p2.m2 })

        // case Action_magic_:
        //     fightLogger.Println("[executeIt] Magic ", v.Token)
        default:
        }
    }
    wsActionsOneLoop.Typ = WSAction_normal_change
    wsActionsOneLoop.Value = wsChange
}


/* 
    游戏之心
    Return Result: 0=>"failed"  1=>"end" 
*/
func Run(roomToken string, eq *eventQ.EventQueue, ws *WSChannel) chan bool {
    fightLogger.Println("Run")
    opt, ok1 := getOpts(roomToken)
    if !ok1 { return nil }
    mm := opt.m

    /* 设置eventQueue 和 fQueryCount */
    var tokens []string
    for _, v := range(opt.userInfo) {
        tokens = append(tokens, v.userToken)
        // 每轮查询的最大值
        fQueryCountMap.Store(v.userToken, &fQueryCounter{ count: default_max_query })
    }
    eq.Initialize(tokens)

    playTimer := time.NewTicker(default_play_timeout) // 游戏总时长
    actionTimer := time.NewTicker(default_action_interval) // 每轮action时间间隔

    gameEnd := make(chan bool) // 游戏结束通知channel

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

                loop_cnt++ // 统计游戏共经过几轮操作
                global_add_cnt++;
                special_add_cnt++;
                
                var wsActionsOneLoop []*WSAction // websocket 操作队列
                
                /* 系统行为 */
                if global_add_cnt >= default_global_add {
                    autoGlobalAdd(opt, mm)
                    global_add_cnt = 0
                    wsActionsOneLoop = append(wsActionsOneLoop, &WSAction{
                        Typ: WSAction_global_add,
                    })
                }
                if special_add_cnt >= default_special_add {
                    autoSpecialAdd(mm)
                    special_add_cnt = 0
                    wsActionsOneLoop = append(wsActionsOneLoop, &WSAction{
                        Typ: WSAction_special_add,
                    })
                }
                /* 系统行为 */

                /* 
                    如果玩家 logout 退出房间 
                    会造成 usertoken 不一致 而无法重新加入房间
                */

                /* 玩家行为 */
                if !eq.Empty() {
                    actions := eq.Get() // 每轮的操作列表
                    if len(actions) > 0 {
                        var wsNormalAction WSAction // 普通的change
                        executeIt(roomToken, actions, &wsNormalAction)
                        wsActionsOneLoop = append(wsActionsOneLoop, &wsNormalAction)
                    }

                    if !ws.WSEmpty() {
                        for _, v := range wsActionsOneLoop {
                            ws.WSBroadcast(v)
                        }
                    }
                    resetQueryCount(tokens) // 重设每轮最大查询次数
                
                } else { /* 已经没有玩家在房间内 */
                    opt.mu.Unlock()
                    clearQueryCount(tokens) // 清除查询次数限制
                    gameEnd <- true // 发送游戏结束信号, 回收WSChannel EventQueue
                    return
                }
                /* 玩家行为 */

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
                clearQueryCount(tokens) // 清除查询次数限制
                gameEnd <- true // 发送游戏结束信号, 回收WSChannel EventQueue
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
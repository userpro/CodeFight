package fight

import (
    "encoding/hex"
    "math/rand"
    "crypto/md5"
    "time"
    "strconv"
)

func (qry *fQueryCounter) dec() bool {
    qry.mu.Lock()
    defer qry.mu.Unlock()
    if qry.count - 1 < 0 { return false }
    qry.count--
    return true
}

func (qry *fQueryCounter) reset() {
    qry.mu.Lock()
    qry.count = default_max_query
    qry.mu.Unlock()
}

/*  
    Return Result: 
    RM_playing_  RM_full_  RM_unauthorized_  RM_already_join_  RM_playing_  RM_joinable_
*/
func (opt *fOpts) join(user *fUser) (int, bool) {
    opt.mu.Lock()
    defer opt.mu.Unlock()
    if opt.playing == true { return RM_playing_, false }
    if len(opt.userInfo) > opt.playerNum { return RM_full_, false }
    if user.status != US_online_ { return RM_unauthorized_, false }
    if user.roomToken == opt.roomToken { return RM_already_join_, false }

    user.id = opt.accId
    user.roomToken = opt.roomToken
    user.score = 0; user.energy = 0

    opt.userInfo[opt.accId] = user
    opt.accId++

    /* 人数满足 游戏开始 */
    if len(opt.userInfo) == opt.playerNum { 
        genOk := opt.generator() 
        if !genOk { fightLogger.Println("[Join] generator Failed.") }
        opt.playing = true
        for _, v := range opt.userInfo {
            v.status = US_playing_
            v.score = 1 // 初始时 只有一个base
        }
        return RM_playing_, true
    }
    return RM_joinable_, true
}

/* 游戏开始 生成地图 分配玩家 */
func (opt *fOpts) generator() bool {
    fightLogger.Println("[generator] OK!")
    rand.Seed(time.Now().Unix())
    
    mp := new(fMap)
    mp.mu.Lock()
    defer mp.mu.Unlock()

    /* 随机数不应该小于最大值的1/2 */
    portalNum := rand.Intn(default_max_portal) + 1
    if portalNum < default_max_portal / 2 { portalNum *= 2 }

    barbackNum := rand.Intn(default_max_barback) + 1
    if barbackNum < default_max_barback / 2 { barbackNum *= 2 }

    barrierNum := rand.Intn(default_max_barrier) + 1
    if barrierNum < default_max_barrier / 2 { barrierNum *= 2 }

    /* 随机 portal 坐标 */
    cnt := 0
    for {
        if cnt >= portalNum { break }
        x := rand.Intn(opt.row)
        y := rand.Intn(opt.col)
        if (mp.m2[x][y] == _space_) {
            mp.m1[x][y] = default_portal_army
            mp.m2[x][y] |= _portal_
            mp.portal = append(mp.portal, Point{X:x,Y:y})
            cnt++;
        }
    }

    /* 随机 barback 坐标 */
    cnt = 0
    for {
        if cnt >= barbackNum { break }
        x := rand.Intn(opt.row)
        y := rand.Intn(opt.col)
        if (mp.m2[x][y] == _space_) {
            mp.m1[x][y] = default_barback_army
            mp.m2[x][y] |= _barback_
            mp.barback = append(mp.barback, Point{X:x,Y:y})
            cnt++
        }
    }

    /* 随机 barrier 坐标 */
    cnt = 0
    for {
        if cnt >= barrierNum { break }
        x := rand.Intn(opt.row)
        y := rand.Intn(opt.col)
        if (mp.m2[x][y] == _space_) {
            mp.m2[x][y] |= _barrier_
            mp.barrier = append(mp.barrier, Point{X:x,Y:y})
            cnt++
        }
    }

    /* 随机分配 base */
    cnt = 0
    for {
        if cnt >= opt.playerNum { break }
        x := rand.Intn(opt.row)
        y := rand.Intn(opt.col)
        if (mp.m2[x][y] == _space_) {
            mp.m1[x][y] = default_base_army
            mp.m2[x][y] |= _base_
            mp.base = append(mp.base, Point{X:x,Y:y})
            cnt++
        }
    }

    /* 随机 User 到 base */
    cnt = 0
    for k, v := range opt.userInfo {
        t := mp.base[cnt]
        mp.m2[t.X][t.Y] = setCellId(mp.m2[t.X][t.Y], k)
        // fmt.Println("[generator] ", getUserId(mp.m2[t.X][t.Y]), k)
        cnt++
        v.baseLoc.X = t.X; v.baseLoc.Y = t.Y
    }

    opt.m = mp
    return true
}

func (opt *fOpts) move(user *fUser, direction, radio int, src *fPoint) (*fPoint, *fPoint, bool) {
    mm := opt.m

    /* 判断移动是否合法 */
    if !checkLoc(opt, src) { 
        fightLogger.Println("[Move] Invalid current Point!")
        return nil,nil,false 
    }

    /* 调出判断 */
    src.m1 = mm.m1[src.x][src.y]
    src.m2 = mm.m2[src.x][src.y]
    srcUid  := getUserId(src.m2)

    if srcUid != user.id { 
        fightLogger.Println("[Move] Not belong to you! src: ", srcUid, " your id: ", user.id)
        return nil,nil,false 
    }
    if isBarrier(src.m2) { 
        fightLogger.Println("[Move] Can't move from, it's a barrier!")
        return nil,nil,false 
    }
    if src.m1 <= 1 { 
        fightLogger.Println("[Move] Don't have enough army!")
        return nil,nil,false 
    }

    /* 调入判断 */
    dest := getNextPoint(src, direction)
    if !checkLoc(opt, dest) { 
        fightLogger.Println("[Move] Invalid next Point!")
        return nil,nil,false 
    }
    dest.m1 = mm.m1[dest.x][dest.y]
    dest.m2 = mm.m2[dest.x][dest.y]

    destCell := dest.m2
    destUid  := getUserId(destCell)
    if isBarrier(destCell) { 
        fightLogger.Println("[Move] Can't move to!")
        return nil,nil,false 
    }

    // fightLogger.Println("[move] x:",x," y:",y," nx:",nx," ny:",ny)

    finalRadio := getFinalRadio(radio)
    // 军队调出{x, y}
    var t1 int
    switch finalRadio {
        case _all_:     t1 = src.m1 - 1
        case _half_:    t1 = src.m1 / 2
        case _quarter_: t1 = src.m1 / 4
    }
    
    src.m1 -= t1
    /* 领地内调动 */
    if srcUid == destUid {
        dest.m1 += t1
    } else {
        /* 敌对领域 attack */
        t2 := dest.m1
        // portal 防御提升
        if isPortal(destCell) { t2 = int(float32(t2) * default_portal_factor) }

        /* 成功占领 */
        if t1 >= t2 {
            t1 -= t2
            if t1 == 0 { t1 = 1 }
            // 更新target cell的兵力和id
            dest.m1 = t1
            dest.m2 = setCellId(destCell, user.id)
            // 更新自己score
            user.score++;
            // 如果是_system_直接改变阵营即可
            if !isSystem(destCell) {
                // 更新敌人score
                opt.userInfo[destUid].score--

                if isBase(destCell) { // base
                    someOneGameOver(user, opt)
                    mm.removeBase(Point{ dest.x, dest.y })
                } else if isPortal(destCell) { // portal
                    dest.m2 = setCellType(destCell, _space_)
                    mm.removePortal(Point{ dest.x, dest.y})
                }
            }
        } else {
            // 消耗
            dest.m1 -= t1
        }
    }

    // fightLogger.Println(src, dest)
    mm.m1[src.x][src.y] = src.m1
    mm.m1[dest.x][dest.y] = dest.m1
    mm.m2[dest.x][dest.y] = dest.m2

    return src, dest, true
}

// Return Result:
//  Game_invalid_point_  Game_success_response_
func (opt *fOpts) eyeshot(user *fUser, loc *fPoint) (*EyeShot, int, bool) {
    if !checkLoc(opt, loc) { return nil, Game_invalid_point_, false }

    mm := opt.m
    mm.mu.RLock()
    defer mm.mu.RUnlock()
    if user.id != getUserId(mm.m2[loc.x][loc.y]) { return nil, Game_not_belong_, false }

    v := &EyeShot{}
    ltx := loc.x - default_eye_level
    lty := loc.y - default_eye_level
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
    return v, Game_success_response_, true
}

func (opt *fOpts) leave(user *fUser) {
    opt.mu.Lock()
    delete(opt.userInfo, user.id)
    opt.mu.Unlock()
    // 一个 room 没有 player 的时候, 回收 room.
    if len(opt.userInfo) == 0 {
        fightLogger.Println("[LeaveRoom] Destroy! ")
        fOptsList.Delete(user.roomToken)
    }
    // 更新 User 信息
    user.id = _visitor_
    user.roomToken = ""
    user.score = 0
    user.energy = 0
    user.status = US_online_
}


func checkLoc(opt *fOpts, loc *fPoint) bool {
    if opt.playing == false { return false }
    return loc.x>=0 && loc.x<opt.row && loc.y>=0 && loc.y<opt.col
}

// 获得合法radio
func getFinalRadio(radio int) int {
    switch radio {
        case _all_:     return _all_
        case _half_:    return _half_
        case _quarter_: return _quarter_
        default:        return _all_
    }
}

// 获得下一个point
func getNextPoint(cur *fPoint, dir int) *fPoint {
    var nx, ny int
    switch dir {
        case _top_:     nx = cur.x - 1; ny = cur.y
        case _right_:   nx = cur.x;     ny = cur.y + 1
        case _bottom_:  nx = cur.x + 1; ny = cur.y
        case _left_:    nx = cur.x;     ny = cur.y - 1
        default: nx = cur.x - 1; ny = cur.y
    }
    return &fPoint{ x:nx, y:ny }
}


// 全局军队增加
func autoGlobalAdd(opt *fOpts, mm *fMap) {
    for x:=0; x<opt.row; x++ {
        for y:=0; y<opt.col; y++ {
            if mm.m1[x][y] != 0 { mm.m1[x][y]++ }
        }
    }
}

// 特殊建筑军队增加 ( base barback... )
func autoSpecialAdd(mm *fMap) {
    for _, v := range mm.base {
        mm.m1[v.X][v.Y]++
    }
    for _, v := range mm.barback {
        mm.m1[v.X][v.Y]++
    }
}

func getUser(userToken string) (*fUser, bool) {
    _v, ok := fUserList.Load(userToken)
    if !ok { return nil, false }
    return _v.(*fUser), true
}

func getOpts(roomToken string) (*fOpts, bool) {
    _v, ok := fOptsList.Load(roomToken)
    if !ok { return nil, false}
    return _v.(*fOpts), true
}

func isPlaying(opt *fOpts) bool { return opt.playing }


func (mm *fMap)removeBase(target Point) {
    for ix, v := range(mm.base) {
        if v == target {
            mm.base = append(mm.base[:ix], mm.base[ix+1:]...)
            break
        }
    }
}

func (mm *fMap)removePortal(target Point) {
    for ix, v := range(mm.portal) {
        if v == target {
            mm.portal = append(mm.portal[:ix], mm.portal[ix+1:]...)
            break
        }
    }
}

func (mm *fMap)removeBarback(target Point) {
    for ix, v := range(mm.barback) {
        if v == target {
            mm.barback = append(mm.barback[:ix], mm.barback[ix+1:]...)
            break
        }
    }
}

func someOneGameOver(user *fUser, opt *fOpts) {
    mm := opt.m
    user.status = US_lose_
    for i:=0; i<opt.row; i++ {
        for j:=0; j<opt.col; j++ {
            if getUserId(mm.m2[i][j]) == user.id {
                mm.m1[i][j] = 0
                mm.m2[i][j] = setCellId(mm.m2[i][j], _system_);
            }
        }
    }
}

func GenToken(str string) string {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    tmp_s := str + strconv.FormatInt(r.Int63(), 10)
    md5_s := md5.New()
    md5_s.Write([]byte(tmp_s))
    return hex.EncodeToString(md5_s.Sum(nil))
}
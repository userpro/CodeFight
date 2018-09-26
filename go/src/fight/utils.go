package fight

import (
    "encoding/hex"
    "math/rand"
    "crypto/md5"
    "time"
    "strconv"
)

func GenToken(str string) string {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    tmp_s := str + strconv.FormatInt(r.Int63(), 10)
    md5_s := md5.New()
    md5_s.Write([]byte(tmp_s))
    return hex.EncodeToString(md5_s.Sum(nil))
}

func checkLoc(roomToken string, loc Point) bool { 
    opt, ok := getOpts(roomToken)
    if !ok { return false }
    return loc.X>=0 && loc.X<opt.row && loc.Y>=0 && loc.Y<opt.col
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


// 游戏开始 生成地图 分配玩家
func (fopt *fOpts) generator() bool {
    fightLogger.Println("[generator] OK!")
    rand.Seed(time.Now().Unix())
    var mp fMap

    // 随机数不应该小于最大值的1/2
    portalNum := rand.Intn(default_max_portal) + 1
    if portalNum < default_max_portal / 2 { portalNum *= 2 }

    barbackNum := rand.Intn(default_max_barback) + 1
    if barbackNum < default_max_barback / 2 { barbackNum *= 2 }

    barrierNum := rand.Intn(default_max_barrier) + 1
    if barrierNum < default_max_barrier / 2 { barrierNum *= 2 }

    // 随机 portal 坐标
    cnt := 0
    for {
        if cnt >= portalNum { break }
        x := rand.Intn(fopt.row)
        y := rand.Intn(fopt.col)
        if (mp.m2[x][y] == _space_) {
            mp.m1[x][y] = default_portal_army
            mp.m2[x][y] |= _portal_
            mp.portal = append(mp.portal, Point{X:x,Y:y})
            cnt++;
        }
    }

    // 随机 barback 坐标
    cnt = 0
    for {
        if cnt >= barbackNum { break }
        x := rand.Intn(fopt.row)
        y := rand.Intn(fopt.col)
        if (mp.m2[x][y] == _space_) {
            mp.m1[x][y] = default_barback_army
            mp.m2[x][y] |= _barback_
            mp.barback = append(mp.barback, Point{X:x,Y:y})
            cnt++
        }
    }

    // 随机 barrier 坐标
    cnt = 0
    for {
        if cnt >= barrierNum { break }
        x := rand.Intn(fopt.row)
        y := rand.Intn(fopt.col)
        if (mp.m2[x][y] == _space_) {
            mp.m2[x][y] |= _barrier_
            mp.barrier = append(mp.barrier, Point{X:x,Y:y})
            cnt++
        }
    }

    // 随机分配 base
    cnt = 0
    for {
        if cnt >= fopt.playerNum { break }
        x := rand.Intn(fopt.row)
        y := rand.Intn(fopt.col)
        if (mp.m2[x][y] == _space_) {
            mp.m1[x][y] = default_base_army
            mp.m2[x][y] |= _base_
            mp.base = append(mp.base, Point{X:x,Y:y})
            cnt++
        }
    }

    // 随机 User 到 base
    cnt = 0
    for k, v := range fopt.userToken {
        t := mp.base[cnt]
        mp.m2[t.X][t.Y] = setCellId(mp.m2[t.X][t.Y], k)
        // fmt.Println("[generator] ", getUserId(mp.m2[t.X][t.Y]), k)
        cnt++
        v.baseLoc.X = t.X; v.baseLoc.Y = t.Y
    }

    // for i:=0;i<fopt.row;i++ {
    //     for j:=0;j<fopt.col;j++ {
    //         fmt.Printf("%08b ",mp.m2[i][j])
    //     }
    //     fmt.Println("")
    // }
    // fmt.Printf("portalNum: %d  barbackNum: %d  barrierNum: %d\n", portalNum, barbackNum, barrierNum)

    fopt.m = &mp
    fMapList.Store(fopt.roomToken, &mp)
    return true
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

func getMap(roomToken string) (*fMap, bool) {
    _v, ok := fMapList.Load(roomToken)
    if !ok { return nil, false }
    return _v.(*fMap), true
}

func isPlaying(opt *fOpts) bool { return opt.playing }

// 两种可能 1.游戏人数不够未开始 2.房间不存在
// (int=>x, int=>y, bool)
func IsStart(userToken, roomToken string) (int, int, bool) {
    user, ok := getUser(userToken)
    if !ok { return -1, -1, false }
    room, ok := getOpts(roomToken)
    if !ok { return -1, -1, false }
    if room.playing { return user.baseLoc.X, user.baseLoc.Y, true }
    return -1, -1, false
}

func (mm *fMap)removeBase(target Point) {
    for ix, v := range(mm.base) {
        if v == target {
            mm.base = append(mm.base[:ix], mm.base[ix+1:]...)
        }
    }
}

func (mm *fMap)removePortal(target Point) {
    for ix, v := range(mm.portal) {
        if v == target {
            mm.portal = append(mm.portal[:ix], mm.portal[ix+1:]...)
        }
    }
}

func (mm *fMap)removeBarback(target Point) {
    for ix, v := range(mm.barback) {
        if v == target {
            mm.barback = append(mm.barback[:ix], mm.barback[ix+1:]...)
        }
    }
}

func someOneGameOver(user *fUser, opt *fOpts, mm *fMap) {
    user.status = US_lose_
    delete(opt.userToken, user.id)
    for i:=0; i<opt.row; i++ {
        for j:=0; j<opt.col; j++ {
            if getUserId(mm.m2[i][j]) == user.id {
                mm.m1[i][j] = 0
                mm.m2[i][j] = setCellId(mm.m2[i][j], _system_);
            }
        }
    }
}

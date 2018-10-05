package fight

import "time"

/* websocket */
const (
    WS_CHANNEL_BUFFER_SIZE  = 50 // websocket channel buffer
    WSAction_game_end       = 0 // 游戏结束
    WSAction_mapinfo        = 1 // 获取map信息
    WSAction_normal_change  = 2 // 获取普通改变 (Move, ...)
    WSAction_global_add     = 3 // 全局军队增加
    WSAction_special_add    = 4 // 特殊建筑军队增加
)

/* 状态 */
const (
    /* 用户状态 */
    US_offline_ = 0 // 未登录
    US_online_  = 1 // 已登录
    US_waiting_ = 2 // 已加入房间 等待游戏开始
    US_playing_ = 3 // 游戏中
    US_win_     = 4
    US_lose_    = 5

    /* 房间状态 */
    RM_not_found_ = 31
    RM_joinable_  = 32
    RM_full_      = 33
    RM_playing_   = 34
    RM_unauthorized_ = 35 // 无权限加入
    RM_already_join_ = 36 // 已经加入该房间

    Game_success_response_ = 51
    Game_failed_response_  = 52
    Game_invalid_point_    = 53 // 无效点
    Game_not_belong_       = 54
)

/* 游戏设置 */
const (
    ScoreBoardKeepTime = time.Second * 10 // 得分榜在游戏结束后对保持时间
    Default_login_timeout = time.Minute * 5   // 登录有效时限 (hour)
    Default_room_timeout  = 3   // 房间等待时限 (s)
    Default_player_num  = 16    // 房间玩家数量
    Default_col = 100
    Default_row = 100
)

/* 游戏设置 */
const (
    default_action_interval = time.Millisecond * 200 // 每轮action的间隔
    default_play_timeout  = time.Second * 600  // 游戏总时限 (s)

    default_max_portal  = 10  // 地图 portal 数量
    default_max_barback = 5   // 地图 barback 数量
    default_max_barrier = 50  // 地图 barrier 数量
    
    default_portal_army = 20  // 初始 portal 军队数量
    default_barback_army= 20  // 初始 barback 军队数量
    default_base_army   = 20  // 初始 base 军队数量
    default_global_add  = 6   // 全局增加 轮数间隔
    default_special_add = 2   // 全局特殊建筑增加(base, portal)
    default_portal_factor float32 = 1.5 // portal 防御提升因子

    default_max_query   = 8   // 每轮每个 player 最多的查询的次数

    /* 视野范围 (从中心坐标往外多少个距离1的矩形) 
    例如 2 => (2+1+2)*(2+1+2) */
    default_eye_level   = 2
)

/* 游戏内部预定义 */
const (
    /* 地形 Cell Type */
    _space_   = 0x00 // 000- ----
    _base_    = 0x20 // 001- ----
    _barback_ = 0x40 // 010- ----
    _portal_  = 0x60 // 011- ----
    _barrier_ = 0x80 // 100- ----
    /* 阵营 User Id */
    _system_  = 0x00 // ---0 0000
    _visitor_ = 0x01 // ---0 0001

    _user_mask_ = 0x1f // 0001 1111
    _type_mask_ = 0xe0 // 1110 0000

    /* direction */
    _top_     = 1
    _right_   = 2
    _bottom_  = 3
    _left_    = 4

    /* army radio */
    _all_     = 1
    _half_    = 2
    _quarter_ = 3
)

func isSpace(cell byte) bool { return cell&_type_mask_==_space_ }
func isBase(cell byte) bool { return cell&_type_mask_==_base_ }
func isBarback(cell byte) bool { return cell&_type_mask_==_barback_ }
func isPortal(cell byte) bool { return cell&_type_mask_==_portal_ }
func isBarrier(cell byte) bool { return cell&_type_mask_==_barrier_ }
func isSystem(cell byte) bool { return cell&_user_mask_==_system_ }
func setCellType(cell, typ byte) byte { return cell&_user_mask_|typ }

func setCellId(cell, uid byte) byte { return cell&_type_mask_|uid }
func getUserId(cell byte) byte { return cell&_user_mask_ }
package fight

import "time"

/* websocket 设置 */
const (
    WS_CHANNEL_BUFFER_SIZE  = 20 // websocket channel buffer
    WSAction_mapinfo        = 1
    WSAction_normal_change  = 2
    WSAction_global_add     = 3 // 全局军队增加
    WSAction_special_add    = 4 // 特殊建筑军队增加
)

const (
    /* 游戏设置 */
    Default_login_timeout = time.Hour * 1   // 登录有效时限 (hour)
    Default_room_timeout  = 3   // 房间等待时限 (s)
    Default_player_num  = 16    // 房间玩家数量
    Default_col = 30
    Default_row = 30

    /* 用户状态 */
    US_offline_ = -1
    US_wait_    = 0
    US_playing_ = 1
    US_win_     = 2
    US_lose_    = 3
)

const (
    default_action_interval = time.Second * 1 // 每轮action的间隔
    default_play_timeout  = time.Second * 100  // 游戏总时限 (s)
    // 
    default_max_portal  = 10  // 地图能量泉
    default_max_barback = 5   // 地图兵营
    default_max_barrier = 10  // 地图障碍物
    /* 视野范围 (从中心坐标往外多少个距离1的矩形) 
    例如 2 => (2+1+2)*(2+1+2) */
    default_eye_level   = 2
    default_portal_army = 50  // 初始 portal 军队数量
    default_barback_army= 50  // 初始 barback 军队数量
    default_base_army   = 50  // 初始 base 军队数量
    default_global_add  = 5   // 全局增加 轮数间隔
    default_special_add = 2   // 全局特殊建筑增加(base portal)
    default_portal_factor float32 = 1.5 // portal 防御提升因子
)

const (
    /* 地形 */
    _space_   = 0x00
    _base_    = 0x20
    _barback_ = 0x40
    _portal_  = 0x60
    _barrier_ = 0x80
    /* 阵营fUser id */
    _system_  = 0x00
    _visitor_ = 0x01

    _user_mask_ = 0x1f // 后5位
    _type_mask_ = 0xe0 // 前3位

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
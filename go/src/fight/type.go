package fight

type (
    Point struct {
        X int `json:"x"`
        Y int `json:"y"`
    }

    fPoint struct {
        x, y, w int
    }

    fUser struct {
        id            byte // 在room中的id
        score, energy int // score => cell的个数
        userToken     string
        roomToken     string
        status        int // 0=>wait  1=>playing 2=>lose 3=>win
        baseLoc       Point
    }

    fOpts struct {
        accId     byte // 当前最小未分配ID (严格递增)
        playerNum int
        row, col  int
        playing   bool
        roomToken string
        userToken map[byte]*fUser // [id]*fUser
        m         *fMap
    }

    // m2 => 8字节 
    // 前三位地形类型 后五位阵营
    fMap struct {
        roomToken string
        m1    [Default_row][Default_col]int
        m2    [Default_row][Default_col]byte

        base, barback, portal, barrier []Point
    }

    /* net */
    EyeShot struct {
        M1 [default_eye_level * 2 + 1][default_eye_level * 2 + 1]int
        M2 [default_eye_level * 2 + 1][default_eye_level * 2 + 1]byte
    }
    /* net */
)
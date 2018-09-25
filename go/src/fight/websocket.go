package fight

import "sync"

type (
    /* websocket */
    WSMapInfoRet struct {
        Row       int     `json:"row,omitempty"`
        Col       int     `json:"col,omitempty"`
        RoomToken string  `json:"roomtoken,omitempty"`
        M1    *[Default_row][Default_col]int  `json:"m1,omitempty"`
        M2    *[Default_row][Default_col]byte `json:"m2,omitempty"`

        UserInfo  []*WSUserInfoRet `json:"userinfo,omitempty"`
    }

    WSUserInfoRet struct {
        Uid    byte   `json:"id,omitempty"`
        Utoken string `json:"token,omitempty"`
        Score  int    `json:"score,omitempty"`
        Energy int    `json:"energy,omitempty"`
        Status int    `json:"status,omitempty"`
    }

    WSPoint struct {
        X int `json:"x"`
        Y int `json:"y"`
        W int `json:"w"`
    }

    WSChange struct {
        Loc []*WSPoint `json:"loc,omitempty"`
    }

    WSAction struct {
        Typ   int         `json:"type"`
        Value interface{} `json:"value,omitempty"`
    }

    WSChannel struct {
        count int
        Ch    sync.Map // map[string]chan *WSAction
    }
    /* websocket */
)

func WSNew() *WSChannel {
    return new(WSChannel)
}

func (ws *WSChannel)WSEmpty() bool {
    if ws.count == 0 { return true }
    return false
}

func (ws *WSChannel)WSRegister(token string) chan *WSAction {
    ws.count++
    ch := make(chan *WSAction, WS_CHANNEL_BUFFER_SIZE)
    ws.Ch.Store(token, ch) // 10 => buffer size
    return ch
}

func (ws *WSChannel)WSCancel(token string) {
    ch, ok := ws.Ch.Load(token)
    if !ok { return }
    _ch := ch.(chan *WSAction)
    close(_ch)
    ws.Ch.Delete(token)
    ws.count--
}

func (ws *WSChannel)WSBroadcast(data *WSAction) {
    ws.Ch.Range(func(k, v interface{}) bool {
        _v := v.(chan *WSAction)
        _v <- data
        return true
    })
}

func (ws *WSChannel)WSDestroy() {
    ws.Ch.Range(func(k, v interface{}) bool {
        _v := v.(chan *WSAction)
        close(_v)
        return true
    })
    ws.count = 0
}


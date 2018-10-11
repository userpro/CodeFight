package net

import (
    "net/http"
    "encoding/json"

    "github.com/labstack/echo"
    "golang.org/x/net/websocket"

    fight "../fight"
)

// GET
// QueryParam
func view(c echo.Context) error {
    rtk := c.Param("roomtoken")
    _, ok := eventQMap.Load(rtk)
    if !ok {
        return c.HTML(http.StatusOK, "<h1>Not found room.</h1>")
    }
    return c.Render(http.StatusOK, "view.html", rtk)
}

func wsocketView(c echo.Context) error {
    websocket.Handler(func(ws *websocket.Conn) {
        defer ws.Close()
        // 接收roomtoken
        rtk := ""
        err := websocket.Message.Receive(ws, &rtk)
        if err != nil {
            c.Logger().Error(err)
            return
        }

        // 获取该房间的基本信息 mapinfo userinfo
        gameInfo := fight.WSGetGameInfo(rtk)
        if gameInfo == nil {
            websocket.Message.Send(ws, "StatusUnauthorized!")
            return
        }
        gameInfoObj, gameInfoErr := json.Marshal(gameInfo)
        if gameInfoErr != nil {
            c.Logger().Error(gameInfoErr)
            return
        }
        // 发送房间基本信息
        err = websocket.Message.Send(ws, string(gameInfoObj))
        if err != nil {
            c.Logger().Error(err)
            return
        }

        // 每个访问页面都有一个唯一的token
        token := fight.GenToken(rtk)
        // 获取该房间的websocket channel
        _wsc, wscok := WSChannelMap.Load(rtk)
        if !wscok { return }
        wsc := _wsc.(*fight.WSChannel)
        ch := wsc.WSRegister(token)

        for {
            sendData := <- ch
            if sendData == nil { 
                // 已无数据发送 游戏结束
                return 
            }
            // fmt.Println("[wsocketView] ",sendData)
            // Write
            sendDataObj, sendDataErr := json.Marshal(sendData)
            if sendDataErr != nil {
                c.Logger().Error(sendDataErr)
                return
            }
            err = websocket.Message.Send(ws, string(sendDataObj))
            if err != nil {
                c.Logger().Error(err)
                // 发送数据失败 客户端关闭
                wsc.WSCancel(token)
                return
            }
        }
    }).ServeHTTP(c.Response(), c.Request())
    return nil
}
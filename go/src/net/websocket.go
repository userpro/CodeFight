package net

import (
    // "fmt"
    "time"
    "net/http"
    "encoding/json"

    "github.com/labstack/echo"
    "golang.org/x/net/websocket"

    fight "../fight"
    viewTemplate "../view"
)

// type 

// GET
// QueryParam
func view(c echo.Context) error {
    rtk := c.Param("roomtoken")
    _, ok := eventQMap.Load(rtk)
    if !ok {
        return c.HTML(http.StatusOK, "<h1>Not found room.</h1>")
    }
    return c.HTML(http.StatusOK, viewTemplate.GetContent(rtk))
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
        gameInfo := fight.GetGameInfo(rtk)
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

        // sendInfo := &RespInfo {
        //     Message: "OK",
        //     Status: 1,
        // }
        for {
            sendData := <- ch
            if sendData == nil { return }
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
                return
            }

            time.Sleep(time.Second * 1)
        }
    }).ServeHTTP(c.Response(), c.Request())
    return nil
}
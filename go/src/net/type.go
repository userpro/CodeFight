package net

import (
    "io"
    "html/template"
    "github.com/labstack/echo"
    fight "../fight"
)

type (
    RespInfo struct {
        Message   string `json:"message,omitempty"`
        Status    int    `json:"status"`
    }

    RoomOptRet struct {
        Row       int    `json:"row,omitempty"`
        Col       int    `json:"col,omitempty"`
        Playernum   int    `json:"playernum,omitempty"`
    }

    netUserInfo struct {
        UserToken string
        RoomToken string
        Id    byte
        Uname string
        RespInfo
    }

    netUserRet struct {
        UserToken string `json:"usertoken,omitempty"`
        RoomToken string `json:"roomtoken,omitempty"`
        RoomOptRet
        RespInfo
    }

    netMoveReq struct {
        UserToken string `json:"usertoken"`
        RoomToken string `json:"roomtoken"`
        Radio     int    `json:"radio"`
        Direction int    `json:"direction"`
        Loc fight.Point  `json:"loc"`
    }

    netMoveRet struct {
        Length    int    `json:"length"`
        RespInfo
    }

    netQueryReq struct {
        UserToken string `json:"usertoken"`
        RoomToken string `json:"roomtoken"`
        Loc fight.Point  `json:"loc"`
    }

    netQueryRet struct {
        Eye       *fight.NetEyeShot `json:"eyeshot,omitempty"`
        RespInfo
    }

    netStartRet struct {
        X int `json:"x"`
        Y int `json:"y"`
        RespInfo
    }

    netScoreBoardRet struct {
        Result    *map[string]int `json:"scoreboard,omitempty"`
        RespInfo
    }
)

/* middleware */
/* html template render */
type Template struct {
    templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface {}, c echo.Context) error {
    return t.templates.ExecuteTemplate(w, name, data)
}
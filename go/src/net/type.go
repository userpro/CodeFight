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

    netUserInfo struct {
        Id        byte   `json:"id"`
        Name      string `json:"name"`
        UserToken string `json:"usertoken"`
        RoomToken string `json:"roomtoken"`
        RespInfo
    }

    netJoinRet  struct {
        Id         byte   `json:"id"`
        Name       string `json:"name"`
        UserToken  string `json:"usertoken"`
        RoomToken  string `json:"roomtoken"`
        Row       int     `json:"row"`
        Col       int     `json:"col"`
        Playernum int     `json:"playernum,omitempty"`
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
        Eye       *fight.EyeShot `json:"eyeshot,omitempty"`
        RespInfo
    }

    netStartRet struct {
        X int `json:"x"`
        Y int `json:"y"`
        Row int `json:"row"`
        Col int `json:"col"`
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
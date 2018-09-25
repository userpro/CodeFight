package net

import (
    "fmt"
    "time"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

type (
    DbWorker struct {
        Dsn string
        Db  *sql.DB
        UserInfo userTB
    }

    userTB struct {
        Id       int64
        Name     string
        Pass     string
        Email    string
        Status   int64
        CreateAt time.Time
    }
)

func (dbw *DbWorker)insertData(username, password string, email string) int64 {
    stmt, err := dbw.Db.Prepare(`INSERT INTO userinfo (username, password, email, CreateAt) VALUES (?, ?, ?, ?)`)
    defer stmt.Close()
    if err != nil {
        fmt.Println(err)
        return -1
    }

    ret, err := stmt.Exec(username, password, email, time.Now())
    if err != nil {
        fmt.Printf("[DB-insert] Error: %v\n", err)
        return -1
    }
    LastInsertId, err := ret.LastInsertId();
    if err == nil {
        fmt.Println("[DB-insert] LastInsertId: ", LastInsertId)
    } else {
        return -1
    }
    if RowsAffected, err := ret.RowsAffected(); err == nil {
        fmt.Println("[DB-insert] RowsAffected: ", RowsAffected)
    } else {
        return -1
    }
    return LastInsertId
}

func (dbw *DbWorker)QueryDataPre() {
    dbw.UserInfo = userTB{}
}

func (dbw *DbWorker)queryData(username, password string) bool {
    stmt, err := dbw.Db.Prepare(`SELECT * FROM userinfo WHERE username = ? AND password = ?`)
    if err != nil {
        fmt.Println("[DB-queryData] ", err)
        return false
    }

    dbw.QueryDataPre()
    err = stmt.QueryRow(username, password).Scan(&dbw.UserInfo.Id, &dbw.UserInfo.Name, &dbw.UserInfo.Pass, &dbw.UserInfo.Email, &dbw.UserInfo.Status, &dbw.UserInfo.CreateAt)
    if err != nil {
        fmt.Println("[DB-queryData]", err)
        return false
    }
    return true
}

func (dbw *DbWorker)updateData(id int64, username, password string, email string) bool {
    stmt, _ := dbw.Db.Prepare("UPDATE userinfo SET username=?, password=? email=? WHERE id=?")
    defer stmt.Close()

    ret, err := stmt.Exec(username, password, email, id)
    if err != nil {
        fmt.Println("[DB-updateData]", err)
        return false
    }
    affectCnt, _ := ret.RowsAffected()
    fmt.Println("[DB-updateData] RowsAffected: ", affectCnt)
    return true
}

func (dbw *DbWorker)deleteData(username, password string) {
    stmt, _ := dbw.Db.Prepare("DELETE FROM userinfo WHERE username=? AND password=?")
    defer stmt.Close()
    stmt.Exec(username, password)
}
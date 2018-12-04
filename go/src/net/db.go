package net

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type (
	DbWorker struct {
		Dsn      string
		Db       *sql.DB
		UserInfo userTB
	}

	userTB struct {
		Id       int64
		Name     string
		Pass     string
		Email    string
		Status   int64
		CreateAt time.Time
		UpdateAt time.Time
	}
)

func (dbw *DbWorker) InsertData(username, password string, email string) bool {
	stmt, _ := dbw.Db.Prepare(`INSERT INTO userinfo (username, password, email) VALUES (?, ?, ?)`)
	defer stmt.Close()

	/*ret*/
	_, err := stmt.Exec(username, password, email)
	if err != nil {
		dbLogger.Printf("[InsertErr] %v\n", err)
		return false
	}
	// LastInsertId, err := ret.LastInsertId();
	// if err == nil {
	//     fmt.Println("[DB-insert] LastInsertId: ", LastInsertId)
	// } else {
	//     return false
	// }
	// if RowsAffected, err := ret.RowsAffected(); err == nil {
	//     fmt.Println("[DB-insert] RowsAffected: ", RowsAffected)
	// } else {
	//     return false
	// }
	return true
}

func (dbw *DbWorker) queryDataPre() {
	dbw.UserInfo = userTB{}
}

func (dbw *DbWorker) QueryData(username, password string) bool {
	stmt, err := dbw.Db.Prepare(`SELECT * FROM userinfo WHERE username = ? AND password = ?`)
	if err != nil {
		dbLogger.Println("[QueryDataErr] ", err)
		return false
	}

	dbw.queryDataPre()
	err = stmt.QueryRow(username, password).Scan(&dbw.UserInfo.Id, &dbw.UserInfo.Name, &dbw.UserInfo.Pass, &dbw.UserInfo.Email, &dbw.UserInfo.Status, &dbw.UserInfo.CreateAt, &dbw.UserInfo.UpdateAt)
	if err != nil {
		dbLogger.Println("[QueryDataErr]", err)
		return false
	}
	return true
}

func (dbw *DbWorker) UpdateData(newUsername, newPassword string, newEmail string, username, password string) bool {
	stmt, err := dbw.Db.Prepare(`UPDATE userinfo SET username = ?, password = ?, email = ? WHERE username = ? AND password = ?`)
	if err != nil {
		dbLogger.Println(err)
	}
	defer stmt.Close()

	/*ret*/
	_, err = stmt.Exec(newUsername, newPassword, newEmail, username, password)
	if err != nil {
		dbLogger.Println("[UpdateDataErr]", err)
		return false
	}
	// affectCnt, _ := ret.RowsAffected()
	// fmt.Println("[DB-UpdateData] RowsAffected: ", affectCnt)
	return true
}

func (dbw *DbWorker) DeleteData(username, password string) {
	stmt, _ := dbw.Db.Prepare("DELETE FROM userinfo WHERE username=? AND password=?")
	defer stmt.Close()
	stmt.Exec(username, password)
}

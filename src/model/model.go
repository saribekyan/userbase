package model

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"model/auth"
)

var (
	UserbaseDB *sql.DB
	Sessions   *auth.SessionManager
	Users      *auth.UsersManager
)

type DatabaseInfo struct {
	Type string
	Path string
}

func Configure(dbinfo DatabaseInfo) {
	var err error
	UserbaseDB, err = sql.Open(dbinfo.Type, dbinfo.Path)
	if err != nil {
		panic(err)
	}
	auth.Configure(UserbaseDB)
}

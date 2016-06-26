package auth

import (
	"database/sql"
)

type UsersManager struct {
	db *sql.DB
}

func (um *UsersManager) AddUser(username string, password string) (bool, string) {
	userExists, err := um.db.Prepare(
		`SELECT EXISTS (
		    SELECT *
                    FROM users
                    WHERE username=?)`)
	defer userExists.Close()
	check(err)

	var exists bool
	err = userExists.QueryRow(username).Scan(&exists)
	check(err)

	if exists {
		return false, "Username already exists."
	}

	userAdd, err := um.db.Prepare(`
                INSERT INTO users (username, derived_key, salt)
                            VALUES (?, ?, ?)`)
	defer userAdd.Close()
	check(err)

	salt := randomString(SALT_LEN)
	dKey := deriveKey(password, salt)

	_, err = userAdd.Exec(username, dKey, salt)
	check(err)
	return true, ""
}

func (um *UsersManager) AuthenticateLogin(username string, password string) bool {
	getUser, err := um.db.Prepare(`
                SELECT derived_key, salt
                FROM users
                WHERE username=?`)
	defer getUser.Close()
	check(err)

	var dKey, salt string
	err = getUser.QueryRow(username).Scan(&dKey, &salt)
	if err == sql.ErrNoRows {
		return false
	}
	check(err)

	return deriveKey(password, salt) == dKey
}

func MakeUM(db *sql.DB) (um *UsersManager) {
	createUsersDB, err := db.Prepare(`
	    CREATE TABLE IF NOT EXISTS users (
		    username TEXT NOT NULL PRIMARY KEY,
	            derived_key TEXT,
	            salt TEXT
	    );`)
	defer createUsersDB.Close()
	check(err)

	_, err = createUsersDB.Exec()
	check(err)
	return &UsersManager{db: db}
}

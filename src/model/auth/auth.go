package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"golang.org/x/crypto/scrypt"
	"time"
)

const (
	SESSION_TIMEOUT = 1 * time.Minute
	SESSION_KEY_LEN = 64
)

const ( // salting
	SALT_LEN = 64
	DKEY_LEN = 64
	DKEY_N   = 1 << 14
	DKEY_r   = 8
	DKEY_p   = 1
)

var (
	Sessions *SessionManager
	Users    *UsersManager
)

// common functions for authentication
func randomString(len int) string {
	bytes := make([]byte, len)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

func deriveKey(password string, salt string) string {
	dkey, err := scrypt.Key([]byte(password), []byte(salt), DKEY_N, DKEY_r, DKEY_p, DKEY_LEN)
	check(err)
	return string(dkey)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func Configure(db *sql.DB) {
	Sessions = MakeSM()
	Users = MakeUM(db)
}

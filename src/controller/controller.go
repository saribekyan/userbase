package controller

import (
	"html/template"
	"model/auth"
	"net/http"
	"path"
	"time"
)

var (
	templatesBase string
)

const (
	COOKIE_NAME = "userbase_session_cookie"
)

func index(w http.ResponseWriter, r *http.Request) {
	var (
		isLogged     bool
		username     string
		sessionKey   string
		loginMessage string
	)

	if r.Method != "GET" && r.Method != "POST" {
		panic("Request not a GET or POST")
		return
	}

	isLogged = false

	cookie, err := r.Cookie(COOKIE_NAME)
	if err != http.ErrNoCookie {
		sessionKey = cookie.Value
		username, isLogged = auth.Sessions.AuthenticateSessionKey(sessionKey)
		loginMessage = ""
	}
	if !isLogged && r.Method == "POST" { // the cookie didn't work and username/password are sent
		r.ParseForm()

		username = template.HTMLEscapeString(r.Form.Get("username"))
		password := template.HTMLEscapeString(r.Form.Get("password"))

		isLogged = auth.Users.AuthenticateLogin(username, password)
		if isLogged {
			sessionKey = auth.Sessions.NewSession(username) // just logged in because there was no valid cookie
		} else {
			loginMessage = "Wrong username or password"
		}
	}

	if isLogged {
		cookie := http.Cookie{
			Name:    COOKIE_NAME,
			Value:   sessionKey,
			Expires: time.Now().Add(auth.SESSION_TIMEOUT),
			Secure:  true}
		http.SetCookie(w, &cookie)

		t, _ := template.ParseFiles(path.Join(templatesBase, "page.html"))

		t.Execute(w, &struct {
			User        string
			NumSessions int
		}{username,
			auth.Sessions.NSessionsOfUser(username)})
	} else {
		t, _ := template.ParseFiles(path.Join(templatesBase, "index.html"))
		t.Execute(w, &struct{ WrongLogin string }{loginMessage})
	}
}

func signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles(path.Join(templatesBase, "signup.html"))
		t.Execute(w, nil)
	} else if r.Method == "POST" {
		r.ParseForm()

		username := template.HTMLEscapeString(r.Form.Get("username"))
		password := template.HTMLEscapeString(r.Form.Get("password"))

		if ok, err := auth.Users.AddUser(username, password); ok {
			t, _ := template.ParseFiles(path.Join(templatesBase, "index.html"))
			t.Execute(w, &struct{ User string }{username})
		} else {
			t, _ := template.ParseFiles(path.Join(templatesBase, "signup.html"))
			t.Execute(w, &struct{ WrongUsername string }{err})
		}
	}
}

func Configure(templatesPath string) {
	templatesBase = templatesPath
	http.HandleFunc("/", index)
	http.HandleFunc("/signup", signup)
}

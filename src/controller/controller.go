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

func ParseUser(r *http.Request) (bool, string, string) {
	var (
		isLogged bool
		username string
	)
	cookie, err := r.Cookie(COOKIE_NAME)
	if err != http.ErrNoCookie {
		sessionKey := cookie.Value
		username, isLogged = auth.Sessions.AuthenticateSessionKey(sessionKey)
		return isLogged, username, sessionKey
	}
	return isLogged, "", ""
}

func generatePage(username string, sessionKey string, w http.ResponseWriter) {
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
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		println("Request not a GET")
		return
	}

	isLogged, username, sessionKey := ParseUser(r)

	if isLogged {
		generatePage(username, sessionKey, w)
	} else {
		t, _ := template.ParseFiles(path.Join(templatesBase, "index.html"))
		t.Execute(w, nil)
	}
}

func signup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		t, _ := template.ParseFiles(path.Join(templatesBase, "signup.html"))
		t.Execute(w, nil)
	case "POST":
		r.ParseForm()

		username := template.HTMLEscapeString(r.Form.Get("username"))
		password := template.HTMLEscapeString(r.Form.Get("password"))

		if ok, err := auth.Users.AddUser(username, password); ok {
			t, _ := template.ParseFiles(path.Join(templatesBase, "index.html"))
			t.Execute(w, nil)
		} else {
			t, _ := template.ParseFiles(path.Join(templatesBase, "signup.html"))
			t.Execute(w, &struct{ WrongUsername string }{err})
		}
	default:
		println("Method not recognized")
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	isLogged, username, sessionKey := ParseUser(r)
	if isLogged {
		generatePage(username, sessionKey, w)
		return
	}
	if r.Method != "POST" {
		t, _ := template.ParseFiles(path.Join(templatesBase, "index.html"))
		t.Execute(w, nil)
		return
	}
	r.ParseForm()

	username = template.HTMLEscapeString(r.Form.Get("username"))
	password := template.HTMLEscapeString(r.Form.Get("password"))

	isLogged = auth.Users.AuthenticateLogin(username, password)

	if isLogged {
		sessionKey = auth.Sessions.NewSession(username) // just logged in because there was no valid cookie
		generatePage(username, sessionKey, w)
	} else {
		t, _ := template.ParseFiles(path.Join(templatesBase, "index.html"))
		t.Execute(w, &struct{ WrongUsername string }{"Wrong username or password."})
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	isLogged, _, sessionKey := ParseUser(r)
	if !isLogged || r.Method != "POST" {
		return
	}

	auth.Sessions.RemoveSession(sessionKey)

	t, _ := template.ParseFiles(path.Join(templatesBase, "index.html"))
	t.Execute(w, nil)
}

func Configure(templatesPath string) {
	templatesBase = templatesPath
	http.HandleFunc("/", index)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
}

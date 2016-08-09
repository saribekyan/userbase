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

func parseUser(r *http.Request) (bool, string, string) {
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
		Fullname  string
		NUsers    int
		NSessions int
	}{auth.Users.Fullname(username),
		auth.Users.NUsers(),
		auth.Sessions.NSessionsOfUser(username)})
}

func generateIndex(w http.ResponseWriter, wrongCredentials bool) {
	t, _ := template.ParseFiles(path.Join(templatesBase, "index.html"))
	t.Execute(w, &struct{ Wrong bool }{Wrong: wrongCredentials})
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		println("Request not a GET")
		return
	}

	isLogged, username, sessionKey := parseUser(r)

	if isLogged {
		generatePage(username, sessionKey, w)
	} else {
		generateIndex(w, false)
	}
}

func signup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		t, _ := template.ParseFiles(path.Join(templatesBase, "signup.html"))
		t.Execute(w, &struct{ Exists bool }{Exists: false})
	case "POST":
		r.ParseForm()

		fullname := template.HTMLEscapeString(r.Form.Get("fullname"))
		username := template.HTMLEscapeString(r.Form.Get("username"))
		password := template.HTMLEscapeString(r.Form.Get("password"))

		if ok, _ := auth.Users.AddUser(username, password, fullname); ok {
			generateIndex(w, false)
		} else {
			t, _ := template.ParseFiles(path.Join(templatesBase, "signup.html"))
			t.Execute(w, &struct{ Exists bool }{true})
		}
	default:
		println("Method not recognized")
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	isLogged, username, sessionKey := parseUser(r)
	if isLogged {
		generatePage(username, sessionKey, w)
		return
	}
	if r.Method != "POST" {
		generateIndex(w, false)
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
		generateIndex(w, true)
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	isLogged, _, sessionKey := parseUser(r)
	if isLogged && r.Method == "POST" {
		auth.Sessions.RemoveSession(sessionKey)
	}
	generateIndex(w, false)
}

func Configure(templatesPath string) {
	templatesBase = templatesPath
	http.HandleFunc("/", index)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)

	http.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("ui"))))
}

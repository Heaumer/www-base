package main

import (
	"flag"
	"github.com/gorilla/context"
	"github.com/gorilla/schema"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
	"sync"
)

var (
	decoder = schema.NewDecoder()
	sstore = sessions.NewCookieStore([]byte(securecookie.GenerateRandomKey(32)))

	registerb, _ = ioutil.ReadFile("templates/register.html")
	loginb, _ = ioutil.ReadFile("templates/login.html")
	footerb, _ = ioutil.ReadFile("templates/footer.html")

	headert = template.Must(
		template.New("header.html").ParseFiles("templates/header.html"))
	indext = template.Must(
		template.New("index.html").ParseFiles("templates/index.html"))
	settingst = template.Must(
		template.New("settings.html").ParseFiles("templates/settings.html"))
)

// connected users
var tokens struct {
	sync.Mutex
	v map[int64]*User
}

// execute a template, login the error if any
func writeTemplate(w http.ResponseWriter, t *template.Template, d interface{}) {
	if err := t.Execute(w, &d); err != nil {
		log.Println(err)
		w.Write([]byte("Internal error on template "+t.Name()))
	}
}

// create a new token
func mkToken() int64 {
	tokens.Lock()
start:
	token := rand.Int63()
	if _, exists := tokens.v[token]; exists {
		goto start
	}

	tokens.Unlock()

	return token
}

// set a new one-time-token for given user
func setToken(w http.ResponseWriter, r *http.Request, user *User) {
	session, _ := sstore.Get(r, "www-base-token")

	token := mkToken()
	session.Values["token"] = token
	tokens.v[token] = user
	session.Save(r, w)
}

// get user one-time-token
func getToken(r *http.Request) int64 {
	session, _ := sstore.Get(r, "www-base-token")

	if token, exists := session.Values["token"]; exists {
		return token.(int64)
	}
	return -1
}

// unset user token
func unsetToken(r *http.Request, token int64) {
	session, _ := sstore.Get(r, "www-base-token")
	delete(session.Values, "token")
	delete(tokens.v, token)
}

// pop error
func getError(w http.ResponseWriter, r *http.Request) string {
	session, _ := sstore.Get(r, "www-base-err")
	if err, exists := session.Values["error"]; exists {
		// XXX cf. flash messages in gorilla/sessions
		delete(session.Values, "error")
		session.Save(r, w)
		return err.(string)
	}
	return ""
}

// set a new error
func setError(w http.ResponseWriter, r *http.Request, err string) {
	session, _ := sstore.Get(r, "www-base-err")
	session.Values["error"] = err
	session.Save(r, w)
}

func index(w http.ResponseWriter, r *http.Request, u *User) {
	if u == nil {
		d := struct {
			Connected	bool
			User		*User
		} { false, &User{} }
		writeTemplate(w, indext, &d)
	} else {
		d := struct {
			Connected	bool
			User		*User
			HasWebsite	bool
			HasFullname	bool
		}{ true, u, u.Website != "", u.Fullname != "" }
		writeTemplate(w, indext, &d)
	}
}

func register(w http.ResponseWriter, r *http.Request, u *User) {
	// already connected?
	if u != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	switch r.Method {
	case "GET":
		w.Write(registerb)
	case "POST":
		if err := r.ParseForm(); err != nil {
			log.Println(err)
		}
		u := new(User)
		if err := decoder.Decode(u, r.PostForm); err != nil {
			log.Println(err)
		}
		if err := u.Register(); err != nil {
			log.Println(err)
			setError(w, r, err.Error())
			http.Redirect(w, r, "/register", http.StatusFound)
		} else {
			setToken(w, r, u)
			http.Redirect(w, r, "/", http.StatusFound)
		}
	}
}

func settings(w http.ResponseWriter, r *http.Request, u *User) {
	switch r.Method {
	case "GET":
		writeTemplate(w, settingst, u)
	case "POST":
		if err := r.ParseForm(); err != nil {
			log.Println(err)
		}
		// Ensure id is not set by user, as struct is automatically
		// filled from POST fields, retrieved from user.
		id := u.Id
		// Save old password. If the field has been left empty
		// by user, it won't be changed; else, if the password
		// is validated, it will be update.
		passwd := u.Passwd
		if err := decoder.Decode(u, r.PostForm); err != nil {
			log.Println(err)
		}
		u.Id = id
		if err := u.Update(passwd); err != nil {
			log.Println(err)
			setError(w, r, err.Error())
			http.Redirect(w, r, "/settings", http.StatusFound)
		} else {
			setToken(w, r, u)
			http.Redirect(w, r, "/", http.StatusFound)
		}
	}
}

func login(w http.ResponseWriter, r *http.Request, u *User) {
	// already connected?
	if u != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	switch r.Method {
	case "GET":
		w.Write(loginb)
	case "POST":
		if err := r.ParseForm(); err != nil {
			log.Println(err)
		}
		u := new(User)
		if err := decoder.Decode(u, r.PostForm); err != nil {
			log.Println(err)
		}
		if err := u.Login(); err != nil {
			log.Println("login failed")
			setError(w, r, err.Error())
			http.Redirect(w, r, "/login", http.StatusFound)
		} else {
			setToken(w, r, u)
			http.Redirect(w, r, "/", http.StatusFound)
		}
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	token := getToken(r)
	unsetToken(r, token)
	getError(w, r)	// pop error if any
	http.Redirect(w, r, "/", http.StatusFound)
}

// page for which user must already be authenticated
var mustauth = map[string]bool {
	"settings" : true,
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, *User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u *User
		token := getToken(r)
		if token > 0 {
			u = tokens.v[token]
			unsetToken(r, token)
			setToken(w, r, u)
		} else if mustauth[r.URL.Path[1:]] {
			setError(w, r, "Not yet connected")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		err := getError(w, r)

		if r.Method == "GET" {
			d := struct {
				Connected bool
				Title     string
				HasError  bool
				Error     string
			}{
				Connected: u != nil,
				Title:     "Sample website",
				HasError:  err != "",
				Error:     err,
			}
			writeTemplate(w, headert, &d)
		}

		fn(w, r, u)

		if r.Method == "GET" {
			w.Write(footerb)
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()
	tokens.v = make(map[int64]*User)

	store = NewSQLite("www-base.db")

	http.HandleFunc("/", makeHandler(index))
	http.HandleFunc("/register", makeHandler(register))
	http.HandleFunc("/settings", makeHandler(settings))
	http.HandleFunc("/login", makeHandler(login))
	http.HandleFunc("/logout", logout)

	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("./static/"))))

	log.Println("Launched on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080",
		context.ClearHandler(http.DefaultServeMux)))
}

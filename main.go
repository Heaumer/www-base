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
	"sync"
	"time"
)

var (
	decoder = schema.NewDecoder()
	sstore  = sessions.NewCookieStore([]byte(securecookie.GenerateRandomKey(32)))

	registerb, _ = ioutil.ReadFile("templates/register.html")
	loginb, _    = ioutil.ReadFile("templates/login.html")
	footerb, _   = ioutil.ReadFile("templates/footer.html")

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
		w.Write([]byte("Internal error on template " + t.Name()))
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
	session, _ := sstore.Get(r, "www-base")

	token := mkToken()
	session.Values["token"] = token
	tokens.v[token] = user
	session.Save(r, w)
}

// get user one-time-token
func getToken(r *http.Request) int64 {
	session, _ := sstore.Get(r, "www-base")

	if token, exists := session.Values["token"]; exists {
		return token.(int64)
	}
	return -1
}

// unset user token
func unsetToken(r *http.Request, token int64) {
	session, _ := sstore.Get(r, "www-base")
	delete(session.Values, "token")
	delete(tokens.v, token)
}

// get a named flash message. Returns an empty string if none.
func getFlash(w http.ResponseWriter, r *http.Request, name string) string {
	session, _ := sstore.Get(r, "www-base")
	if flashes := session.Flashes(name); len(flashes) > 0 {
		session.Save(r, w)
		return flashes[0].(string)
	}

	return ""
}

// set a named flash message
func setFlash(w http.ResponseWriter, r *http.Request, name, value string) {
	session, _ := sstore.Get(r, "www-base")
	session.AddFlash(value, name)
	session.Save(r, w)
}

func index(w http.ResponseWriter, r *http.Request, u *User) {
	if u == nil {
		d := struct {
			Connected bool
			User      *User
		}{false, &User{}}
		writeTemplate(w, indext, &d)
	} else {
		d := struct {
			Connected   bool
			User        *User
			HasWebsite  bool
			HasFullname bool
			Data        []Data
		}{true, u, u.Website != "", u.Fullname != "", u.GetData()}
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
			setFlash(w, r, "error", err.Error())
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
		// XXX Save old password. If the field has been left empty
		// by user, it won't be changed; else, if the password
		// is validated, it will be update.
		passwd := u.Passwd
		if err := decoder.Decode(u, r.PostForm); err != nil {
			log.Println(err)
		}
		u.Id = id
		if err := u.Update(passwd); err != nil {
			log.Println(err)
			setFlash(w, r, "error", err.Error())
			http.Redirect(w, r, "/settings", http.StatusFound)
		} else {
			setToken(w, r, u)
			setFlash(w, r, "info", "settings updated")
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
			setFlash(w, r, "error", err.Error())
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
	getFlash(w, r, "error") // pop error if any
	http.Redirect(w, r, "/", http.StatusFound)
}

func unregister(w http.ResponseWriter, r *http.Request, u *User) {
	u.Delete()
	setFlash(w, r, "info", "account deleted")
	logout(w, r)
}

func add(w http.ResponseWriter, r *http.Request, u *User) {
	if err := r.ParseForm(); err != nil {
		log.Println(err)
	}
	d := new(Data)

	if err := decoder.Decode(d, r.PostForm); err != nil {
		log.Println(err)
	}

	d.Uid = u.Id
	log.Println(d)
	if err := d.Add(); err != nil {
		setFlash(w, r, "error", err.Error())
	}

	setFlash(w, r, "info", "new element added")
	http.Redirect(w, r, "/", http.StatusFound)
}

func editdel(w http.ResponseWriter, r *http.Request, u *User) {
	if err := r.ParseForm(); err != nil {
		log.Println(err)
	}
	d := new(Data)

	if err := decoder.Decode(d, r.PostForm); err != nil {
		log.Println(err)
	}

	// within the store, we edit/delete on both id/uid
	// so this should be ok even if user change d.Id
	d.Uid = u.Id
	log.Println(d)

	switch r.FormValue("action") {
	case "edit":
		if err := d.Edit(); err != nil {
			setFlash(w, r, "error", err.Error())
		} else {
			setFlash(w, r, "info", "element edited")
		}
	case "delete":
		if err := d.Delete(); err != nil {
			setFlash(w, r, "error", err.Error())
		} else {
			setFlash(w, r, "info", "element deleted")
		}
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// page for which user must already be authenticated
var mustauth = map[string]bool{
	"settings":   true,
	"unregister": true,
	"add":        true,
	"editdel":    true,
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
			setFlash(w, r, "error", "Not yet connected")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		err  := getFlash(w, r, "error")
		info := getFlash(w, r, "info")

		if r.Method == "GET" {
			d := struct {
				Connected bool
				Title     string
				HasError  bool
				Error     string
				HasInfo   bool
				Info      string
			}{
				Connected: u != nil,
				Title:     "Sample website",
				HasError:  err != "",
				Error:     err,
				HasInfo:   info != "",
				Info:      info,
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
	http.HandleFunc("/unregister", makeHandler(unregister))
	http.HandleFunc("/add", makeHandler(add))
	http.HandleFunc("/editdel", makeHandler(editdel))

	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("./static/"))))

	log.Println("Launched on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080",
		context.ClearHandler(http.DefaultServeMux)))
}

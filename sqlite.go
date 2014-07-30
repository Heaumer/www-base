package main

import (
	//	"database/sql"
	"errors"
	"github.com/kuroneko/gosqlite3"
	"log"
	"runtime"
)

type SQLite struct {
	*sqlite3.Database
}

func NewSQLite(fn string) (db *SQLite) {
	tmp, err := sqlite3.Open(fn)
	if err != nil {
		log.Fatal(err)
	}

	db = &SQLite{tmp}
	db.CreateTables()

	return
}

func iserr(err error) bool {
	return err != nil && err != sqlite3.ROW
}

func (db *SQLite) Execute2(s string, v ...interface{}) (stmt *sqlite3.Statement, err error) {

	stmt, err = db.Prepare(s, v...)
	if err == nil {
		err = stmt.Step()
	}
	if iserr(err) {
		_, file, line, _ := runtime.Caller(1)
		log.Printf("Error with SQLite: %s, at %s:%d\n", err, file, line)
	}
	return
}

func (db *SQLite) CreateTables() {
	db.Execute2(`CREATE TABLE IF NOT EXISTS user (
		id			integer		PRIMARY KEY AUTOINCREMENT,
		nick		text		UNIQUE NOT NULL,
		passwd		text		NOT NULL,
		email		text		UNIQUE NOT NULL,
		type		integer		,
		website		text		,
		fullname	text)
	`)

	db.Execute2(`CREATE TABLE IF NOT EXISTS data (
		id			integer		PRIMARY KEY AUTOINCREMENT,
		uid			integer		NOT NULL,
		name		text		NOT NULL,
		content		text		NOT NULL,
		public		integer		NOT NULL,
		FOREIGN KEY(uid) REFERENCES user(id) ON DELETE CASCADE)
	`)
}

// return the user. login may either be the login of the user
// or its email.
func (db *SQLite) GetUser(nick, passwd string) (User, error) {
	stmt, err := db.Execute2(`
		SELECT id, nick, passwd, email, type, website, fullname
		FROM user
		WHERE	(nick = (?) OR email = (?))
		AND		passwd = (?)`, nick, nick, passwd)

	v := stmt.Row()

	if v == nil || v[0] == nil || iserr(err) {
		return User{}, errors.New("Wrong nick/email or password")
	}

	return User{
		v[0].(int64),  // id
		v[1].(string), // nick
		v[2].(string), // passwd
		v[3].(string), // email
		v[4].(int64),  // type
		v[5].(string), // website
		v[6].(string), // fullname
	}, nil
}

func (db *SQLite) AddUser(u *User) error {
	_, err := db.Execute2(`
		INSERT INTO user(nick, passwd, email, type, website, fullname)
		VALUES (?, ?, ?, ?, ?, ?)`,
		u.Nick, u.Passwd, u.Email, u.Type, u.Website, u.Fullname)

	if iserr(err) {
		return errors.New("Nickname or email already taken")
	}

	// XXX safe? (maybe lock/unlock)
	u.Id = db.LastInsertRowID()

	return nil
}

func (db *SQLite) UpdateUser(u *User) error {
	_, err := db.Execute2(`
		UPDATE user
		SET
			passwd = (?),
			email = (?),
			website = (?),
			fullname = (?)
		WHERE id = (?)`,
		u.Passwd, u.Email, u.Website, u.Fullname, u.Id)

	if iserr(err) {
		return errors.New("Email already taken")
	}

	return nil
}

func (db *SQLite) RemUser(u *User) error {
	_, err := db.Execute2(`
		DELETE FROM user
		WHERE id = (?)`, u.Id)

	if iserr(err) {
		return errors.New("fortune: It's the only avant-garde we got.")
	}

	return nil
}

// Get every data owned by user
func (db *SQLite) GetData(uid int64) []Data {
	data := make([]Data, 0)
	var err error
	var stmt *sqlite3.Statement

	// connected?
	if uid != 0 {
		stmt, err = db.Prepare(`SELECT id, uid, name, content, public
			FROM data
			WHERE public = 1
			OR uid = (?)`, uid)
	} else {
		stmt, err = db.Prepare(`SELECT id, uid, name, content, public
			FROM data
			WHERE public = 1`)
	}

	if err != nil {
		log.Println(err)
		return data
	}

	stmt.All(func(s *sqlite3.Statement, v ...interface{}) {
		d := Data{
			v[0].(int64),
			v[1].(int64),
			v[2].(string),
			v[3].(string),
			v[4].(int64) == 1,
		}
		data = append(data, d)
	})

	return data
}

func (db *SQLite) AddData(d *Data) error {
	public := 0
	if d.Public { public = 1 }

	_, err := db.Execute2(`
		INSERT INTO data(uid, name, content, public)
		VALUES(?, ?, ?, ?)`,
		d.Uid, d.Name, d.Content, public)

	if iserr(err) {
		return errors.New("A spark, somewhere deep in the machine.")
	}

	// XXX safe? (maybe lock/unlock)
	d.Id = db.LastInsertRowID()

	return nil
}

func (db *SQLite) UpdateData(d *Data) error {
	public := 0
	if d.Public { public = 1 }

	_, err := db.Execute2(`
		UPDATE data
		SET
			name = (?),
			content = (?)
			public = (?)
		WHERE id = (?)
		AND uid = (?)`,
		d.Name, d.Content, public, d.Id, d.Uid)

	if iserr(err) {
		return errors.New("Who let that ant get there?")
	}

	return nil
}

func (db *SQLite) RemData(d *Data) error {
	_, err := db.Execute2(`
		DELETE FROM data
		WHERE id = (?)
		AND uid = (?)`, d.Id, d.Uid)

	if iserr(err) {
		return errors.New("A mischevious being made a move.")
	}

	return nil
}

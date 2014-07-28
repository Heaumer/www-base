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

func (db *SQLite) Execute2(s string, v ...interface{}) (stmt *sqlite3.Statement, err error) {

	stmt, err = db.Prepare(s, v...)
	if err == nil {
		err = stmt.Step()
	}
	if err != nil && err != sqlite3.ROW {
		_, file, line, _ := runtime.Caller(1)
		log.Printf("Error with SQLite: %s, at %s:%d\n", err, file, line)
	}
	return
}

func (db *SQLite) CreateTables() {
	db.Execute2(`CREATE TABLE IF NOT EXISTS users (
		id			integer		PRIMARY KEY AUTOINCREMENT,
		nick		text		UNIQUE NOT NULL,
		passwd		text		NOT NULL,
		email		text		UNIQUE NOT NULL,
		type		integer		,
		website		text		,
		fullname	text)
	`)
}

// return the user. login may either be the login of the user
// or its email.
func (db *SQLite) GetUser(nick, passwd string) (User, error) {
	stmt, err := db.Execute2(`
		SELECT id, nick, passwd, email, type, website, fullname
		FROM users
		WHERE	(nick = (?) OR email = (?))
		AND		passwd = (?)`, nick, nick, passwd)

	v := stmt.Row()

	if v == nil || v[0] == nil || err != nil && err != sqlite3.ROW {
		return User{}, errors.New("Wrong nick/email or password")
	}

	return User{
		v[0].(int64),	// id
		v[1].(string),	// nick
		v[2].(string),	// passwd
		v[3].(string),	// email
		v[4].(int64),		// type
		v[5].(string),	// website
		v[6].(string),	// fullname
	}, nil
}

func (db *SQLite) AddUser(u *User) error {
	_, err := db.Execute2(`
		INSERT INTO users(nick, passwd, email, type, website, fullname)
		VALUES (?, ?, ?, ?, ?, ?)`,
		u.Nick, u.Passwd, u.Email, u.Type, u.Website, u.Fullname)

	if err != nil && err != sqlite3.ROW {
		return errors.New("Nickname or email already taken")
	}

	u.Id = db.LastInsertRowID()

	return nil
}

func (db *SQLite) RemUser(u *User) error {
	return errors.New("Not implemented")
}
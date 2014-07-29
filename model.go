package main

import (
	//	"log"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"strings"
)

const (
	Single = iota
	Assoc
	Admin
)

type User struct {
	Id     int64
	Nick   string
	Passwd string
	Email  string
	Type   int64

	// Optional
	Website  string
	Fullname string
}

type Data struct {
	Id      int64
	Uid     int64	// owner
	Name    string
	Content string
}

func (u *User) Validate() error {
	if u.Nick == "" {
		return errors.New("Empty login")
	}
	if len(u.Passwd) < 8 {
		return errors.New("Password should be at least 8 characters")
	}
	if !strings.Contains(u.Email, "@") {
		return errors.New("Wrong Email format")
	}
	if u.Type < Single || u.Type > Assoc {
		return errors.New("No.")
	}

	return nil
}

func hashPasswd(passwd string) string {
	h := sha512.New()
	h.Write([]byte(passwd))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (u *User) Register() error {
	if err := u.Validate(); err != nil {
		return err
	}
	u.Passwd = hashPasswd(u.Passwd)
	return store.AddUser(u)
}

func (u *User) Update(oldpasswd string) error {
	if err := u.Validate(); err != nil {
		return err
	}

	// password has changed, and is valid
	if u.Passwd != oldpasswd {
		u.Passwd = hashPasswd(u.Passwd)
	}
	return store.UpdateUser(u)
}

func (u *User) Login() (err error) {
	*u, err = store.GetUser(u.Nick, hashPasswd(u.Passwd))
	return
}

func (u *User) String() string {
	web, fn := "", ""

	if u.Website != "" {
		web = "(" + u.Website + ")"
	}
	if u.Fullname != "" {
		fn = u.Fullname + "/"
	}

	return fn + u.Nick + " " + web
}

func (u *User) GetData() []Data {
	return store.GetData(u)
}

func (d *Data) Validate() error {
	if d.Name == "" {
		return errors.New("Empty name")
	}

	if d.Content == "" {
		return errors.New("No content")
	}

	return nil
}

func (d *Data) Add() error {
	if err := d.Validate(); err != nil {
		return err
	}

	return store.AddData(d)
}

func (d *Data) Delete() error {
	return errors.New("Not implemented")
}

func (d *Data) Edit() error {
	if err := d.Validate(); err != nil {
		return err
	}

	return store.UpdateData(d)
}

func (d *Data) String() string {
	return d.Name +": '"+d.Content+"'"
}
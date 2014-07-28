package main

type Store interface {
	GetUser(login, passwd string) (User, error)
	AddUser(u *User) error
	RemUser(u *User) error
}

var store Store

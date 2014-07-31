package main

type Store interface {
	// User
	GetUser(login, passwd string) (User, error)
	AddUser(*User) error
	UpdateUser(*User) error
	RemUser(*User) error

	// Data
	GetData(uid int64) []Data
	AddData(*Data) error
	UpdateData(*Data) error
	RemData(*Data) error

	// Permission
	Owns(uid, id int64) bool
}

var store Store

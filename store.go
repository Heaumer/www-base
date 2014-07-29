package main

type Store interface {
	// User related
	GetUser(login, passwd string) (User, error)
	AddUser(*User) error
	UpdateUser(*User) error
	RemUser(*User) error

	// Data related
	GetData(*User) []Data
	AddData(*Data) error
	UpdateData(*Data) error
	RemData(*Data) error
}

var store Store

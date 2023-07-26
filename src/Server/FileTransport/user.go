package fileTrans

import (
	"net"
)

type TconnType struct {
	Tconn    net.Conn
	Filename string
}

type UserModel struct {
	Name       string
	Token      string
	Conn       net.Conn
	TransConns map[string]net.Conn // 文件名-Conn
	IP         string
	Server     *FileServer
	ChTconn    chan TconnType
}

func NewUserModel(username string) UserModel {
	um := UserModel{
		Name:       username,
		TransConns: make(map[string]net.Conn),
		ChTconn:    make(chan TconnType, 100),
	}
	return um
}

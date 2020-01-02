package core

import "time"

type Event struct {
	Id     string `gorm:"primary_key"`
	Expire time.Time
	Detail EventDetail
}
type EventDetail struct {
	ServerInfo *ServerInfo
	Message    *Message
}
type ServerInfo struct {
	Name      string
	Addresses []string // stream address
}
type Message struct {
	Title string
	Body  string
}

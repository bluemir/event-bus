package core

import (
	"time"

	"github.com/sirupsen/logrus"
)

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

func (core *Core) gcEvent() error {
	result := core.db.Where("expire < ?", time.Now()).Delete(&Event{})

	if err := result.Error; err != nil {
		logrus.Warn("fail to gc Event", err)
	}
	logrus.Debugf("event gc done, delete %d", result.RowsAffected)
	return nil
}

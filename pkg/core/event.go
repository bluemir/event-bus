package core

import (
	"context"
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

func (core *Core) broadcast(evt *Event) error {
	// mark
	if err := core.db.Save(evt).Error; err != nil {
		return err
	}
	for _, a := range core.peers {
		// TODO ErrHandler. collect error...
		if err := a.encoder.Encode(evt); err != nil {
			logrus.Trace(err)
		}
	}
	return nil
}

func (core *Core) gcEvent(ctx context.Context) error {
	result := core.db.Where("expire < ?", time.Now()).Delete(&Event{})

	if err := result.Error; err != nil {
		logrus.Warn("fail to gc Event", err)
	}
	logrus.Debugf("event gc done, delete %d", result.RowsAffected)
	return nil
}

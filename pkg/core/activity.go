package core

import (
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

type ActivityKind int

const (
	Connected ActivityKind = iota + 1
	PacketRecived
)

type Activity struct {
	Kind ActivityKind
	URL  string
	At   time.Time
}

func (core *Core) collectActivity(addr net.Addr, kind ActivityKind) {
	a, ok := addr.(*websocket.Addr)
	if !ok {
		logrus.Error("not url")
		return
	}
	a.RawQuery = "" // remove query...

	if err := core.db.Save(&Activity{
		Kind: kind,
		URL:  addr.String(),
		At:   time.Now(),
	}).Error; err != nil {
		logrus.Warn(err)
		return
	}

	logrus.Tracef("activity: %s", a.String())
}
func (core *Core) gcActivity() error {
	result := core.db.Where("at < ?", time.Now().Add(-1*time.Hour)).Delete(&Activity{})

	if err := result.Error; err != nil {
		logrus.Warn("fail to gc activity", err)
	}
	logrus.Debugf("activity gc done, delete %d", result.RowsAffected)
	return nil
}

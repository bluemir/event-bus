package core

import (
	"net"
	"time"

	"github.com/sirupsen/logrus"
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
	//a, ok := addr.(*websocket.Addr)

	if err := core.db.Save(&Activity{
		Kind: kind,
		URL:  addr.String(),
		At:   time.Now(),
	}).Error; err != nil {
		logrus.Warn(err)
		return
	}
}

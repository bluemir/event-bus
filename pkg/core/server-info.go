package core

import (
	"context"
	"time"

	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
)

func (core *Core) broadcastServerInfo(ctx context.Context) error {
	if err := core.broadcast(&Event{
		Id:     xid.New().String(),
		Expire: time.Now().Add(60 * time.Second),
		Detail: EventDetail{
			ServerInfo: core.buildServerInfo(),
		},
	}); err != nil {
		logrus.Error(err)
		// ignore it because broadcast server info is minor function
	}
	return nil
}
func (core *Core) buildServerInfo() *ServerInfo {
	serverInfo := &ServerInfo{Name: core.serverName}
	result := []struct {
		Count int64
		URL   string
	}{}
	if err := core.db.Raw("select url, count(*) as count from activities group by url").Scan(&result).Error; err != nil {
		logrus.Error(err)
	}

	// TODO make weight info
	for _, c := range result {
		serverInfo.Addresses = append(serverInfo.Addresses, c.URL)
	}

	logrus.Debugf("serverInfo: %#v", serverInfo)

	return serverInfo
}

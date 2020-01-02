package core

import (
	"time"

	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
)

func (core *Core) broadcastServerInfo() error {
	core.broadcast(&Event{
		Id:     xid.New().String(),
		Expire: time.Now().Add(60 * time.Second),
		Detail: EventDetail{
			ServerInfo: core.buildServerInfo(),
		},
	})
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

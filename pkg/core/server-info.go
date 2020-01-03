package core

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
)

type ServerInfo struct {
	Name      string
	Addresses []string // stream address
	// Labels map[string]string
}

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
func (core *Core) updatePeerServerInfo(info *ServerInfo) {
	logrus.Infof("%#v", info)

	if err := core.db.Save(&PeerInfo{
		Id:            info.Name,
		ServerInfo:    info,
		LastHeartBeat: time.Now(),
	}).Error; err != nil {
		logrus.Error(err)
	}
}
func (core *Core) DebugPeerInfo() ([]PeerInfo, error) {
	result := []PeerInfo{}
	if err := core.db.Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

type PeerInfo struct {
	Id string `gorm:"primary_key"`
	*ServerInfo
	LastHeartBeat time.Time
	Score         int
}

func (info *ServerInfo) Value() (driver.Value, error) {
	return json.Marshal(info)
}
func (info *ServerInfo) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, info)
	case string:
		return json.Unmarshal([]byte(v), info)
	default:
		return errors.Errorf("not []byte or string")
	}
}

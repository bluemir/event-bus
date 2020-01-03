package core

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (core *Core) updateNodeInfo(info *ServerInfo) {
	logrus.Infof("%#v", info)

	if err := core.db.Save(&NodeInfo{
		Id:            info.Name,
		ServerInfo:    info,
		LastHeartBeat: time.Now(),
	}).Error; err != nil {
		logrus.Error(err)
	}
}
func (core *Core) DebugNodeInfo() ([]NodeInfo, error) {
	result := []NodeInfo{}
	if err := core.db.Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

type NodeInfo struct {
	Id string `gorm:"primary_key" json:"-"` // for gorm
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

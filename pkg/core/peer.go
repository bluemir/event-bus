package core

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

type Peer struct {
	encoder *json.Encoder
}

func (core *Core) updatePeerInfo(info *ServerInfo) {
	// TODO Implement
	logrus.Infof("%#v", info)
}

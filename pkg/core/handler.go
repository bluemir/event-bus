package core

import (
	"encoding/json"
	"time"

	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

func (core *Core) HandleConnection(conn *websocket.Conn) {
	defer conn.Close()

	if conn.IsServerConn() {
		// it is child...
		logrus.
			WithField("request.remoteAddr", conn.Request().RemoteAddr).
			WithField("remoteAddr", conn.RemoteAddr()). // origin...
			WithField("localAddr", conn.LocalAddr()).
			Tracef("client accept")

		// XXX save origin for address?

	}

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	id := xid.New().String()

	core.peers[id] = &Peer{encoder}
	defer func() {
		delete(core.peers, id)
	}()

	evt := &Event{}
	for {
		if err := decoder.Decode(evt); err != nil {
			logrus.Error(err)
			encoder.Encode(map[string]interface{}{"msg": err.Error(), "error": true}) // ignore error. already error occur
			return
		}

		logrus.Debug(evt)
		// if evt.Expire < time.Now()
		if evt.Expire.Before(time.Now()) {
			logrus.WithField("eid", evt.Id).Tracef("ignore event. because expired")
			continue // ignore
		}

		if !core.db.Where(&Event{Id: evt.Id}).Take(&Event{}).RecordNotFound() {
			logrus.WithField("eid", evt.Id).Tracef("ignore event. because already received")
			continue
		}

		if err := core.db.Save(evt).Error; err != nil {
			logrus.Error(err) // error on save event
			continue
		}

		// fire
		core.broadcast(evt)
	}
}

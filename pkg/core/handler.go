package core

import (
	"encoding/json"
	"time"

	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

func getServerId(conn *websocket.Conn) string {
	if conn.IsServerConn() {
		id := conn.Request().Header.Get(HeaderServerId)
		if id != "" {
			return id
		}
	}
	return xid.New().String()
}

func (core *Core) HandleConnection(conn *websocket.Conn) {
	defer conn.Close()

	if conn.IsServerConn() {
		logrus.
			WithField("request.remoteAddr", conn.Request().RemoteAddr).
			WithField("localAddr", conn.LocalAddr()).
			Tracef("client accept")

		core.collectActivity(conn.LocalAddr(), Connected)
	}

	// auth

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	// maybe need lock?
	id := getServerId(conn)
	if exist := core.peerAdd(id, &Peer{
		encoder:      encoder,
		isClientConn: conn.IsClientConn(),
	}); exist {
		logrus.Warn("already connected")
		return
	}
	defer core.peerDel(id)

	evt := &Event{}
	for {

		if err := decoder.Decode(evt); err != nil {
			logrus.Error(err)
			encoder.Encode(map[string]interface{}{"msg": err.Error(), "error": true}) // ignore error. already error occur
			return
		}
		if conn.IsServerConn() {
			core.collectActivity(conn.LocalAddr(), PacketRecived)
		}

		// ===== check duplication

		// if evt.Expire < time.Now()
		if evt.Expire.Before(time.Now()) {
			logrus.WithField("eid", evt.Id).Tracef("ignore event. because expired")
			continue // ignore
		}

		if !core.db.Where(&Event{Id: evt.Id}).Take(&Event{}).RecordNotFound() {
			logrus.WithField("eid", evt.Id).Tracef("ignore event. because already received")
			continue
		}

		// ===== send to other peer
		if err := core.broadcast(evt); err != nil {
			logrus.Error(err) // error on broadcast
			continue
		}

		// ===== handle event
		if evt.Detail.ServerInfo != nil {
			// it is server event
			core.updateNodeInfo(evt.Detail.ServerInfo)
		}

		if evt.Detail.Message != nil {
			logrus.WithField("eid", evt.Id).WithField("at", evt.Expire).Trace(evt.Detail.Message)
		}
	}
}

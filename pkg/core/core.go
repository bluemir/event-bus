package core

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

type Config struct {
	NetworkId  string
	NetworkKey string
	Peers      []string
	Retry      int
}

func New(db *gorm.DB, conf *Config) (*Core, error) {
	return &Core{db, conf, map[string]*Peer{}}, nil
}

type Core struct {
	db     *gorm.DB
	config *Config
	peers  map[string]*Peer
}

func (core *Core) Run(ctx context.Context) error {
	for _, peer := range core.config.Peers {
		if err := core.tryConnect(ctx, peer); err != nil {
			return err
		}
	}
	// TODO wait?
	return nil
}
func (core *Core) tryConnect(ctx context.Context, peer string) error {
	logrus.Tracef("try to connect '%s'", peer)

	conf, err := websocket.NewConfig("ws://"+peer+"/v1/stream", peer)
	if err != nil {
		return err
	}
	conf.Header = map[string][]string{
		"Authorization": []string{
			"token " + core.getToken(),
		},
	}
	for retry := 0; retry < core.config.Retry; retry++ {
		conn, err := websocket.DialConfig(conf)
		if err != nil {
			logrus.Errorf("connection failed(retry: %d): %s", retry, err)
			time.Sleep(1*time.Second + time.Duration(retry*retry)*time.Second)
			continue
		}
		logrus.Trace("connected. reset retry")
		retry = 0
		core.HandleConnection(conn)
	}
	return errors.Errorf("connection failed")
}
func (core *Core) HandleConnection(conn *websocket.Conn) {
	defer conn.Close()

	if conn.IsServerConn() {
		// it is child...
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

		// fire
		core.broadcast(evt)
	}
}
func (core *Core) broadcast(evt *Event) {
	for _, a := range core.peers {
		// TODO ErrHandler
		a.encoder.Encode(evt)
	}
}

type Peer struct {
	encoder *json.Encoder
}

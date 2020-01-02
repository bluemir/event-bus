package core

import (
	"context"
	"encoding/json"
	"net"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
	"golang.org/x/sync/errgroup"
)

type Config struct {
	NetworkId  string
	NetworkKey string
	Peers      []string
	Retry      int
}

func New(db *gorm.DB, conf *Config) (*Core, error) {
	if err := db.AutoMigrate(&Event{}).Error; err != nil {
		return nil, err
	}
	return &Core{db, conf, map[string]*Peer{}}, nil
}

type Core struct {
	db     *gorm.DB
	config *Config
	peers  map[string]*Peer
}

func (core *Core) Run(ctx context.Context) error {
	eg, nCtx := errgroup.WithContext(ctx)

	for _, peer := range core.config.Peers {
		eg.Go(func() error {
			addr := peer // copy
			return core.tryConnect(ctx, addr)
		})
	}

	eg.Go(func() error {
		tick := time.NewTicker(1 * time.Minute)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				core.broadcast(&Event{
					Id:     xid.New().String(),
					Expire: time.Now().Add(60 * time.Second),
					Detail: EventDetail{
						ServerInfo: &ServerInfo{},
					},
				}) // TODO send server info
			case <-nCtx.Done():
				return nCtx.Err()
			}
		}
	})
	eg.Go(func() error {
		tick := time.NewTicker(5 * time.Minute)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				result := core.db.Where("expire < ?", time.Now()).Delete(&Event{})

				if err := result.Error; err != nil {
					logrus.Warn("fail to gc Event", err)
				}
				logrus.Debugf("gc done, delete %d", result.RowsAffected)
			case <-nCtx.Done():
				return nCtx.Err()
			}
		}
	})

	return eg.Wait()
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

func (core *Core) broadcast(evt *Event) {
	for _, a := range core.peers {
		// TODO ErrHandler
		if err := a.encoder.Encode(evt); err != nil {
			logrus.Trace(err)
		}
	}
}

type Peer struct {
	encoder *json.Encoder
}

func (core *Core) getAddrs() ([]string, error) {
	result := []string{}
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				result = append(result, v.IP.String())
			case *net.IPAddr:
				result = append(result, v.IP.String())
			}
		}
	}
	return result, nil
}

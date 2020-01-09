package core

import (
	"context"
	"math/rand"
	"net/url"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
	"golang.org/x/sync/errgroup"
)

const (
	HeaderServerId = "Server-Id"
)

type Config struct {
	NetworkId  string
	NetworkKey string
	Peers      []*url.URL
	Retry      int
	ConnNumber int
}

func New(db *gorm.DB, conf *Config) (*Core, error) {
	if err := db.AutoMigrate(
		&Event{},
		&Activity{},
		&NodeInfo{},
	).Error; err != nil {
		return nil, err
	}
	return &Core{
		db:         db,
		config:     conf,
		peers:      map[string]*Peer{},
		serverName: xid.New().String(),
	}, nil
}

type Core struct {
	db         *gorm.DB
	config     *Config
	peers      map[string]*Peer
	serverName string
}

func (core *Core) Run(ctx context.Context) error {
	eg, nCtx := errgroup.WithContext(ctx)

	for _, peer := range core.config.Peers {
		eg.Go(func() error {
			addr := peer // copy
			for {
				if err := core.tryConnect(ctx, addr.String()); err != nil {
					logrus.Errorf("connection fail in initial peer: %s", err)
					// or return err
				}
			}
			// return core.tryConnect(ctx, addr, core.delayDefault)
		})
	}

	eg.Go(cron(nCtx, core.broadcastServerInfo, 1*time.Minute))
	eg.Go(cron(nCtx, core.gcEvent, 30*time.Second))
	eg.Go(cron(nCtx, core.gcActivity, 1*time.Minute))
	eg.Go(cron(nCtx, core.makePeerConnection, 30*time.Second))
	//eg.Go(cron(nCtx, core.makeServerConnection, 30*time.Second))

	return eg.Wait()
}

func (core *Core) tryConnect(ctx context.Context, peerAddresses ...string) error {
	for try := 0; try < core.config.Retry; try++ {
		for _, peer := range peerAddresses {
			logrus.Tracef("try to connect '%s'", peer)

			conf, err := websocket.NewConfig(peer, peer)
			if err != nil {
				return err
			}
			conf.Header = map[string][]string{
				"Authorization": []string{
					"token " + core.getToken(),
				},
				HeaderServerId: []string{core.serverName},
			}
			//
			conn, err := websocket.DialConfig(conf)
			if err != nil {
				logrus.Debugf("connection failed: %s", err)
				continue
			}

			logrus.Trace("connected.")
			core.HandleConnection(conn)
			logrus.Info("Connection closed")
			return nil // connection closed.
		}

		// connection failed...
		delay := 1*time.Second + time.Duration(try*try)*time.Second
		if delay > 60*time.Second {
			// random duration 1min ~ 2min
			delay = 60*time.Second + time.Duration(rand.Intn(60))*time.Second
		}
		logrus.Errorf("connection failed(retry after:%s, count: %d)", delay, try)

		select {
		case <-time.After(delay):
			// keep goinig
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return errors.Errorf("connection failed.")
}

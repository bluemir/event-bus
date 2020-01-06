package core

import (
	"context"
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

			return core.tryConnect(ctx, addr, core.delayNoExit)
		})
	}

	eg.Go(cron(nCtx, core.broadcastServerInfo, 1*time.Minute))
	eg.Go(cron(nCtx, core.gcEvent, 30*time.Second))
	eg.Go(cron(nCtx, core.gcActivity, 1*time.Minute))

	return eg.Wait()
}

func (core *Core) tryConnect(ctx context.Context, peer *url.URL, retryDelayFunc func(int) time.Duration) error {
	logrus.Tracef("try to connect '%s'", peer)

	conf, err := websocket.NewConfig(peer.String(), peer.String())
	if err != nil {
		return err
	}
	conf.Header = map[string][]string{
		"Authorization": []string{
			"token " + core.getToken(),
		},
		HeaderServerId: []string{core.serverName},
	}

	for retry, delay := 0, retryDelayFunc(0); true; retry++ {
		conn, err := websocket.DialConfig(conf)
		if err != nil {
			delay = retryDelayFunc(retry)
			if delay < 0 {
				return errors.Errorf("connection failed. count %d", retry)
			}
			logrus.Errorf("connection failed(retry after:%s, count: %d): %s", delay, retry, err)
			time.Sleep(delay)
			continue
		}

		logrus.Trace("connected. reset retry")
		retry = 0

		core.HandleConnection(conn)
	}
	return errors.Errorf("connection failed")
}

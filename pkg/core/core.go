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
			return core.tryConnect(ctx, addr)
		})
	}

	eg.Go(cron(nCtx, core.broadcastServerInfo, 1*time.Minute))
	eg.Go(cron(nCtx, core.gcEvent, 30*time.Second))
	eg.Go(cron(nCtx, core.gcActivity, 1*time.Minute))

	return eg.Wait()
}

func (core *Core) tryConnect(ctx context.Context, peer *url.URL) error {
	logrus.Tracef("try to connect '%s'", peer)

	conf, err := websocket.NewConfig(peer.String(), peer.Host)
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

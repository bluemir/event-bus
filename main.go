package main

import (
	"context"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/bluemir/event-bus/pkg/core"
	"github.com/bluemir/event-bus/pkg/server"
	"github.com/bluemir/event-bus/pkg/util"
)

var Version string
var AppName string

type Config struct {
	Server server.Config
	Core   core.Config
}

func main() {
	LogLevel := 0
	conf := Config{}

	// setup flags
	app := kingpin.New(AppName, AppName+" description")
	app.Version(Version)

	app.Flag("verbose", "Log level").Short('v').CounterVar(&LogLevel)
	app.Flag("bind", "Bind address").Default(":8080").StringVar(&conf.Server.Bind)
	app.Flag("peer", "Peer address").StringsVar(&conf.Core.Peers)
	app.Flag("network", "Network ID").Default(xid.New().String()).PlaceHolder("RandomId").StringVar(&conf.Core.NetworkId)
	app.Flag("key", "Network secret key").Default(util.RandomString(32)).PlaceHolder("RandomString").StringVar(&conf.Core.NetworkKey)
	app.Flag("retry", "Retry count if connection closed").Default("10").IntVar(&conf.Core.Retry)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	// print version
	logrus.Info(Version)

	level := logrus.Level(LogLevel) + logrus.InfoLevel
	//logrus.SetOutput(os.Stderr)
	logrus.SetLevel(level) // Info level is default
	logrus.SetReportCaller(true)
	logrus.Infof("error level: %s", level)

	logrus.Debugf("%#v", conf)

	logrus.Warnf("Network ID:  '%s'", conf.Core.NetworkId)
	logrus.Warnf("Network Key: '%s'", conf.Core.NetworkKey)

	// Init DB
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		logrus.Error(errors.Wrap(err, "failed to connect database"))
	}
	db.DB().SetMaxOpenConns(1)
	c, err := core.New(db, &conf.Core)
	if err != nil {
		logrus.Error(err)
	}

	eg, ctx := errgroup.WithContext(context.Background())
	eg.Go(func() error {
		return server.Run(ctx, c, &conf.Server)
	})
	eg.Go(func() error {
		return c.Run(ctx)
	})
	if err := eg.Wait(); err != nil {
		logrus.Fatal(err)
	}

}

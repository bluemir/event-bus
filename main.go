package main

import (
	"context"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/bluemir/event-bus/pkg/auth"
	"github.com/bluemir/event-bus/pkg/server"
)

var Version string
var AppName string

type Config struct {
	DBPath string
	Server server.Config
	Auth   auth.Config
}

func main() {
	LogLevel := 0
	conf := Config{}

	// setup flags
	app := kingpin.New(AppName, AppName+" description")
	app.Version(Version)

	app.Flag("verbose", "Log level").Short('v').CounterVar(&LogLevel)
	app.Flag("bind", "Bind address").Default(":8080").StringVar(&conf.Server.Bind)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	// print version
	logrus.Info(Version)

	level := logrus.Level(LogLevel) + logrus.InfoLevel
	//logrus.SetOutput(os.Stderr)
	logrus.SetLevel(level) // Info level is default
	logrus.Infof("error level: %s", level)

	logrus.Debugf("%#v", conf)

	// Init DB
	db, err := gorm.Open("sqlite3", conf.DBPath)
	if err != nil {
		logrus.Error(errors.Wrap(err, "failed to connect database"))
	}

	authManager, err := auth.New(db, &conf.Auth)
	if err != nil {
		logrus.Panic(err)
	}

	eg, ctx := errgroup.WithContext(context.Background())
	eg.Go(func() error {
		return server.Run(ctx, authManager, &conf.Server)
	})
	if err := eg.Wait(); err != nil {
		logrus.Panic(err)
	}

}

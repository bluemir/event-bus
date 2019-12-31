package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/bluemir/event-bus/pkg/core"
	"github.com/bluemir/event-bus/pkg/dist"
)

type Config struct {
	Bind string
}

func Run(ctx context.Context, c *core.Core, conf *Config) error {
	server := &Server{conf, c}

	app := gin.New()

	// add template
	if html, err := NewRenderer(); err != nil {
		return err
	} else {
		app.SetHTMLTemplate(html)
	}

	// setup Logger
	writer := logrus.New().Writer()
	defer writer.Close()

	app.Use(gin.LoggerWithWriter(writer))
	app.Use(gin.Recovery())

	// handle static
	app.StaticFS("/static/", dist.Apps.HTTPBox())

	// Static pages
	//app.Use(server.Authn)
	app.GET("/", server.static("index.html"))

	// Core
	{
		v1 := app.Group("/v1", server.checkNetworkId)
		v1.GET("/ping")
		v1.GET("/stream", server.Stream)
	}

	// TODO graceful shutdown
	errc := make(chan error)
	go func() {
		errc <- app.Run(conf.Bind)
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		logrus.Tracef("context done")
		return ctx.Err()
	}
}

type Server struct {
	config *Config
	core   *core.Core
}

func (server *Server) static(path string) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, path, c)
	}
}

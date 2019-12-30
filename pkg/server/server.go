package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"

	"github.com/bluemir/event-bus/pkg/dist"
)

type Config struct {
	NetworkId string
	Bind      string
}

func Run(ctx context.Context, db *gorm.DB, conf *Config) error {
	server := &Server{db, map[string]*Agent{}}

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
		v1 := app.Group("/v1")
		v1.GET("/ping")
		v1.GET("/stream", server.Stream)
	}

	return app.Run(conf.Bind)
}

type Server struct {
	db     *gorm.DB
	agents map[string]*Agent
}

func (server *Server) static(path string) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, path, c)
	}
}

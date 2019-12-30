package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/bluemir/event-bus/pkg/auth"
	"github.com/bluemir/event-bus/pkg/dist"
)

type Config struct {
	Bind string
}

func Run(ctx context.Context, authManager *auth.Manager, conf *Config) error {
	server := &Server{authManager}

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

	app.GET("/user/register", server.static("register.html"))
	app.POST("/user/register", server.handleRegister)

	// Static pages
	app.Use(server.Authn)
	app.GET("/", server.static("index.html"))

	return app.Run(conf.Bind)
}

type Server struct {
	Auth *auth.Manager
}

func (server *Server) static(path string) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, path, c)
	}
}

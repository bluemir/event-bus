package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

func (server *Server) checkNetworkId(c *gin.Context) {
	h := c.GetHeader("Authorization")
	if t, ok := c.GetQuery("token"); h == "" && ok {
		h = "token " + t
	}

	if err := server.core.Auth(h); err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}
}

func (server *Server) Stream(c *gin.Context) {
	websocket.Handler(server.core.HandleConnection).ServeHTTP(c.Writer, c.Request)
}

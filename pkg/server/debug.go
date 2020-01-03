package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (server *Server) debugServerInfo(c *gin.Context) {
	info, err := server.core.DebugPeerInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}
	c.JSON(http.StatusOK, info)
}

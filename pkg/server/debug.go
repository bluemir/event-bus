package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (server *Server) debugNodeInfo(c *gin.Context) {
	info, err := server.core.DebugNodeInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}
	c.JSON(http.StatusOK, info)
}

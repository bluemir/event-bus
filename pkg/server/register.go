package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (server *Server) handleRegister(c *gin.Context) {
	req := &struct {
		Name     string `form:"name"`
		Password string `form:"password"`
		Confirm  string `form:"confirm"`
	}{}

	if err := c.ShouldBind(req); err != nil {
		c.HTML(http.StatusBadRequest, "errors/bad-request.html", gin.H{})
		c.Abort()
		return
	}
	if req.Password != req.Confirm {
		c.HTML(http.StatusBadRequest, "errors/bad-request.html", gin.H{
			"retryURL": "",
			"message":  "password and password confirm not matched",
		})
		c.Abort()
		return
	}

	if err := server.Auth.CreateUser(req.Name, map[string]string{}); err != nil {
		c.HTML(http.StatusInternalServerError, "errors/internal-server-error.html", gin.H{
			"retryURL": "",
			"message":  err.Error(),
		})
		c.Abort()
		return
	}

	if _, err := server.Auth.IssueToken(req.Name, req.Password); err != nil {
		c.HTML(http.StatusInternalServerError, "errors/internal-server-error.html", gin.H{
			"retryURL": "",
			"message":  err.Error(),
		})
		c.Abort()
		return
	}

	c.HTML(http.StatusOK, "welcome.html", gin.H{})
}

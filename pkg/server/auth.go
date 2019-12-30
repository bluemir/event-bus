package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/bluemir/event-bus/pkg/auth"
)

const (
	TOKEN = "token"
	Relem = "event-bus"
)

func (server *Server) Authn(c *gin.Context) {
	log := logrus.WithField("method", "BaiscAuthn")
	token, err := server.Auth.HTTP(c.Request.Header)

	if err != nil {
		switch err {
		case auth.ErrEmptyHeader:
			// check Authz with guest
			// if failed, must request login
			return
		default:
			c.Header(auth.HeaderWWWAuthenticate, auth.HttpRealm(Relem))
			c.HTML(http.StatusUnauthorized, "/errors/unauthorized.html", gin.H{})
			c.Abort()
			return
		}
	}
	log.Tracef("auth as '%s'", token.UserName)

	// check Authz with token

	c.Set(TOKEN, token)

	return // next

}
func (server *Server) Authz(action string) func(c *gin.Context) {
	log := logrus.WithField("method", "Authz")
	return func(c *gin.Context) {
		subject, err := server.Auth.Subject(Token(c))
		if err != nil {
			c.HTML(http.StatusInternalServerError, "/errors/internal-server-error.html", gin.H{})
			c.Abort()
			return
		}
		//object, err := server.Backend.Object(c.Request.RequestURI)
		object := map[string]string{}
		if err != nil {
			c.HTML(http.StatusInternalServerError, "/errors/internal-server-error.html", gin.H{})
			c.Abort()
			return
		}

		ctx := &auth.Context{
			Subject: subject,
			Object:  object,
			Action:  action,
		}

		log.Trace(ctx)

		switch server.Auth.Authz(ctx) {
		case auth.Accept:
			log.Trace("accepted")
			return
		case auth.Reject:
			log.Trace("rejected")
			c.HTML(http.StatusForbidden, "/errors/forbidden.html", gin.H{})
			c.Abort()
			return
		case auth.Error:
			c.HTML(http.StatusInternalServerError, "/errors/internal-server-error.html", gin.H{"err": err.Error()})
			c.Abort()
			return
		case auth.NeedAuthn:
			c.Header(auth.HeaderWWWAuthenticate, auth.HttpRealm(Relem))
			c.HTML(http.StatusUnauthorized, "/errors/unauthorized.html", gin.H{})
			c.Abort()
			return
		default:
			c.HTML(http.StatusInternalServerError, "/errors/internal-server-error.html", gin.H{})
			c.Abort()
			return
		}
	}
}

func Token(c *gin.Context) *auth.Token {
	token, ok := c.Get(TOKEN)
	if ok {
		return token.(*auth.Token)
	}
	return nil
}

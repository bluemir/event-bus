package server

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"golang.org/x/net/websocket"
)

func (server *Server) Stream(c *gin.Context) {
	websocket.Handler(func(conn *websocket.Conn) {
		defer conn.Close()

		encoder := json.NewEncoder(conn)
		decoder := json.NewDecoder(conn)

		id := xid.New().String()

		server.agents[id] = &Agent{encoder}
		defer func() {
			delete(server.agents, id)
		}()

		evt := &Event{}
		for {
			if err := decoder.Decode(evt); err != nil {
				encoder.Encode(gin.H{"msg": err.Error(), "error": true})
				return
			}

			// fire
			server.broadcast(evt)
		}
	}).ServeHTTP(c.Writer, c.Request)
}

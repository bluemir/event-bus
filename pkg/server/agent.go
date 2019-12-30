package server

import "encoding/json"

type Agent struct {
	encoder *json.Encoder
}

func (server *Server) broadcast(evt *Event) {
	for _, a := range server.agents {
		// TODO ErrHandler
		a.encoder.Encode(evt)
	}
}

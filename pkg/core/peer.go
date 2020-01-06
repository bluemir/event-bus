package core

import (
	"encoding/json"
)

type Peer struct {
	encoder *json.Encoder
}

func (core *Core) countPeer() int {
	return len(core.peers)
}
func (core *Core) peerAdd(id string, peer *Peer) bool {
	if _, ok := core.peers[id]; ok {
		return ok
	}
	core.peers[id] = peer
	return false
}
func (core *Core) peerDel(id string) {
	delete(core.peers, id)
}

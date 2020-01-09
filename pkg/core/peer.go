package core

import (
	"encoding/json"
)

type Peer struct {
	encoder      *json.Encoder
	isClientConn bool
}

func (core *Core) getPeer(id string) (*Peer, bool) {
	core.peersLock.RLock()
	defer core.peersLock.RUnlock()
	p, ok := core.peers[id]
	return p, ok
}

func (core *Core) countPeer() int {
	core.peersLock.RLock()
	defer core.peersLock.RUnlock()
	return len(core.peers)
}
func (core *Core) countClientConn() int {
	core.peersLock.RLock()
	defer core.peersLock.RUnlock()
	count := 0
	for _, p := range core.peers {
		if p.isClientConn {
			count++
		}
	}
	return count
}
func (core *Core) peerAdd(id string, peer *Peer) bool {
	core.peersLock.Lock()
	defer core.peersLock.Unlock()
	if _, ok := core.peers[id]; ok {
		return ok
	}
	core.peers[id] = peer
	return false
}
func (core *Core) peerDel(id string) {
	core.peersLock.Lock()
	defer core.peersLock.Unlock()
	delete(core.peers, id)
}

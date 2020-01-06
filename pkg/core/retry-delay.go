package core

import (
	"math/rand"
	"time"
)

func (core *Core) delayDefault(retry int) time.Duration {
	if retry > core.config.Retry {
		return -1
	}
	d := 1*time.Second + time.Duration(retry*retry)*time.Second
	if d > 60*time.Second {
		// random duration 1min ~ 2min
		return 60*time.Second + time.Duration(rand.Intn(60))*time.Second
	}
	return d
}

func (core *Core) delayNoExit(retry int) time.Duration {
	d := 1*time.Second + time.Duration(retry*retry)*time.Second
	if d > 60*time.Second {
		// random duration 1min ~ 2min
		return 60*time.Second + time.Duration(rand.Intn(60))*time.Second
	}
	return d
}

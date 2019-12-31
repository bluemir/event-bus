package core

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

func (core *Core) Auth(h string) error {
	arr := strings.SplitN(h, " ", 2)
	if len(arr) != 2 {
		return errors.Errorf("string no have ' '")
	}
	authType, value := arr[0], arr[1]

	switch authType {
	case "basic", "Basic", "token", "Token":
		if string(value) != core.getToken() {
			return errors.Errorf("not collected auth")
		}
	}
	return nil
}

func (core *Core) getToken() string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", core.config.NetworkId, core.config.NetworkKey)))
}

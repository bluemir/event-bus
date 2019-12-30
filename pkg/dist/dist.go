package dist

import (
	rice "github.com/GeertJohan/go.rice"
)

var (
	Apps      = rice.MustFindBox("../../build/dist")
	Templates = rice.MustFindBox("../../app/html")
)

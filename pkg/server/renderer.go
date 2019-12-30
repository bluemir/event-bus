package server

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/bluemir/event-bus/pkg/dist"
)

func NewRenderer() (*template.Template, error) {
	log := logrus.WithField("method", "NewRenderer")
	tmpl := template.New("__root__")

	dist.Templates.Walk("/", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && info.Name()[0] == '.' && path != "/" {
			return filepath.SkipDir
		}
		if info.IsDir() || info.Name()[0] == '.' || !strings.HasSuffix(path, ".html") {
			return nil
		}
		log.Debugf("parse template: path: %s", path)

		tmpl, err = tmpl.Parse(dist.Templates.MustString(path))
		if err != nil {
			return err
		}
		return nil
	})

	return tmpl, nil
}

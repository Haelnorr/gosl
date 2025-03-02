package embedfs

import (
	"embed"
	"io/fs"

	"github.com/pkg/errors"
)

//go:embed files/*
var embeddedFiles embed.FS

// Gets the embedded files
func GetEmbeddedFS() (fs.FS, error) {
	subFS, err := fs.Sub(embeddedFiles, "files")
	if err != nil {
		return nil, errors.Wrap(err, "fs.Sub")
	}
	return subFS, nil
}

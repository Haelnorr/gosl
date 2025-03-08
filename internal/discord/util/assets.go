package util

import (
	"bytes"
	"io/fs"

	"github.com/bwmarrin/discordgo"
)

// Get the asset with the provided name from the filesystem
func GetAsset(name string, files *fs.FS) (*discordgo.File, error) {
	fileData, err := fs.ReadFile(*files, "assets/"+name)
	if err != nil {
		return nil, err
	}
	file := &discordgo.File{
		Name:   "error.png",
		Reader: bytes.NewReader(fileData),
	}
	return file, nil
}

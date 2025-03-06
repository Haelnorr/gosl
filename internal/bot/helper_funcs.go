package bot

import (
	"bytes"
	"io/fs"

	"github.com/bwmarrin/discordgo"
)

func getAsset(name string, files *fs.FS) (*discordgo.File, error) {
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

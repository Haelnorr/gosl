package commands

import (
	"io/fs"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

func Test(logger *zerolog.Logger, files *fs.FS) Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		errorMsg := `
The interaction you attempted has failed. Please try again.
Testing newlines
Testing newlines 2
`
		errorResponse("An error occured", &errorMsg, files, s, i)
	}
}

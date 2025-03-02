package bot

import (
	"github.com/bwmarrin/discordgo"
)

func cmdTest(b *Bot) *command {
	return buildCommand("test", "Test", handleTest(b))
}

func handleTest(b *Bot) handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		errorMsg := `
The interaction you attempted has failed. Please try again.
Testing newlines
Testing newlines 2
`
		errorResponse("An error occured", &errorMsg, b.files, s, i)
	}
}

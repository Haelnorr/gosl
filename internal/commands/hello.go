package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

func Hi(logger *zerolog.Logger) Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		user := i.Member.User
		if user == nil {
			user = i.User
		}
		username := user.Username
		message := "Hello " + username + "!"
		if user.ID == "467654673084448788" {
			message = "Brother ewwwwwwww"
		}
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: message,
			},
		})
		if err != nil {
			s.ChannelMessageSend(i.ChannelID, "Failed to respond to /hi command")
			logger.Error().Err(err).Msg("Failed to respond to /hi command")
		}
	}
}

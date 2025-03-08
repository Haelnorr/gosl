package messages

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Reply to an interaction with a simple message
func Reply(
	msg string,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
		},
	})
	if err != nil {
		return errors.Wrap(err, "s.InteractionRespond")
	}
	return nil
}

// Reply to an interaction with an ephemeral message that deletes after
// 10 seconds
func ReplyEphemeral(
	msg string,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	logger *zerolog.Logger,
) error {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return errors.Wrap(err, "s.InteractionRespond")
	}
	// Wait for 10 seconds before deleting
	go func() {
		time.Sleep(10 * time.Second) // Adjust timeout as needed
		err := s.InteractionResponseDelete(i.Interaction)
		if err != nil {
			logger.Warn().Err(err).Str("msg", msg).
				Msg("Failed to delete emphemeral message")
		}
	}()
	return nil
}

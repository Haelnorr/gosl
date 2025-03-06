package bot

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) replyStatic(
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
		b.logger.Error().Err(err).Str("msg", msg).Msg("Failed to respond")
		return err
	}
	return nil
}

// Responds to the interaction with an ephemeral message that deletes after
// 10 seconds
func (b *Bot) replyEphemeral(
	msg string,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		b.logger.Error().Err(err).Str("msg", msg).Msg("Failed to respond")
		return err
	}
	// Wait for 10 seconds before deleting
	go func() {
		time.Sleep(10 * time.Second) // Adjust timeout as needed
		err := s.InteractionResponseDelete(i.Interaction)
		if err != nil {
			b.logger.Warn().Err(err).Str("msg", msg).
				Msg("Failed to delete emphemeral message")
		}
	}()
	return nil
}

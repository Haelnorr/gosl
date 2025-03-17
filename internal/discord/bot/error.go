package bot

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// Send an ephemeral error message to the user with details of the error
func (b *Bot) Error(
	pMsg string,
	sMsg string,
	i *discordgo.InteractionCreate,
	ack bool,
) error {
	embed := &discordgo.MessageEmbed{
		Color: 0xff1919,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "Error",
			IconURL: "attachment://error.png",
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   pMsg,
				Value:  sMsg,
				Inline: false,
			},
		},
	}
	errIco, err := GetAsset("error.png", b.Files)
	if err != nil {
		return errors.Wrap(err, "getAsset")
	}
	deleteafter := 60 * time.Second
	deleteat := time.Now().Add(deleteafter)
	if ack {
		_, err = b.Session.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("*This message will delete %s*", DiscordUntil(&deleteat)),
			Embeds:  []*discordgo.MessageEmbed{embed},
			Files:   []*discordgo.File{errIco},
			Flags:   discordgo.MessageFlagsEphemeral,
		})
	} else {
		err = b.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("*This message will delete %s*", DiscordUntil(&deleteat)),
				Embeds:  []*discordgo.MessageEmbed{embed},
				Files:   []*discordgo.File{errIco},
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
	// Wait for for the delay before deleting
	go func() {
		time.Sleep(deleteafter)
		err := b.Session.InteractionResponseDelete(i.Interaction)
		if err != nil {
			b.Logger.Warn().Err(err).Msg("Failed to delete emphemeral message")
		}
	}()
	return err
}

// Helper function to send a Error response to the user advising them they
// do not have permission to perform the action
func (b *Bot) Forbidden(
	i *discordgo.InteractionCreate,
	ack bool,
) {
	msg := "You do not have permission for this action"
	err := b.Error("Forbidden", msg, i, ack)
	if err != nil {
		b.Logger.Warn().Err(err).Msg("Failed to reply to interaction")
	}
}

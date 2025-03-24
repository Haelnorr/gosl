package bot

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// Acknowledge the interaction and prepare for an ephemeral response when
// interaction handling is complete. call FollowUP() to follow up
func (b *Bot) Acknowledge(
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.Logger.Debug().Msg("Acknowledging interaction")
	err := b.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return errors.Wrap(err, "s.InteractionRespond")
	}
	if ack != nil {
		*ack = true
	}
	return nil
}

// Acknowledge the interaction silently for no response
func (b *Bot) SilentAcknowledge(
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.Logger.Debug().Msg("Acknowledging interaction")
	err := b.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		return errors.Wrap(err, "s.InteractionRespond")
	}
	if ack != nil {
		*ack = true
	}
	return nil
}

// Reply to the interaction with a follow up ephemeral response that deletes after 5 seconds
func (b *Bot) FollowUp(
	msg string,
	i *discordgo.InteractionCreate,
) error {
	b.Logger.Debug().Msg("Responding to interaction")
	_, err := b.Session.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: msg,
	})
	if err != nil {
		return errors.Wrap(err, "s.FollowupMessageCreate")
	}
	// Wait for 5 seconds before deleting
	go func() {
		time.Sleep(5 * time.Second)
		err := b.Session.InteractionResponseDelete(i.Interaction)
		if err != nil {
			b.Logger.Warn().Err(err).Msg("Failed to delete emphemeral message")
		}
	}()
	return nil
}

// Reply to the interaction with a follow up message
func (b *Bot) FollowUpComplex(
	contents *MessageContents,
	i *discordgo.InteractionCreate,
	deleteafter time.Duration,
) error {
	b.Logger.Debug().Msg("Responding to interaction")
	deleteat := time.Now().Add(deleteafter)
	_, err := b.Session.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content:    fmt.Sprintf("*This message will delete %s*", DiscordUntil(&deleteat)),
		Embeds:     []*discordgo.MessageEmbed{contents.Embed},
		Components: contents.Components,
	})
	if err != nil {
		return errors.Wrap(err, "s.FollowupMessageCreate")
	}
	// Wait for the delay before deleting
	go func() {
		time.Sleep(deleteafter)
		err := b.Session.InteractionResponseDelete(i.Interaction)
		if err != nil {
			b.Logger.Warn().Err(err).Msg("Failed to delete emphemeral message")
		}
	}()
	return nil
}

// Reply to an interaction with a modal
func (b *Bot) ReplyModal(
	title string,
	customID string,
	components []discordgo.MessageComponent,
	i *discordgo.InteractionCreate,
) error {
	err := b.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			Title:      title,
			CustomID:   customID,
			Components: components,
		},
	})
	if err != nil {
		return errors.Wrap(err, "s.InteractionRespond")
	}
	return nil
}

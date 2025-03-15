package bot

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// TODO: change all interactions to use acknowledge and follow up

// Acknowledge the interaction and prepare for an ephemeral response when
// interaction handling is complete. call FollowUP() to follow up
func (b *Bot) Acknowledge(
	i *discordgo.InteractionCreate,
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
	// TODO: test this actually works
	go func() {
		time.Sleep(5 * time.Second)
		err := b.Session.InteractionResponseDelete(i.Interaction)
		if err != nil {
			b.Logger.Warn().Err(err).Str("msg", msg).
				Msg("Failed to delete emphemeral message")
		}
	}()
	return nil
}

// Reply to the interaction with a follow up message
func (b *Bot) FollowUpComplex(
	embed *discordgo.MessageEmbed,
	components []discordgo.MessageComponent,
	i *discordgo.InteractionCreate,
) error {
	b.Logger.Debug().Msg("Responding to interaction")
	_, err := b.Session.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
	if err != nil {
		return errors.Wrap(err, "s.FollowupMessageCreate")
	}
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

// Reply to an interaction with an ephemeral message that deletes after 5 seconds
func (b *Bot) Reply(
	msg string,
	i *discordgo.InteractionCreate,
) error {
	err := b.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return errors.Wrap(err, "s.InteractionRespond")
	}
	// Wait for 5 seconds before deleting
	go func() {
		time.Sleep(5 * time.Second)
		err := b.Session.InteractionResponseDelete(i.Interaction)
		if err != nil {
			b.Logger.Warn().Err(err).Str("msg", msg).
				Msg("Failed to delete reply message")
		}
	}()
	return nil
}

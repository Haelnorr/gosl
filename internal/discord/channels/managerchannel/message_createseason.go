package managerchannel

import (
	"context"
	"gosl/internal/discord/channels/channels"
	"gosl/internal/discord/messages"
	"gosl/internal/discord/util"
	"gosl/internal/models"
	"gosl/pkg/db"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// Get the message contents for the create season component
func createSeasonComponents(
	ctx context.Context,
	b *util.Bot,
) (util.MessageContents, error) {
	components := []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					CustomID: "create_season_button",
					Label:    "Create Season",
				},
			},
		},
	}
	return func() (
		string,
		*discordgo.MessageEmbed,
		[]discordgo.MessageComponent,
	) {
		b.Logger.Debug().Msg("Retreiving create season components")
		return "",
			&discordgo.MessageEmbed{
				Title: "Create Season",
				Description: `
Create a new season.
Season ID and Name must be unique.`,
				Color: 0x00ff00, // Green color
			},
			components
	}, nil
}

func handleCreateSeasonButtonInteraction(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	components := []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.TextInput{
					CustomID:    "season_id",
					Label:       "Season ID",
					Style:       discordgo.TextInputShort,
					Placeholder: "Season ID...",
					Required:    true,
				},
			},
		},
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.TextInput{
					CustomID:    "season_name",
					Label:       "Season Name",
					Style:       discordgo.TextInputShort,
					Placeholder: "Season Name...",
					Required:    true,
				},
			},
		},
	}
	err := messages.ReplyModal("Create Season", "create_season_modal", components, s, i)
	if err != nil {
		return errors.Wrap(err, "messages.ReplyModal")
	}
	return nil
}

func handleCreateSeasonModalInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *util.Bot,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	seasonID := i.ModalSubmitData().Components[0].(*discordgo.ActionsRow).
		Components[0].(*discordgo.TextInput).Value
	seasonName := i.ModalSubmitData().Components[1].(*discordgo.ActionsRow).
		Components[0].(*discordgo.TextInput).Value

	season, err := models.CreateSeason(ctx, tx, seasonID, seasonName)
	if err != nil {
		if strings.Contains(err.Error(), "must be unique") {
			b.Error("Error creating season", err.Error(), s, i)
			return nil
		}
		return errors.Wrap(err, "models.CreateSeason")
	}
	msg := "New Season created: " + season.Name
	b.Log().UserEvent(i.Member, msg)
	messages.ReplyEphemeral(msg, s, i, b.Logger)

	channelID, err := channels.GetChannel(ctx, tx, channels.PurposeManager)
	if err != nil {
		msg = "Failed to get update manager channel after season creation"
		b.Logger.Warn().Err(err).
			Msg(msg)
		b.Log().Error(msg, err)
		return nil
	}

	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		b.Logger.Debug().Msg("Updating season select")
		err := messages.UpdateChannelMessage(
			ctx,
			b,
			selectSeasonComponents,
			messages.ManagerSelectSeason,
			channelID,
		)
		if err != nil {
			msg := "Failed to update select season message after interaction"
			b.Logger.Warn().Err(err).
				Msg(msg)
			b.Log().Error(msg, err)
		}
	}()
	return nil
}

package managerchannel

import (
	"context"
	"gosl/internal/discord/messages"
	"gosl/internal/discord/util"
	"gosl/pkg/db"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// Get the message contents for the select admin roles component
func selectSeasonComponents(
	ctx context.Context,
	b *util.Bot,
) (util.MessageContents, error) {
	b.Logger.Debug().Msg("Setting up select season components")
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := b.Conn.Begin(timeout)
	if err != nil {
		return nil, errors.Wrap(err, "b.Conn.Begin")
	}
	defer tx.Rollback()
	b.Logger.Debug().Msg("Getting default values for select season components")
	// TODO: get default values for select season components
	tx.Commit()

	// defaultValues := []discordgo.SelectMenuDefaultValue{
	// 	discordgo.SelectMenuDefaultValue{
	// 		ID:   role,
	// 		Type: discordgo.SelectMenuDefaultValueRole,
	// 	},
	// }
	options := []discordgo.SelectMenuOption{}
	options = append(options, discordgo.SelectMenuOption{
		Label: "Season 21",
		Value: "s21",
	})
	return func() (
		string,
		*discordgo.MessageEmbed,
		[]discordgo.MessageComponent,
	) {
		b.Logger.Debug().Msg("Retreiving select season components")
		return "",
			&discordgo.MessageEmbed{
				Title: "Current Season",
				Description: `
Select the season to be set as the active season.
Can be set as null.

**NOTE**
This will update all related messages to show data for the selected season.
(i.e. team rosters, fixtures).`,
				Color: 0x00ff00, // Green color
			},
			messages.StringSelect(
				"season_select",
				"Select active season",
				nil,
				options,
				0,
				1,
			)
	}, nil
}

func handleSelectSeasonInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *util.Bot,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	thisChannel := i.ChannelID
	season := i.MessageComponentData().Values[0]
	// TODO: set the active season in the database

	msg := "Active season set to: " + season
	b.Log().UserEvent(i.Member, msg)
	messages.ReplyEphemeral(msg, s, i, b.Logger)
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		b.Logger.Debug().Msg("Updating season select")
		err := messages.UpdateChannelMessage(
			ctx,
			b,
			selectSeasonComponents,
			messages.ManagerSelectSeason,
			thisChannel,
		)
		if err != nil {
			b.Logger.Warn().Err(err).
				Msg("Failed to update select log channel message after interaction")
		}
		// TODO: update any other messages that display data from the active season
	}()
	return nil
}

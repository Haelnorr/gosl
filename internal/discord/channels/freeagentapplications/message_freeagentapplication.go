package freeagentapplications

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/components"
	"gosl/internal/models"
	"gosl/pkg/db"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func NewFreeAgentApplicationMsg(ctx context.Context, b *bot.Bot) (*bot.DynamicMessage, error) {
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := b.Conn.RBegin(timeout, "NewFreeAgentApplicationMsg")
	if err != nil {
		return nil, errors.Wrap(err, "b.Conn.RBegin")
	}
	defer tx.Rollback()
	channelID, err := models.GetChannel(ctx, tx, models.ChannelFreeAgentApplications)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetChannel")
	}
	msg := bot.NewDynamicMessage("Free Agent Registration Application", channelID, b)
	return msg, nil
}

func FreeAgentApplicationContents(
	ctx context.Context,
	tx db.SafeTX,
	app *models.FreeAgentRegistration,
) (*bot.MessageContents, error) {
	statusMsg := "Pending"
	canPlace := false
	if app.Approved != nil {
		if *app.Approved == 1 && app.Placed == 0 {
			statusMsg = "Approved"
			canPlace = true
		} else if *app.Approved == 1 && app.Placed != 0 {
			statusMsg = "Placed"
		} else {
			statusMsg = "Rejected"
		}
	}
	leagues, err := models.GetLeagues(ctx, tx, app.SeasonID, false)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetLeagues")
	}
	leagueOpts := []discordgo.SelectMenuOption{}
	for _, league := range *leagues {
		leagueOpts = append(leagueOpts, discordgo.SelectMenuOption{
			Label: league.Division,
			Value: fmt.Sprintf("%v", league.ID),
		})
	}
	selectLeague := components.StringSelect(
		fmt.Sprintf("place_freeagent_league_select_%v", app.ID),
		"Select League Placement",
		leagueOpts,
		1,
		1,
		!canPlace,
	)
	embed := &discordgo.MessageEmbed{
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "Free Agent Application",
				Value: fmt.Sprintf(`
**%s has applied as a Free Agent for %s!**
__Preferred League:__ %s
__Status:__ %s
__Placed In:__ %s
`,
					app.PlayerName, app.SeasonName, app.PreferredLeague,
					statusMsg, app.PlacedLeagueName),
				Inline: false,
			},
		},
	}
	msgcomps := []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					CustomID: fmt.Sprintf("approve_freeagent_application_%v", app.ID),
					Label:    "Approve application",
					Style:    discordgo.SuccessButton,
					Disabled: app.Approved != nil,
				},
				&discordgo.Button{
					CustomID: fmt.Sprintf("reject_freeagent_application_%v", app.ID),
					Label:    "Reject the application",
					Style:    discordgo.DangerButton,
					Disabled: app.Placed != 0,
				},
			},
		},
	}
	msgcomps = append(msgcomps, selectLeague...)
	contents := &bot.MessageContents{
		Embed:      embed,
		Components: msgcomps,
	}
	return contents, nil
}

func updateAppMsg(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	app *models.FreeAgentRegistration,
	locked bool,
) error {
	appMsg, err := b.GetDynamicMessage("Free Agent application", i.Message.ID, i.ChannelID)
	if err != nil {
		return errors.Wrap(err, "b.GetDynamicMessage")
	}
	contents, err := FreeAgentApplicationContents(ctx, tx, app)
	if err != nil {
		return errors.Wrap(err, "FreeAgentApplicationContents")
	}
	if locked {
		err = appMsg.Expire(contents)
		if err != nil {
			return errors.Wrap(err, "appMsg.Expire")
		}
	} else {
		err = appMsg.Update(contents)
		if err != nil {
			return errors.Wrap(err, "appMsg.Update")
		}
	}
	return nil
}

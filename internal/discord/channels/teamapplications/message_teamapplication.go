package teamapplications

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/components"
	"gosl/internal/models"
	"gosl/pkg/db"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func NewTeamApplicationMsg(ctx context.Context, b *bot.Bot) (*bot.DynamicMessage, error) {
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := b.Conn.RBegin(timeout, "NewTeamApplicationMsg")
	if err != nil {
		return nil, errors.Wrap(err, "b.Conn.RBegin")
	}
	defer tx.Rollback()
	channelID, err := models.GetChannel(ctx, tx, models.ChannelTeamApplications)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetChannel")
	}
	msg := bot.NewDynamicMessage("Team Registration Application", channelID, b)
	return msg, nil
}

func TeamApplicationContents(
	ctx context.Context,
	tx db.SafeTX,
	app *models.TeamRegistration,
) (*bot.MessageContents, error) {
	team, err := models.GetTeamByID(ctx, tx, app.TeamID)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetTeamByID")
	}
	now := time.Now()
	currentPlayers, err := team.Players(ctx, tx, &now, &now)
	if err != nil {
		return nil, errors.Wrap(err, "team.Players")
	}
	playersmsg := "**Players:**"
	sort.SliceStable(*currentPlayers, func(i, j int) bool {
		if (*currentPlayers)[i].ID == team.ManagerID {
			return true
		}
		if (*currentPlayers)[j].ID == team.ManagerID {
			return false
		}
		return false
	})
	for _, player := range *currentPlayers {
		if player.ID == team.ManagerID {
			playersmsg = playersmsg + "\n%s (Manager)"
		} else {
			playersmsg = playersmsg + "\n%s"
		}
		playersmsg = fmt.Sprintf(playersmsg, player.Name)
	}
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
		fmt.Sprintf("place_team_league_select_%v", app.ID),
		"Select League Placement",
		leagueOpts,
		1,
		1,
		!canPlace,
	)
	embed := &discordgo.MessageEmbed{
		Color: team.Color,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    fmt.Sprintf("%s (%s)", team.Name, team.Abbreviation),
			IconURL: team.Logo,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "Team Application",
				Value: fmt.Sprintf(`
**%s has applied to join %s!**
__Preferred League:__ %s
__Status:__ %s
__Placed In:__ %s

%s
`,
					team.Name, app.SeasonName, app.PreferredLeague,
					statusMsg, app.PlacedLeagueName, playersmsg),
				Inline: false,
			},
		},
	}
	msgcomps := []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					CustomID: fmt.Sprintf("approve_team_application_%v", app.ID),
					Label:    "Approve application",
					Style:    discordgo.SuccessButton,
					Disabled: app.Approved != nil,
				},
				&discordgo.Button{
					CustomID: fmt.Sprintf("reject_team_application_%v", app.ID),
					Label:    "Reject the application",
					Style:    discordgo.DangerButton,
					Disabled: app.Placed != 0,
				},
				&discordgo.Button{
					CustomID: fmt.Sprintf("refresh_team_application_%v", app.ID),
					Label:    "Refresh",
					Disabled: app.Placed != 0,
				},
				// TODO: add deleting team for rule breaking submissions
				// &discordgo.Button{
				// 	CustomID: fmt.Sprintf("delete_team_%v", team.ID),
				// 	Label:    "Delete the Team",
				// 	Style:    discordgo.DangerButton,
				// },
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
	app *models.TeamRegistration,
	locked bool,
) error {
	appMsg, err := b.GetDynamicMessage("Team application", i.Message.ID, i.ChannelID)
	if err != nil {
		return errors.Wrap(err, "b.GetDynamicMessage")
	}
	contents, err := TeamApplicationContents(ctx, tx, app)
	if err != nil {
		return errors.Wrap(err, "TeamApplicationContents")
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

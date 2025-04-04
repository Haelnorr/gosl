package teamrosters

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

var teamRosters = &bot.Message{
	Label:       "Team Rosters Info",
	Purpose:     models.MsgTeamRosters,
	GetContents: teamRostersContents,
}

func teamRostersContents(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
) (*bot.MessageContents, error) {
	contents, err := getTeamRostersContents(ctx, tx)
	if err != nil {
		return nil, errors.Wrap(err, "getTeamRostersContents")
	}

	return contents, nil
}

func getTeamRostersContents(
	ctx context.Context,
	tx db.SafeTX,
) (*bot.MessageContents, error) {
	proteamsmsg := ""
	imteamsmsg := ""
	openteamsmsg := ""
	proFAsmsg := ""
	imFAsmsg := ""
	openFAsmsg := ""
	unplacedteamsmsg := ""
	unplacedFAsmsg := ""
	currentSeason, err := models.GetActiveSeason(ctx, tx)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetActiveSeason")
	}
	err = updateTeamListsMessages(ctx, tx, currentSeason, &proteamsmsg, &imteamsmsg, &openteamsmsg)
	if err != nil {
		return nil, errors.Wrap(err, "updateTeamListsMessages")
	}
	err = updateUnplacedTeamListMessage(ctx, tx, currentSeason, &unplacedteamsmsg)
	if err != nil {
		return nil, errors.Wrap(err, "updateUnplacedTeamListMessage")
	}
	err = updateFreeAgentListsMessages(ctx, tx, currentSeason, &proFAsmsg, &imFAsmsg, &openFAsmsg)
	if err != nil {
		return nil, errors.Wrap(err, "updateFreeAgentListsMessages")
	}
	err = updateUnplacedFreeAgentsListMessage(ctx, tx, currentSeason, &unplacedFAsmsg)
	if err != nil {
		return nil, errors.Wrap(err, "updateUnplacedFreeAgentsListMessage")
	}

	contents := &bot.MessageContents{
		Embed: &discordgo.MessageEmbed{
			Title: "Team/Free Agent Rosters!",
			Description: fmt.Sprintf(`
**Teams:**
%s%s%s%s

**Free Agents:**
%s%s%s%s
`, proteamsmsg, imteamsmsg, openteamsmsg, unplacedteamsmsg,
				proFAsmsg, imFAsmsg, openFAsmsg, unplacedFAsmsg),
		},
		Components: []discordgo.MessageComponent{
			&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.Button{
						CustomID: "refresh_team_rosters",
						Label:    "Refresh",
					},
				},
			},
		},
	}
	return contents, nil
}

func UpdateTeamRosters(ctx context.Context, tx db.SafeTX, b *bot.Bot) error {
	msg, err := b.GetMessage(models.ChannelTeamRosters, models.MsgTeamRosters)
	if err != nil {
		return errors.Wrap(err, "b.GetMessage")
	}
	msg.StartUpdate(true)
	b.Logger.Debug().Msg("Updating team rosters")
	errch := make(chan error)
	go msg.Update(ctx, tx, errch)
	hadErr := false
	for err := range errch {
		if err != nil {
			hadErr = true
			msg := "Failed to update team rosters"
			b.DoubleError(msg, err)
		}
	}
	if hadErr {
		return errors.New("Failed to update message(s) after interaction")
	}
	return nil
}

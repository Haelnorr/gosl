package teamrosters

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"
	"time"

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
	b *bot.Bot,
) (*bot.MessageContents, error) {
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	// HACK:
	// we use a WTX here to force it to block until commit if an update to
	// the rosters resulted in this function being called
	// this is because the message update uses the Message.GetContents field
	// and doesnt take in a tx as an input
	tx, err := b.Conn.Begin(timeout, "teamRostersContents()")
	if err != nil {
		return nil, errors.Wrap(err, "b.Conn.RBegin")
	}
	defer tx.Rollback()

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

func UpdateTeamRosters(ctx context.Context, b *bot.Bot) error {
	msg, err := b.GetMessage(models.ChannelTeamRosters, models.MsgTeamRosters)
	if err != nil {
		return errors.Wrap(err, "b.GetMessage")
	}
	msg.StartUpdate(true)
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		b.Logger.Debug().Msg("Updating team rosters")
		errch := make(chan error)
		go msg.Update(ctx, errch)
		for err := range errch {
			if err != nil {
				msg := "Failed to update team rosters"
				b.DoubleError(msg, err)
			}
		}
	}()
	return nil
}

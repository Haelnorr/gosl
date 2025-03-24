package teamrosters

import (
	"context"
	"fmt"
	"gosl/internal/models"
	"gosl/pkg/db"
	"time"

	"github.com/pkg/errors"
)

func updateTeamListsMessages(
	ctx context.Context,
	tx db.SafeTX,
	currentSeason *models.Season,
	proteamsmsg *string,
	imteamsmsg *string,
	openteamsmsg *string,
) error {
	if currentSeason == nil {
		return nil
	}
	if proteamsmsg == nil || imteamsmsg == nil || openteamsmsg == nil {
		return errors.New("One or more teams messages are nil pointers")
	}
	activeleagues, err := models.GetLeagues(ctx, tx, currentSeason.ID, true)
	if err != nil {
		return errors.Wrap(err, "models.GetLeagues")
	}
	for _, league := range *activeleagues {
		teams, err := league.GetTeams(ctx, tx)
		if err != nil {
			return errors.Wrap(err, "league.GetTeams")
		}
		if len(*teams) == 0 {
			continue
		}
		var msg *string
		switch league.Division {
		case "Pro":
			*proteamsmsg = "__Pro Teams:__"
			msg = proteamsmsg
		case "IM":
			*imteamsmsg = "\n__IM Teams:__"
			msg = imteamsmsg
		case "Open":
			*openteamsmsg = "\n__Open Teams:__"
			msg = openteamsmsg
		}
		for _, team := range *teams {
			now := time.Now()
			players, err := team.Players(ctx, tx, &now, &now)
			if err != nil {
				return errors.Wrap(err, "team.Players")
			}
			playerslist := "Players:"
			for i, player := range *players {
				playerslist = playerslist + player.Name
				if i < len(*players)-1 {
					playerslist = playerslist + ", "
				}
			}
			*msg = fmt.Sprintf(`
%s
%s (%s) - managed by %s
%s
`, *msg, team.Name, team.Abbreviation, team.ManagerName, playerslist)
		}
	}
	return nil
}

func updateUnplacedTeamListMessage(
	ctx context.Context,
	tx db.SafeTX,
	currentSeason *models.Season,
	unplacedTeamsMsg *string,
) error {
	if currentSeason == nil {
		return nil
	}
	teams, err := currentSeason.GetApprovedTeams(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "currentSeason.GetApprovedTeams")
	}
	for _, team := range *teams {
		now := time.Now()
		players, err := team.Players(ctx, tx, &now, &now)
		if err != nil {
			return errors.Wrap(err, "team.Players")
		}
		playerslist := "Players:"
		for i, player := range *players {
			playerslist = playerslist + player.Name
			if i < len(*players)-1 {
				playerslist = playerslist + ", "
			}
		}
		*unplacedTeamsMsg = fmt.Sprintf(`
__Approved Teams:__
%s (%s) - managed by %s
%s
`, team.Name, team.Abbreviation, team.ManagerName, playerslist)
	}
	return nil
}

package teamrosters

import (
	"context"
	"gosl/internal/models"
	"gosl/pkg/db"

	"github.com/pkg/errors"
)

func updateFreeAgentListsMessages(
	ctx context.Context,
	tx db.SafeTX,
	currentSeason *models.Season,
	proFAsmsg *string,
	imFAsmsg *string,
	openFAsmsg *string,
) error {
	if currentSeason == nil {
		return nil
	}
	if proFAsmsg == nil || imFAsmsg == nil || openFAsmsg == nil {
		return errors.New("One or more free agent messages are nil pointers")
	}
	activeleagues, err := models.GetLeagues(ctx, tx, currentSeason.ID, true)
	if err != nil {
		return errors.Wrap(err, "models.GetLeagues")
	}
	for _, league := range *activeleagues {
		FAs, err := league.GetFreeAgents(ctx, tx)
		if err != nil {
			return errors.Wrap(err, "league.GetFreeAgents")
		}
		if len(*FAs) == 0 {
			continue
		}
		var msg *string
		switch league.Division {
		case "Pro":
			*proFAsmsg = "__Pro Free Agents:__"
			msg = proFAsmsg
		case "IM":
			*imFAsmsg = "\n__IM Free Agents:__"
			msg = imFAsmsg
		case "Open":
			*openFAsmsg = "\n__Open Free Agents:__"
			msg = openFAsmsg
		}
		for _, FA := range *FAs {
			*msg = *msg + "\n - " + FA.Name
		}
	}
	return nil
}

func updateUnplacedFreeAgentsListMessage(
	ctx context.Context,
	tx db.SafeTX,
	currentSeason *models.Season,
	unplacedFAsMsg *string,
) error {
	if currentSeason == nil {
		return nil
	}
	FAs, err := currentSeason.GetApprovedFreeAgents(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "currentSeason.GetApprovedFreeAgents")
	}
	if len(*FAs) == 0 {
		return nil
	}
	*unplacedFAsMsg = "__Approved Free Agents:__"
	for _, FA := range *FAs {
		*unplacedFAsMsg = *unplacedFAsMsg + "\n - " + FA.Name
	}
	return nil
}

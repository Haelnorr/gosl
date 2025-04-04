package freeagentapplications

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/channels/teamrosters"
	"gosl/internal/models"
	"gosl/pkg/db"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handlePlaceFreeAgentLeagueSelect(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
	appIDstr string,
) error {
	b.Acknowledge(i, ack)
	appID, err := strconv.ParseUint(appIDstr, 10, 0)
	if err != nil {
		return errors.Wrap(err, "strconv.ParseUint")
	}
	app, err := models.GetFreeAgentRegistration(ctx, tx, uint32(appID))
	if err != nil {
		return errors.Wrap(err, "models.GetFreeAgentRegistration")
	}
	if app.Approved == nil || *app.Approved == 0 {
		return b.Error("Failed to place free agent", "Application is not approved", i, *ack)
	}
	leagueIDstr := i.MessageComponentData().Values[0]
	leagueID, err := strconv.ParseUint(leagueIDstr, 10, 0)
	if err != nil {
		return errors.Wrap(err, "strconv.ParseUint")
	}

	err = app.Place(ctx, tx, uint16(leagueID))
	if err != nil {
		if strings.Contains(err.Error(), "VE:") {
			return b.Error("Failed to place free agent",
				strings.TrimPrefix(err.Error(), "VE:"), i, *ack)
		}
		return errors.Wrap(err, "app.Place")
	}
	msg := fmt.Sprintf("%s has been placed in %s for %s",
		app.PlayerName, app.PlacedLeagueName, app.SeasonName)

	player, err := models.GetPlayerByID(ctx, tx, app.PlayerID)
	if err != nil {
		return errors.Wrap(err, "models.GetPlayerByID")
	}
	err = b.SendDirectMessage("Free Agent Application Approved", msg, player.DiscordID)
	if err != nil {
		return errors.Wrap(err, "b.SendDirectMessage")
	}
	err = teamrosters.UpdateTeamRosters(ctx, b)
	if err != nil {
		return errors.Wrap(err, "teamrosters.UpdateTeamRosters")
	}
	var rolePerm uint16
	switch app.PlacedLeagueName {
	case "Pro":
		rolePerm = models.PermProFreeAgent
	case "IM":
		rolePerm = models.PermIMFreeAgent
	case "Open":
		rolePerm = models.PermOpenFreeAgent
	}
	roleIDs, err := models.GetRoles(ctx, tx, rolePerm)
	if err != nil {
		return errors.Wrap(err, "models.GetRoles")
	}
	if len(roleIDs) == 0 {
		return errors.New("No roles for that purpose configured")
	}
	if len(roleIDs) > 1 {
		return errors.New("Multiple roles for that purpose configured, cannot add to user")
	}
	roleID := roleIDs[0]
	err = b.AddRoleToUser(player, roleID)
	if err != nil {
		return errors.Wrap(err, "b.AddRoleToUser")
	}

	err = updateAppMsg(ctx, tx, b, i, app, true)
	if err != nil {
		return errors.Wrap(err, "updateAppMsg")
	}
	b.Log().UserEvent(i.Member, msg)
	return b.FollowUp(msg, i)
}

package teamapplications

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/channels/teamrosters"
	"gosl/internal/models"
	"gosl/pkg/db"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handlePlaceTeamLeagueSelect(
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
	app, err := models.GetTeamRegistration(ctx, tx, uint16(appID))
	if err != nil {
		return errors.Wrap(err, "models.GetTeamRegistration")
	}
	if app.Approved == nil || *app.Approved == 0 {
		return b.Error("Failed to place team", "Application is not approved", i, *ack)
	}

	leagueIDstr := i.MessageComponentData().Values[0]
	leagueID, err := strconv.ParseUint(leagueIDstr, 10, 0)
	if err != nil {
		return errors.Wrap(err, "strconv.ParseUint")
	}

	err = app.Place(ctx, tx, uint16(leagueID))
	if err != nil {
		if strings.Contains(err.Error(), "VE:") {
			return b.Error("Failed to place team",
				strings.TrimPrefix(err.Error(), "VE:"), i, *ack)
		}
		return errors.Wrap(err, "app.Place")
	}

	msg := fmt.Sprintf("%s has been placed in %s for %s",
		app.TeamName, app.PlacedLeagueName, app.SeasonName)

	err = b.SendDirectMessage("Team Application Approved", msg, app.ManagerID)
	if err != nil {
		return errors.Wrap(err, "b.SendDirectMessage")
	}
	err = teamrosters.UpdateTeamRosters(ctx, tx, b)
	if err != nil {
		return errors.Wrap(err, "teamrosters.UpdateTeamRosters")
	}
	team, err := models.GetTeamByID(ctx, tx, app.TeamID)
	if err != nil {
		return errors.Wrap(err, "models.GetTeamByID")
	}

	err = applyTeamRoles(ctx, tx, b, team, app.PlacedLeagueName)
	if err != nil {
		return errors.Wrap(err, "applyTeamRoles")
	}

	err = updateAppMsg(ctx, tx, b, i, app, true)
	if err != nil {
		return errors.Wrap(err, "updateAppMsg")
	}
	b.Log().UserEvent(i.Member, msg)
	return b.FollowUp(msg, i)
}

func applyTeamRoles(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	team *models.Team,
	placedLeagueName string,
) error {
	roleExists := false
	roleID := ""
	var err error
	if team.RoleID != "" {
		roleExists, err = b.CheckRoleExists(team.RoleID)
		if err != nil {
			return errors.Wrap(err, "b.CheckRoleExists")
		}
		roleID = team.RoleID
	}
	if !roleExists {
		roleID, err = b.CreateRole(team.Name, &team.Color, true)
		if err != nil {
			return errors.Wrap(err, "b.CreateRole")
		}
		defer func() {
			if err != nil {
				err = b.DeleteRole(roleID)
				if err != nil {
					b.Logger.Error().Err(err).Msg("Failed to delete role after failure during role assignment")
				}
			}
		}()
		err = team.AddRole(ctx, tx, roleID)
		if err != nil {
			return errors.Wrap(err, "team.AddRole")
		}
		faRoles, err := models.GetRoles(ctx, tx, models.PermOpenFreeAgent)
		if err != nil {
			return errors.Wrap(err, "models.GetRoles")
		}
		if len(faRoles) == 0 {
			return errors.New("Failed creating team role, no Open FA configured")
		}
		err = b.FloatRoleUnder(roleID, faRoles[0])
		if err != nil {
			return errors.Wrap(err, "b.FloatRoleUnder")
		}
	}
	if roleID == "" {
		return errors.New("Failed to setup role for team")
	}
	var teamMgrRolePerm uint16
	switch placedLeagueName {
	case "Pro":
		teamMgrRolePerm = models.PermProTeamManager
	case "IM":
		teamMgrRolePerm = models.PermIMTeamManager
	case "Open":
		teamMgrRolePerm = models.PermOpenTeamManager
	}
	roleIDs, err := models.GetRoles(ctx, tx, teamMgrRolePerm)
	if err != nil {
		return errors.Wrap(err, "models.GetRoles")
	}
	if len(roleIDs) == 0 {
		return errors.New("No roles for that purpose configured")
	}
	if len(roleIDs) > 1 {
		return errors.New("Multiple roles for that purpose configured, cannot add to user")
	}
	managerRoleID := roleIDs[0]

	now := time.Now()
	players, err := team.Players(ctx, tx, &now, &now)
	if err != nil {
		return errors.Wrap(err, "team.Players")
	}
	for _, player := range *players {
		err = b.AddRoleToUser(&player, roleID)
		if err != nil {
			return errors.Wrap(err, "b.AddRoleToUser")
		}
		if player.ID == team.ManagerID {
			err = b.AddRoleToUser(&player, managerRoleID)
			if err != nil {
				return errors.Wrap(err, "b.AddRoleToUser")
			}
		}
	}

	return nil
}

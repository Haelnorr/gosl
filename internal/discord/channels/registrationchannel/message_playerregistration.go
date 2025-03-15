package registrationchannel

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/components"
	"gosl/internal/models"
	"gosl/pkg/db"
	"gosl/pkg/slapshotapi"
	"gosl/pkg/steamapi"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

var selectManagerRoles = &bot.Message{
	Label:       "Player Registration",
	Purpose:     models.MsgPlayerRegistration,
	GetContents: playerRegistrationContents,
}

// Get the message contents for the select manager roles component
func playerRegistrationContents(
	ctx context.Context,
	b *bot.Bot,
) (bot.MessageContents, error) {
	return func() (
		string,
		*discordgo.MessageEmbed,
		[]discordgo.MessageComponent,
	) {
		b.Logger.Debug().Msg("Retrieving player registration components")
		return "",
			&discordgo.MessageEmbed{
				Title: "Player Registration",
				Description: `
Register as a player in the Oceanic Slapshot League!
TODO: update here instructions
`,
				Color: 0x00ff00, // Green color
			},
			[]discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.Button{
							Label:    "Player Registration",
							CustomID: "player_registration_button",
						},
					},
				},
			}
	}, nil
}

// Handle an interaction with the select manager roles component
func handlePlayerRegistrationButtonInteraction(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
) error {
	player, err := models.GetPlayerByDiscordID(ctx, tx, i.Member.User.ID)
	if err != nil {
		return errors.Wrap(err, "models.GetPlayerByDiscordID")
	}
	if player != nil {
		msg := fmt.Sprintf("__Player Name:__ %s\n__Slap ID:__ %v", player.Name, player.SlapID)
		b.Error("You are already registered", msg, i)
		return nil
	}
	steamcmp := []discordgo.MessageComponent{
		components.TextInput("steam_id", "Steam ID", true, ""),
	}

	err = b.ReplyModal("Player Registration", "player_reg_steam_id", steamcmp, i)
	if err != nil {
		return errors.Wrap(err, "b.ReplyModal")
	}
	return nil
}

func handleSteamIDModalSubmit(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
) error {
	b.Acknowledge(i)
	steamID := i.ModalSubmitData().Components[0].(*discordgo.ActionsRow).
		Components[0].(*discordgo.TextInput).Value

	steamuser, err := steamapi.GetUser(steamID, b.Config.SteamAPIKey)
	if err != nil {
		return errors.Wrap(err, "steamapi.GetUser")
	}
	if steamuser == nil {
		b.ErrorFollowUp("Invalid Steam ID", "No steam user was found", i)
		return nil
	}
	slapid, err := slapshotapi.GetSlapID(
		steamuser.SteamID,
		b.Config.SlapshotAPIKey,
		b.Config.SlapshotAPIEnv,
	)
	if err != nil {
		return errors.Wrap(err, "slapshotapi.GetSlapID")
	}
	if slapid == 0 {
		b.ErrorFollowUp("Invalid Steam ID", "Steam account hasn't played slapshot", i)
		return nil
	}
	existingPlayer, err := models.GetPlayerBySlapID(ctx, tx, slapid)
	if err != nil {
		return errors.Wrap(err, "models.GetPlayerBySlapID")
	}
	if existingPlayer != nil {
		b.ErrorFollowUp("Invalid Steam ID", "Account already linked to a player", i)
		return nil
	}
	embed := &discordgo.MessageEmbed{
		Color: 0xeb7d34,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    steamuser.PersonaName,
			IconURL: steamuser.Avatar,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Steam User Found",
				Value:  fmt.Sprintf("__SlapID:__ %v", slapid),
				Inline: false,
			},
		},
	}
	msgcomps := []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					CustomID: fmt.Sprintf("confirm_slapid_%v", slapid),
					Label:    "Confirm",
					Style:    discordgo.SuccessButton,
				},
			},
		},
	}

	b.FollowUpComplex(embed, msgcomps, i)
	return nil
}

func handleSteamIDConfirm(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	slapid string,
) error {
	player, err := models.GetPlayerByDiscordID(ctx, tx, i.Member.User.ID)
	if err != nil {
		return errors.Wrap(err, "models.GetPlayerByDiscordID")
	}
	if player != nil {
		msg := fmt.Sprintf("__Player Name:__ %s\n__Slap ID:__ %v", player.Name, player.SlapID)
		b.Error("You are already registered", msg, i)
		return nil
	}
	regcmp := []discordgo.MessageComponent{
		components.TextInput("player_name", "Display Name", true, ""),
	}
	err = b.ReplyModal("Player Registration", "player_reg_display_name_"+slapid, regcmp, i)
	if err != nil {
		return errors.Wrap(err, "b.ReplyModal")
	}
	return nil
}

func handleDisplayNameSubmit(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	slapidstr string,
) error {
	b.Acknowledge(i)
	displayname := i.ModalSubmitData().Components[0].(*discordgo.ActionsRow).
		Components[0].(*discordgo.TextInput).Value
	slapid, err := strconv.ParseUint(slapidstr, 10, 0)
	if err != nil {
		return errors.Wrap(err, "strconv.ParseUint")
	}
	player, err := models.GetPlayerBySlapID(ctx, tx, uint32(slapid))
	if err != nil {
		return errors.Wrap(err, "models.GetPlayerBySlapID")
	}
	if player != nil {
		if player.DiscordID != "" {
			return errors.New("Player already linked to a discord account")
		}
		player.UpdateDiscordID(ctx, tx, i.Member.User.ID)
	} else {
		err = models.CreatePlayer(ctx, tx, uint32(slapid), i.Member.User.ID, displayname)
		if err != nil {
			// TODO: handle non unique player name
			if strings.Contains(err.Error(), "Display name must be unique") {
				b.ErrorFollowUp("Registration failed", err.Error(), i)
				return nil
			}
			return errors.Wrap(err, "models.CreatePlayer")
		}
	}
	b.FollowUp("Player registration successful!", i)
	return nil
}

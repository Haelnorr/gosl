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
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleSteamIDModalSubmit(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.Acknowledge(i, ack)
	steamID := i.ModalSubmitData().Components[0].(*discordgo.ActionsRow).
		Components[0].(*discordgo.TextInput).Value

	steamuser, err := steamapi.GetUser(steamID, b.Config.SteamAPIKey)
	if err != nil {
		return errors.Wrap(err, "steamapi.GetUser")
	}
	if steamuser == nil {
		return b.Error("Invalid Steam ID", "No steam user was found", i, true)
	}
	slapid, err := slapshotapi.GetSlapID(
		ctx,
		steamuser.SteamID,
		b.Config.SlapshotAPIConfig,
	)
	if err != nil {
		return errors.Wrap(err, "slapshotapi.GetSlapID")
	}
	if slapid == 0 {
		return b.Error("Invalid Steam ID", "Steam account hasn't played slapshot", i, true)
	}
	existingPlayer, err := models.GetPlayerBySlapID(ctx, tx, slapid)
	if err != nil {
		return errors.Wrap(err, "models.GetPlayerBySlapID")
	}
	if existingPlayer != nil {
		return b.Error("Invalid Steam ID", "Account already linked to a player", i, true)
	}
	contents := confirmSlapIDContents(steamuser, slapid)
	err = b.FollowUpComplex(contents, i, 60*time.Second)
	if err != nil {
		return errors.Wrap(err, "b.FollowUpComplex")
	}
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
		return b.Error("You are already registered", msg, i, false)
	}
	regcmp := []discordgo.MessageComponent{
		components.TextInput("player_name", "Display Name", true, "", 1, 32),
	}
	err = b.ReplyModal("Player Registration", "player_reg_display_name_"+slapid, regcmp, i)
	if err != nil {
		return errors.Wrap(err, "b.ReplyModal")
	}
	return nil
}

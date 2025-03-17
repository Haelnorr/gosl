package registrationchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleDisplayNameSubmit(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
	slapidstr string,
) error {
	b.Acknowledge(i, ack)
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
			if strings.Contains(err.Error(), "Display name must be unique") {
				b.Error("Registration failed", err.Error(), i, true)
				return nil
			}
			return errors.Wrap(err, "models.CreatePlayer")
		}
	}
	err = b.FollowUp("Player registration successful!", i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}
	return nil
}

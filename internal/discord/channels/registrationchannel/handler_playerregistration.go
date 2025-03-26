package registrationchannel

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/components"
	"gosl/internal/models"
	"gosl/pkg/db"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

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
		return b.Error("You are already registered", msg, i, false)
	}
	steamcmp := []discordgo.MessageComponent{
		components.TextInput("steam_id", "Steam ID", true, "", 1, 256),
	}

	err = b.ReplyModal("Player Registration", "player_reg_steam_id", steamcmp, i)
	if err != nil {
		return errors.Wrap(err, "b.ReplyModal")
	}
	return nil
}

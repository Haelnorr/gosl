package directmessages

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/components"
	"gosl/internal/discord/util"
	"gosl/pkg/db"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleColorButton(
	b *bot.Bot,
	i *discordgo.InteractionCreate,
) error {
	modalComps := []discordgo.MessageComponent{
		components.TextInput("color_hex", "Color Hex", true, "", 6, 6),
	}
	err := b.ReplyModal(
		"Set Team Color",
		fmt.Sprintf("set_color_modal_%s", i.Message.ID),
		modalComps, i)
	if err != nil {
		return errors.Wrap(err, "b.ReplyModal")
	}
	return nil
}

func handleSetTeamColor(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
	panelMsgID string,
) error {
	b.Acknowledge(i, ack)
	_, team, err := util.CheckPlayerIsManager(ctx, tx, i.User.ID)
	if err != nil {
		if strings.Contains(err.Error(), "VE:") {
			return b.Error("Interaction failed", err.Error(), i, *ack)
		}
		return errors.Wrap(err, "checkPlayerIsManager")
	}
	hexStr := i.ModalSubmitData().Components[0].(*discordgo.ActionsRow).
		Components[0].(*discordgo.TextInput).Value
	valid, _ := regexp.MatchString("^[0-9a-fA-F]{6}$", hexStr)
	if !valid {
		return b.Error("Failed to set color", "Invalid Hex code provided", i, *ack)
	}
	if hexStr == "181825" {
		return b.Error("Failed to set color", "You cannot use that color. It's literally the only color you cannot use. I'm actually kinda impressed. There are 16777215 different colors you could have picked from and you chose the only one I blacklisted. It's not even for a good reason, it's because im lazy and I picked a single color to be the default 'unset' color. Anyway, congrats or something. Pick another fucking color though please.", i, *ack)
	}
	err = team.SetColor(ctx, tx, hexStr)
	if err != nil {
		return errors.Wrap(err, "team.SetColor")
	}
	updateTeamManagerPanel(ctx, tx, b, team, panelMsgID, i.User.ID)
	err = b.FollowUp(fmt.Sprintf("Updated color for %s to #%s", team.Name, hexStr), i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}
	return nil
}

package transferapprovals

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

func NewTransferRequestMsg(ctx context.Context, b *bot.Bot) (*bot.DynamicMessage, error) {
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := b.Conn.RBegin(timeout, "NewTransferRequestMsg()")
	if err != nil {
		return nil, errors.Wrap(err, "b.Conn.RBegin")
	}
	defer tx.Rollback()
	channelID, err := models.GetChannel(ctx, tx, models.ChannelTransferApprovals)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetChannel")
	}
	msg := bot.NewDynamicMessage("Transfer Request", channelID, b)
	return msg, nil
}

func TransferRequestContents(
	ctx context.Context,
	tx db.SafeTX,
	pti *models.PlayerTeamInvite,
) (*bot.MessageContents, error) {
	embed := &discordgo.MessageEmbed{
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "Transfer Request",
				Value: fmt.Sprintf(`**%s has been invited to join %s!**`,
					pti.PlayerName, pti.TeamName),
				Inline: false,
			},
		},
	}
	msgcomps := []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					CustomID: fmt.Sprintf("approve_transfer_%v", pti.ID),
					Label:    "Approve application",
					Style:    discordgo.SuccessButton,
				},
				&discordgo.Button{
					CustomID: fmt.Sprintf("reject_transfer_%v", pti.ID),
					Label:    "Reject the application",
					Style:    discordgo.DangerButton,
				},
			},
		},
	}
	contents := &bot.MessageContents{
		Embed:      embed,
		Components: msgcomps,
	}
	return contents, nil
}

func updateRequestMsg(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	pti *models.PlayerTeamInvite,
	remove bool,
) error {
	reqMsg, err := b.GetDynamicMessage("Transfer request", i.Message.ID, i.ChannelID)
	if err != nil {
		return errors.Wrap(err, "b.GetDynamicMessage")
	}
	contents, err := TransferRequestContents(ctx, tx, pti)
	if err != nil {
		return errors.Wrap(err, "TransferRequestContents")
	}
	if remove {
		err = reqMsg.Delete(contents)
		if err != nil {
			return errors.Wrap(err, "reqMsg.Delete")
		}
	} else {
		err = reqMsg.Update(contents)
		if err != nil {
			return errors.Wrap(err, "reqMsg.Update")
		}
	}
	return nil
}

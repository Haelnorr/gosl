package directmessages

import (
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/models"

	"github.com/bwmarrin/discordgo"
)

func TeamInviteComponents(
	b *bot.Bot,
	invite *models.PlayerTeamInvite,
	panelMsgid string,
) (*bot.MessageContents, error) {
	msg := "This invite is no longer valid"
	inviteID := uint32(0)
	if invite != nil {
		msg = fmt.Sprintf("You've been invited to join ***%s***!", invite.TeamName)
		inviteID = invite.ID
	}
	embed := &discordgo.MessageEmbed{
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Invite to team",
				Value:  msg,
				Inline: false,
			},
		},
	}
	msgcomps := []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					CustomID: fmt.Sprintf("accept_invite_%v_%s", inviteID, panelMsgid),
					Label:    "Accept",
					Style:    discordgo.SuccessButton,
				},
				&discordgo.Button{
					CustomID: fmt.Sprintf("reject_invite_%v_%s", inviteID, panelMsgid),
					Label:    "Reject",
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

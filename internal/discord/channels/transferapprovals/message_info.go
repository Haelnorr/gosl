package transferapprovals

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"

	"github.com/bwmarrin/discordgo"
)

var infoMsg = &bot.Message{
	Label:       "Transfer Approvals Info",
	Purpose:     models.MsgTransferApprovalsInfo,
	GetContents: transferapprovalsinfoContents,
}

func transferapprovalsinfoContents(
	ctx context.Context,
	b *bot.Bot,
) (*bot.MessageContents, error) {
	contents := &bot.MessageContents{
		Embed: &discordgo.MessageEmbed{
			Title: "Transfer Approvals",
			Description: `
When players are invited to join a team after their application has been approved, they will require approval.
These will appear in this channel as transfer requests that require staff approval.
`,
		},
		Components: []discordgo.MessageComponent{},
	}
	return contents, nil
}

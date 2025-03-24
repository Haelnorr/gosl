package directmessages

import (
	"gosl/internal/discord/bot"
	"gosl/internal/models"
)

func expireInvite(
	b *bot.Bot,
	inviteMsgID string,
	userID string,
	invite *models.PlayerTeamInvite,
) {
	invMsg, err := b.GetDirectMessage(inviteMsgID, userID, "Team invite", 0, false)
	if err != nil {
		b.Logger.Warn().Err(err).Msg("Failed to get direct message")
		return
	}
	contents, err := TeamInviteComponents(b, invite, "")
	if err != nil {
		b.Logger.Warn().Err(err).Msg("Failed to get team invite components")
		return
	}
	err = invMsg.Expire(contents)
	if err != nil {
		b.Logger.Warn().Err(err).Msg("Failed to expire invite message")
		return
	}
}

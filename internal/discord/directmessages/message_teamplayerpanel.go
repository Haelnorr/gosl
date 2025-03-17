package directmessages

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"
)

// TODO: finish this
func TeamPlayerComponents(
	ctx context.Context,
	tx db.SafeTX,
	team *models.Team,
) (*bot.MessageContents, error) {
	// TODO: get team info and interactions for player view
	contents := &bot.MessageContents{
		Embed:      nil,
		Components: nil,
	}
	return contents, nil
}

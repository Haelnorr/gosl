package logchannel

import (
	"gosl/internal/discord/bot"
	"gosl/internal/models"
)

var LogChannel = &bot.Channel{
	Purpose: models.ChannelLog,
	Name:    "gosl-bot-log",
	Label:   "Log channel",
}

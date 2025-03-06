package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// TODO: find all logger calls and determine if logging to log file is needed

type logmsg struct {
	b       *Bot
	level   string
	message string
	color   int
}

func (b *Bot) setLogChannel(channelID string) {
	b.logchannel = channelID
}

func (b *Bot) Log() *logmsg {
	return &logmsg{b: b}
}

func (l *logmsg) Error(err error) {
	l.level = "Error"
	l.message = err.Error()
	l.color = 0xff0000
	l.b.createStaticMessage(l.logMsgContents(), l.b.logchannel)
}

func (l *logmsg) UserEvent(user *discordgo.Member, msg string) {
	l.level = "User Event"
	evtmsg := `
**User:** %s

**Event:**
%s
`
	l.message = fmt.Sprintf(evtmsg, user.User.Username, msg)
	l.color = 0x0096FF
	l.b.createStaticMessage(l.logMsgContents(), l.b.logchannel)
}

func (l *logmsg) Info(msg string) {
	l.level = "Info"
	l.message = msg
	l.color = 0x00ff00
	l.b.createStaticMessage(l.logMsgContents(), l.b.logchannel)
}

func (l *logmsg) logMsgContents() MessageContents {
	return func() (
		string,
		*discordgo.MessageEmbed,
		[]discordgo.MessageComponent,
	) {
		return "",
			&discordgo.MessageEmbed{
				Title:       l.level,
				Description: l.message,
				Color:       l.color,
			},
			nil
	}
}

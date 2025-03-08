package util

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// TODO: find all logger calls and determine if logging to log file is needed

// Log message object for logging to discord channel
type logmsg struct {
	b       *Bot
	level   string
	message string
	color   int
}

// Create a new logmsg
func (b *Bot) Log() *logmsg {
	return &logmsg{b: b}
}

// Send the logmsg as an error message
func (l *logmsg) Error(msg string, err error) {
	l.level = "Error"
	l.message = msg + "\n" + err.Error()
	l.color = 0xff0000
	createComplexMessage(l.logMsgContents(), l.b.Logchannel, l.b.Session)
}

// Send the logmsg as a user event
func (l *logmsg) UserEvent(user *discordgo.Member, msg string) {
	l.level = "User Event"
	evtmsg := `
**User:** %s

**Event:**
%s
`
	l.message = fmt.Sprintf(evtmsg, user.User.Username, msg)
	l.color = 0x0096FF
	createComplexMessage(l.logMsgContents(), l.b.Logchannel, l.b.Session)
}

// Send the logmsg as an info event
func (l *logmsg) Info(msg string) {
	l.level = "Info"
	l.message = msg
	l.color = 0x00ff00
	createComplexMessage(l.logMsgContents(), l.b.Logchannel, l.b.Session)
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

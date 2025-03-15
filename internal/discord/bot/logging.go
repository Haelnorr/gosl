package bot

import (
	"fmt"
	"gosl/internal/models"

	"github.com/bwmarrin/discordgo"
)

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

// Helper function to log an error with all three logging methods.
// Logs to primary logger, log channel, and reponds to the interaction
func (b *Bot) TripleError(
	msg string,
	err error,
	i *discordgo.InteractionCreate,
) {
	b.Logger.Error().Err(err).Msg(msg)
	b.Error(msg, err.Error(), i)
	b.Log().Error(msg, err)
}

// Helper function to log an error with the two main logging methods.
// Logs to primary logger and the log channel
func (b *Bot) DoubleError(
	msg string,
	err error,
) {
	b.Logger.Error().Err(err).Msg(msg)
	b.Log().Error(msg, err)
}

// Send the logmsg as an error message
func (l *logmsg) Error(msg string, err error) {
	l.level = "Error"
	l.message = msg + "\n" + err.Error()
	l.color = 0xff0000
	l.b.Channels[models.ChannelLog].SendMessage(l.logMsgContents())
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
	l.b.Channels[models.ChannelLog].SendMessage(l.logMsgContents())
}

// Send the logmsg as an info event
func (l *logmsg) Info(msg string) {
	l.level = "Info"
	l.message = msg
	l.color = 0x00ff00
	l.b.Channels[models.ChannelLog].SendMessage(l.logMsgContents())
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

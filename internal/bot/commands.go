package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func (b *Bot) setupCommands() {
	b.commands = []*command{
		cmdTest(b),
		cmdUploadLogs(b),
	}
}

func (b *Bot) registerCommands() error {
	for _, cmd := range b.commands {
		_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", &discordgo.ApplicationCommand{
			Name:        cmd.Name,
			Description: cmd.Description,
			Options:     cmd.Options,
		})
		if err != nil {
			b.logger.Error().Err(err).Str("command", cmd.Name).Msg("Failed to register command")
			return errors.Wrapf(err, "b.session.ApplicationCommandCreate: %s", cmd.Name)
		}
		b.logger.Debug().Str("command", cmd.Name).Msg("Registering command")
	}

	b.session.AddHandler(b.handleInteractions)
	return nil
}

func (b *Bot) handleInteractions(s *discordgo.Session, i *discordgo.InteractionCreate) {
	for _, cmd := range b.commands {
		if i.ApplicationCommandData().Name == cmd.Name {
			cmd.Handler(s, i)
			b.logger.Debug().Str("command", cmd.Name).Msg("Handled command")
			return
		}
	}
}

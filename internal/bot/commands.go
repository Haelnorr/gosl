package bot

import (
	"gosl/internal/commands"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

type command struct {
	Name        string
	Description string
	Handler     commands.Handler
}

func (b *Bot) setupCommands() {
	b.commands = []*command{
		{
			Name:        "hi",
			Description: "Say hi to the bot",
			Handler:     commands.Hi(b.logger),
		},
		{
			Name:        "bye",
			Description: "Say goodbye to the bot",
			Handler:     commands.Bye(b.logger),
		},
	}
}

func (b *Bot) registerCommands() error {
	for _, cmd := range b.commands {
		_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", &discordgo.ApplicationCommand{
			Name:        cmd.Name,
			Description: cmd.Description,
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

func (b *Bot) clearCommands() error {
	commands, err := b.session.ApplicationCommands(b.session.State.User.ID, "")
	if err != nil {
		return errors.Wrap(err, "fetching existing commands")
	}

	for _, cmd := range commands {
		err := b.session.ApplicationCommandDelete(b.session.State.User.ID, "", cmd.ID)
		if err != nil {
			b.logger.Error().Err(err).Str("command", cmd.Name).Msg("Failed to delete command")
		} else {
			b.logger.Debug().Str("command", cmd.Name).Msg("Deleted old command")
		}
	}

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

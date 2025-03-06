package bot

import (
	"context"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func (b *Bot) setupCommands() {
	b.commands = []*command{
		cmdTest(b),
		cmdUploadLogs(b),
	}
}

func (b *Bot) registerCommands(
	wg *sync.WaitGroup,
	errch chan error,
	ctx context.Context,
) error {
	defer wg.Done()
	for _, cmd := range b.commands {
		_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", &discordgo.ApplicationCommand{
			Name:        cmd.Name,
			Description: cmd.Description,
			Options:     cmd.Options,
		})
		if err != nil {
			b.logger.Error().Err(err).Str("command", cmd.Name).Msg("Failed to register command")
			errch <- errors.Wrapf(err, "b.session.ApplicationCommandCreate: %s", cmd.Name)
			continue
		}
		b.logger.Debug().Str("command", cmd.Name).Msg("Registering command")
	}

	b.session.AddHandler(b.handleCommandInteractions(ctx))
	b.logger.Info().Msg("Finished registering commands")
	return nil
}

func (b *Bot) handleCommandInteractions(ctx context.Context) handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			for _, cmd := range b.commands {
				if i.ApplicationCommandData().Name == cmd.Name {
					cmd.Handler(s, i)
					b.logger.Debug().Str("command", cmd.Name).Msg("Handled command")
					return
				}
			}
		}
	}
}

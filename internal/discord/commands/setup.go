package commands

import (
	"context"
	"gosl/internal/discord/util"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type Command struct {
	Name        string
	Description string
	Handler     util.Handler
	Options     []*discordgo.ApplicationCommandOption
}

// Get all the commands registered
func getCommands(b *util.Bot) []*Command {
	return []*Command{
		cmdUploadLogs(b),
	}
}

// Setup the bot commands
func Setup(
	wg *sync.WaitGroup,
	errch chan error,
	ctx context.Context,
	b *util.Bot,
) {
	defer wg.Done()
	commands := getCommands(b)
	for _, cmd := range commands {
		_, err := b.Session.ApplicationCommandCreate(
			b.Session.State.User.ID,
			"",
			&discordgo.ApplicationCommand{
				Name:        cmd.Name,
				Description: cmd.Description,
				Options:     cmd.Options,
			})
		if err != nil {
			b.Logger.Error().Err(err).Str("command", cmd.Name).
				Msg("Failed to register command")
			errch <- errors.Wrapf(err, "b.session.ApplicationCommandCreate: %s", cmd.Name)
			continue
		}
		b.Logger.Debug().Str("command", cmd.Name).Msg("Registering command")
	}

	b.Session.AddHandler(handleCommandInteractions(ctx, b.Logger, commands))
	b.Logger.Info().Msg("Finished registering commands")
}

// Handle the command interactions
func handleCommandInteractions(
	ctx context.Context,
	logger *zerolog.Logger,
	commands []*Command,
) util.Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			for _, cmd := range commands {
				if i.ApplicationCommandData().Name == cmd.Name {
					cmd.Handler(s, i)
					logger.Debug().Str("command", cmd.Name).Msg("Handled command")
					return
				}
			}
		}
	}
}

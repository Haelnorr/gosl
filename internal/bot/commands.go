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
	Options     []*discordgo.ApplicationCommandOption
}

func (b *Bot) setupCommands() {
	b.commands = []*command{
		{
			Name:        "test",
			Description: "Test",
			Handler:     commands.Test(b.logger, b.files),
		},
		{
			Name:        "uploadlogs",
			Description: "Upload match logs",
			Handler:     commands.UploadLogs(b.logger, b.files),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionAttachment,
					Name:        "period1",
					Description: "Period 1 log file",
					Required:    true,
				},
				// {
				// 	Type:        discordgo.ApplicationCommandOptionAttachment,
				// 	Name:        "period2",
				// 	Description: "Period 2 log file",
				// 	Required:    true,
				// },
				// {
				// 	Type:        discordgo.ApplicationCommandOptionAttachment,
				// 	Name:        "period3",
				// 	Description: "Period 3 log file",
				// 	Required:    true,
				// },
			},
		},
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

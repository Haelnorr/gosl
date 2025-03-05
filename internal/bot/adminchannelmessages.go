package bot

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func (b *Bot) selectLogChannelContents(ctx context.Context) (MessageContents, error) {
	b.logger.Debug().Msg("Setting up select log channel components")
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := b.conn.Begin(timeout)
	if err != nil {
		return nil, errors.Wrap(err, "b.conn.Begin")
	}
	defer tx.Rollback()
	b.logger.Debug().Msg("Getting default values for select log channel components")
	logChannelID, err := queryChannelForPurpose(ctx, tx, channelLog)
	if err != nil {
		return nil, errors.Wrap(err, "getChannelForPurpose")
	}
	tx.Commit()
	var defaultValues []discordgo.SelectMenuDefaultValue
	if b.checkChannelExists(logChannelID) {
		defaultValues = append(defaultValues, discordgo.SelectMenuDefaultValue{
			ID:   logChannelID,
			Type: discordgo.SelectMenuDefaultValueChannel,
		})
	}
	return func() (
		string,
		*discordgo.MessageEmbed,
		[]discordgo.MessageComponent,
	) {

		b.logger.Debug().Msg("Retrieving select log channel components")
		return "",
			&discordgo.MessageEmbed{
				Title:       "Log output channel",
				Description: `Select the channel to output bot logs to`,
				Color:       0x00ff00, // Green color
			},
			[]discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.SelectMenu{
							MenuType:      discordgo.ChannelSelectMenu,
							CustomID:      "log_channel_select",
							Placeholder:   "Select the channel for log output",
							DefaultValues: defaultValues,
						},
					},
				},
			}
	}, nil
}
func (b *Bot) selectAdminRolesContents(ctx context.Context) (MessageContents, error) {
	b.logger.Debug().Msg("Setting up select admin roles components")
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := b.conn.Begin(timeout)
	if err != nil {
		return nil, errors.Wrap(err, "b.conn.Begin")
	}
	defer tx.Rollback()
	// TODO: get the values from the DB and set them in defaults
	b.logger.Debug().Msg("Getting default values for select admin roles components")
	tx.Commit()

	return func() (
		string,
		*discordgo.MessageEmbed,
		[]discordgo.MessageComponent,
	) {
		b.logger.Debug().Msg("Retreiving select admin roles components")
		return "",
			&discordgo.MessageEmbed{
				Title:       "Admin roles",
				Description: `Select the roles that should have admin access`,
				Color:       0x00ff00, // Green color
			},
			[]discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.SelectMenu{
							MenuType:    discordgo.RoleSelectMenu,
							CustomID:    "admin_role_select",
							Placeholder: "Select admin roles",
							MaxValues:   10,
						},
					},
				},
			}
	}, nil
}
func (b *Bot) selectManagerRolesContents(ctx context.Context) (MessageContents, error) {
	b.logger.Debug().Msg("Setting up select manager roles components")
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := b.conn.Begin(timeout)
	if err != nil {
		return nil, errors.Wrap(err, "b.conn.Begin")
	}
	defer tx.Rollback()
	// TODO: get the values from the DB and set them in defaults
	b.logger.Debug().Msg("Getting default values for select manager roles components")
	tx.Commit()

	return func() (
		string,
		*discordgo.MessageEmbed,
		[]discordgo.MessageComponent,
	) {
		b.logger.Debug().Msg("Retrieving select manager roles components")
		return "",
			&discordgo.MessageEmbed{
				Title:       "Manager roles",
				Description: `Select the roles that should have manager access`,
				Color:       0x00ff00, // Green color
			},
			[]discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.SelectMenu{
							MenuType:    discordgo.RoleSelectMenu,
							CustomID:    "manager_role_select",
							Placeholder: "Select manager roles",
							MaxValues:   10,
						},
					},
				},
			}
	}, nil
}

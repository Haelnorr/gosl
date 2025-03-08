package commands

import (
	"encoding/json"
	"gosl/internal/discord/messages"
	"gosl/internal/discord/util"
	"gosl/internal/gamelogs"
	"io"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

func cmdUploadLogs(b *util.Bot) *Command {
	return &Command{
		Name:        "uploadlogs",
		Description: "Upload match logs",
		Handler:     handleUploadLogs(b),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Name:        "period1",
				Description: "Period 1 log file",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Name:        "period2",
				Description: "Period 2 log file",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Name:        "period3",
				Description: "Period 3 log file",
				Required:    true,
			},
		},
	}
}

func handleUploadLogs(
	b *util.Bot,
) util.Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// TODO: check user has permission to upload logs

		// get message attachments
		attachments := i.ApplicationCommandData().Resolved.Attachments
		logs := []*gamelogs.Gamelog{}

		for _, attachment := range attachments {
			// TODO: make this concurrent
			resp, err := http.Get(attachment.URL)
			if err != nil {
				msg := "Failed to download attachment: " + attachment.Filename
				b.Logger.Error().Err(err).Str("url", attachment.URL).Msg(msg)
				b.Error(msg, err.Error(), s, i)
				b.Log().Error(msg, err)
				return
			}
			defer resp.Body.Close()

			content, err := io.ReadAll(resp.Body)
			if err != nil {
				msg := "Failed to read file: " + attachment.Filename
				b.Logger.Error().Err(err).Str("filename", attachment.Filename).Msg(msg)
				b.Error(msg, err.Error(), s, i)
				b.Log().Error(msg, err)
				return
			}
			var log gamelogs.Gamelog
			err = json.Unmarshal(content, &log)
			if err != nil {
				msg := "Failed to parse log file: " + attachment.Filename
				b.Logger.Error().Err(err).Str("filename", attachment.Filename).Msg(msg)
				b.Error(msg, err.Error(), s, i)
				b.Log().Error(msg, err)
				return
			}
			logs = append(logs, &log)
		}

		// TODO: actually do something with the log data

		err := messages.ReplyEphemeral("Log files uploaded", s, i, b.Logger)
		if err != nil {
			b.Logger.Error().Err(err).Msg("Failed to reply to interaction")
		}
	}
}

package bot

import (
	"encoding/json"
	"gosl/internal/gamelogs"
	"io"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

func cmdUploadLogs(b *Bot) *command {
	return buildCommand(
		"uploadlogs",
		"Upload match logs",
		handleUploadLogs(b),
		&discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionAttachment,
			Name:        "period1",
			Description: "Period 1 log file",
			Required:    true,
		},
		&discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionAttachment,
			Name:        "period2",
			Description: "Period 2 log file",
			Required:    true,
		},
		&discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionAttachment,
			Name:        "period3",
			Description: "Period 3 log file",
			Required:    true,
		},
	)
}

// TODO: rewrite responses, this shit broke
func handleUploadLogs(b *Bot) handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// TODO: check user has permission to upload logs

		// get message attachments
		attachments := i.ApplicationCommandData().Resolved.Attachments
		logs := []*gamelogs.Gamelog{}

		for _, attachment := range attachments {
			resp, err := http.Get(attachment.URL)
			if err != nil {
				b.logger.Error().Err(err).Str("url", attachment.URL).
					Msg("Failed to download attachment")
				errorResponse("Failed to download attachment", &attachment.Filename, b.files, s, i)
				return
			}
			defer resp.Body.Close()

			content, err := io.ReadAll(resp.Body)
			if err != nil {
				b.logger.Error().Err(err).Str("filename", attachment.Filename).Msg("Failed to read attachment contents")
				errorResponse("Failed to read file", &attachment.Filename, b.files, s, i)
				return
			}
			var log gamelogs.Gamelog
			err = json.Unmarshal(content, &log)
			if err != nil {
				b.logger.Error().Err(err).Str("filename", attachment.Filename).Msg("Failed to parse log")
				errorResponse("Failed to parse log file", &attachment.Filename, b.files, s, i)
				return
			}
			logs = append(logs, &log)
		}

		// TODO: actually do something with the log data

		replyMessage("Log files uploaded", b.logger, s, i)
	}
}

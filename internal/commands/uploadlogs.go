package commands

import (
	"encoding/json"
	"gosl/internal/gamelogs"
	"io"
	"io/fs"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

// TODO: rewrite responses, this shit broke
func UploadLogs(logger *zerolog.Logger, files *fs.FS) Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// TODO: check user has permission to upload logs

		// get message attachments
		attachments := i.ApplicationCommandData().Resolved.Attachments
		logs := []*gamelogs.Gamelog{}

		for _, attachment := range attachments {
			resp, err := http.Get(attachment.URL)
			if err != nil {
				logger.Error().Err(err).Str("url", attachment.URL).
					Msg("Failed to download attachment")
				errorResponse("Failed to download attachment", &attachment.Filename, files, s, i)
				return
			}
			defer resp.Body.Close()

			content, err := io.ReadAll(resp.Body)
			if err != nil {
				logger.Error().Err(err).Str("filename", attachment.Filename).Msg("Failed to read attachment contents")
				errorResponse("Failed to read file", &attachment.Filename, files, s, i)
				return
			}
			var log gamelogs.Gamelog
			err = json.Unmarshal(content, &log)
			if err != nil {
				logger.Error().Err(err).Str("filename", attachment.Filename).Msg("Failed to parse log")
				errorResponse("Failed to parse log file", &attachment.Filename, files, s, i)
				return
			}
			logs = append(logs, &log)
		}

		// TODO: actually do something with the log data

		replyMessage("Log files uploaded", logger, s, i)
	}
}

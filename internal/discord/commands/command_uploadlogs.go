package commands

import (
	"context"
	"encoding/json"
	"gosl/internal/discord/bot"
	"gosl/internal/gamelogs"
	"gosl/internal/models"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func cmdUploadLogs(ctx context.Context, b *bot.Bot) *Command {
	return &Command{
		Name:        "uploadlogs",
		Description: "Upload match logs",
		Handler:     handleUploadLogs(ctx, b),
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
	ctx context.Context,
	b *bot.Bot,
) bot.Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		b.Acknowledge(i, nil)
		timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		tx, err := b.Conn.Begin(timeout, "Handle /uploadlogs command")
		if err != nil {
			b.TripleError("Log upload failed", err, i, true)
			return
		}
		defer tx.Rollback()
		member := i.Member
		if member == nil {
			member, err = s.GuildMember(b.Config.DiscordGuildID, i.User.ID)
			if err != nil {
				b.TripleError("Log upload failed", err, i, true)
				return
			}
		}
		isLeagueMgr, err := models.MemberHasPermission(ctx, tx, s,
			b.Config.DiscordGuildID, member, models.PermLeagueManager)
		if err != nil {
			b.TripleError("Log upload failed", err, i, true)
			return
		}
		if !isLeagueMgr {
			b.Forbidden(i, true)
			return
		}

		// get message attachments
		attachments := i.ApplicationCommandData().Resolved.Attachments
		logs := []*gamelogs.Gamelog{}

		for _, attachment := range attachments {
			if strings.Contains(attachment.ContentType, "application/json") {
				err = b.Error("Logs upload failed", "This attachment is not a JSON", i, true)
				if err != nil {
					b.Logger.Error().Err(err).Msg("Failed to notify user of validation error")
				}
				return
			}
			// TODO: make this concurrent
			resp, err := http.Get(attachment.URL)
			if err != nil {
				msg := "Failed to download attachment: " + attachment.URL
				b.TripleError(msg, err, i, true)
				return
			}
			defer resp.Body.Close()

			content, err := io.ReadAll(resp.Body)
			if err != nil {
				msg := "Failed to read file: " + attachment.Filename
				b.TripleError(msg, err, i, true)
				return
			}
			var log gamelogs.Gamelog
			err = json.Unmarshal(content, &log)
			if err != nil {
				msg := "Failed to parse log file: " + attachment.Filename
				b.TripleError(msg, err, i, true)
				return
			}
			logs = append(logs, &log)
		}

		// TODO: actually do something with the log data

		tx.Commit()
		err = b.FollowUp("Log files uploaded", i)
		if err != nil {
			b.Logger.Error().Err(err).Msg("Failed to reply to interaction")
		}
	}
}

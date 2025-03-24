package commands

import (
	"bytes"
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func cmdUploadLogo(ctx context.Context, b *bot.Bot) *Command {
	return &Command{
		Name:        "uploadlogo",
		Description: "Upload team logo",
		Handler:     handleUploadLogo(ctx, b),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Name:        "logo",
				Description: "Team Logo",
				Required:    true,
			},
		},
	}
}

func handleUploadLogo(
	ctx context.Context,
	b *bot.Bot,
) bot.Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		b.Acknowledge(i, nil)
		timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		tx, err := b.Conn.Begin(timeout, "Handle /uploadlogo command")
		if err != nil {
			b.TripleError("Logo upload failed", err, i, true)
			return
		}
		defer tx.Rollback()
		member := i.Member
		if member == nil {
			member, err = s.GuildMember(b.Config.DiscordGuildID, i.User.ID)
			if err != nil {
				b.TripleError("Logo upload failed", err, i, true)
				return
			}
		}
		player, err := models.GetPlayerByDiscordID(ctx, tx, member.User.ID)
		if err != nil {
			b.TripleError("Logo upload failed", err, i, true)
			return
		}
		if player == nil {
			err = b.Error("Logo upload failed", "You are not registered as a player", i, true)
			if err != nil {
				b.Logger.Error().Err(err).Msg("Failed to notify user of validation error")
			}
			return
		}
		team, err := player.CurrentTeam(ctx, tx)
		if err != nil {
			b.TripleError("Logo upload failed", err, i, true)
			return
		}
		if team == nil {
			err = b.Error("Logo upload failed", "You are not on a team", i, true)
			if err != nil {
				b.Logger.Error().Err(err).Msg("Failed to notify user of validation error")
			}
			return
		}
		if team.ManagerID != player.ID {
			err = b.Error("Logo upload failed", "You are not the manager of this team", i, true)
			if err != nil {
				b.Logger.Error().Err(err).Msg("Failed to notify user of validation error")
			}
			return
		}

		// get message attachments
		attachments := i.ApplicationCommandData().Resolved.Attachments
		var logo io.Reader
		fileext := ""
		for _, att := range attachments {
			if !strings.Contains(att.ContentType, "image/") {
				err = b.Error("Logo upload failed", "This attachment is not an image", i, true)
				if err != nil {
					b.Logger.Error().Err(err).Msg("Failed to notify user of validation error")
				}
				return
			}
			fileext = strings.TrimPrefix(att.ContentType, "image/")
			// Limit att size to 10MB
			if att.Size > 10000000 {
				err = b.Error("Logo upload failed", "Max attachment size is 10MB", i, true)
				if err != nil {
					b.Logger.Error().Err(err).Msg("Failed to notify user of validation error")
				}
				return
			}
			if att.Height != att.Width {
				err = b.Error("Logo upload failed", "Height and Width must be the same", i, true)
				if err != nil {
					b.Logger.Error().Err(err).Msg("Failed to notify user of validation error")
				}
				return
			}
			resp, err := http.Get(att.URL)
			if err != nil {
				b.TripleError("Logo upload failed", err, i, true)
				return
			}
			buf := new(bytes.Buffer)
			_, err = io.Copy(buf, resp.Body)
			if err != nil {
				b.TripleError("Logo upload failed", err, i, true)
				return
			}
			resp.Body.Close()
			logo = buf
		}
		if logo == nil {
			err = b.Error("Logo upload failed", "No file was found", i, true)
			if err != nil {
				b.Logger.Error().Err(err).Msg("Failed to notify user of validation error")
			}
			return
		}
		logoChan := b.Channels[models.ChannelTeamLogos]

		msg := fmt.Sprintf("Team Logo for %s", team.TeamName)
		now := time.Now().Unix()
		filename := fmt.Sprintf("%s_logo_%v.%s", team.TeamName, now, fileext)
		logoMsg, err := logoChan.SendFile(msg, filename, logo)
		if err != nil {
			b.TripleError("Logo upload failed", err, i, true)
			return
		}

		logoURL := logoMsg.Attachments[0].URL
		err = models.UploadTeamLogo(ctx, tx, logoURL, logoMsg.ChannelID, logoMsg.ID, team.TeamID)
		if err != nil {
			b.TripleError("Logo upload failed", err, i, true)
			return
		}

		err = b.FollowUp("Logo uploaded. Refresh the team panel to view it", i)
		if err != nil {
			b.Logger.Error().Err(err).Msg("Failed to reply to interaction")
		}
		tx.Commit()
	}
}

package bot

import (
	"bytes"
	"io/fs"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

func replyMessage(
	msg string,
	logger *zerolog.Logger,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
		},
	})
	if err != nil {
		logger.Error().Err(err).Str("msg", msg).Msg("Failed to respond")
		return err
	}
	return nil
}

func emphemeralMessage(
	msg string,
	logger *zerolog.Logger,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		logger.Error().Err(err).Str("msg", msg).Msg("Failed to respond")
		return err
	}
	return nil
}

func getAsset(name string, files *fs.FS) (*discordgo.File, error) {
	fileData, err := fs.ReadFile(*files, "assets/"+name)
	if err != nil {
		return nil, err
	}
	file := &discordgo.File{
		Name:   "error.png",
		Reader: bytes.NewReader(fileData),
	}
	return file, nil
}

package bot

import "github.com/bwmarrin/discordgo"

type handler = func(s *discordgo.Session, i *discordgo.InteractionCreate)

type command struct {
	Name        string
	Description string
	Handler     handler
	Options     []*discordgo.ApplicationCommandOption
}

func buildCommand(
	name string,
	desc string,
	handler handler,
	options ...*discordgo.ApplicationCommandOption,
) *command {
	return &command{
		Name:        name,
		Description: desc,
		Handler:     handler,
		Options:     options,
	}
}

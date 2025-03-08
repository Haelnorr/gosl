package util

import "github.com/bwmarrin/discordgo"

// Function that handles a user interaction
type Handler = func(s *discordgo.Session, i *discordgo.InteractionCreate)

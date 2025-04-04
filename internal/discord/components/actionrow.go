package components

import "github.com/bwmarrin/discordgo"

// Wraps the given components in an ActionRow
func ActionRow(comps ...discordgo.MessageComponent) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: comps,
		},
	}
}

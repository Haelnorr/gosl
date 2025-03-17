package components

import "github.com/bwmarrin/discordgo"

func TextInput(
	customID string,
	label string,
	required bool,
	value string,
	minlen int,
	maxlen int,
) *discordgo.ActionsRow {
	return &discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			&discordgo.TextInput{
				CustomID:    customID,
				Label:       label,
				Style:       discordgo.TextInputShort,
				Placeholder: label + "...",
				Required:    required,
				Value:       value,
				MinLength:   minlen,
				MaxLength:   maxlen,
			},
		},
	}
}

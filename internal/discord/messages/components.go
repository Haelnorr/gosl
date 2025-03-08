package messages

import "github.com/bwmarrin/discordgo"

// Return ChannelSelectMenu component wrapped in an ActionsRow
func ChannelSelect(
	customid string,
	placeholder string,
	defaults []discordgo.SelectMenuDefaultValue,
	minValues int,
	maxValues int,
	channelTypes []discordgo.ChannelType,
) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.SelectMenu{
					MenuType:      discordgo.ChannelSelectMenu,
					CustomID:      customid,
					Placeholder:   placeholder,
					DefaultValues: defaults,
					MinValues:     &minValues,
					MaxValues:     maxValues,
					ChannelTypes:  channelTypes,
				},
			},
		},
	}
}

// Return RoleSelectMenu component wrapped in an ActionsRow
func RoleSelect(
	customid string,
	placeholder string,
	defaults []discordgo.SelectMenuDefaultValue,
	minValues int,
	maxValues int,
) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.SelectMenu{
					MenuType:      discordgo.RoleSelectMenu,
					CustomID:      customid,
					Placeholder:   placeholder,
					DefaultValues: defaults,
					MinValues:     &minValues,
					MaxValues:     maxValues,
				},
			},
		},
	}
}

// Return RoleSelectMenu component wrapped in an ActionsRow
func StringSelect(
	customid string,
	placeholder string,
	options []discordgo.SelectMenuOption,
	minValues int,
	maxValues int,
) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.SelectMenu{
					MenuType:    discordgo.StringSelectMenu,
					CustomID:    customid,
					Placeholder: placeholder,
					Options:     options,
					MinValues:   &minValues,
					MaxValues:   maxValues,
				},
			},
		},
	}
}

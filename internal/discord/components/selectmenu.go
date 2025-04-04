package components

import "github.com/bwmarrin/discordgo"

// Return ChannelSelectMenu component
func ChannelSelect(
	customid string,
	placeholder string,
	defaults []discordgo.SelectMenuDefaultValue,
	minValues int,
	maxValues int,
	channelTypes []discordgo.ChannelType,
) *discordgo.SelectMenu {
	return &discordgo.SelectMenu{
		MenuType:      discordgo.ChannelSelectMenu,
		CustomID:      customid,
		Placeholder:   placeholder,
		DefaultValues: defaults,
		MinValues:     &minValues,
		MaxValues:     maxValues,
		ChannelTypes:  channelTypes,
	}
}

// Return RoleSelectMenu component
func RoleSelect(
	customid string,
	placeholder string,
	defaults []discordgo.SelectMenuDefaultValue,
	minValues int,
	maxValues int,
) *discordgo.SelectMenu {
	return &discordgo.SelectMenu{
		MenuType:      discordgo.RoleSelectMenu,
		CustomID:      customid,
		Placeholder:   placeholder,
		DefaultValues: defaults,
		MinValues:     &minValues,
		MaxValues:     maxValues,
	}
}

// Return RoleSelectMenu component
func StringSelect(
	customid string,
	placeholder string,
	options []discordgo.SelectMenuOption,
	minValues int,
	maxValues int,
	disabled bool,
) *discordgo.SelectMenu {
	return &discordgo.SelectMenu{
		MenuType:    discordgo.StringSelectMenu,
		CustomID:    customid,
		Placeholder: placeholder,
		Options:     options,
		MinValues:   &minValues,
		MaxValues:   maxValues,
		Disabled:    disabled,
	}
}

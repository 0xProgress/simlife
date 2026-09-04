package economy

import "github.com/bwmarrin/discordgo"

// getFloatOption extracts a float64 option from the interaction.
func getFloatOption(i *discordgo.InteractionCreate, name string) (float64, bool) {
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == name && opt.Type == discordgo.ApplicationCommandOptionNumber {
			return opt.FloatValue(), true
		}
	}
	return 0, false
}

// getUserIDOption extracts a user ID string from the interaction options.
func getUserIDOption(i *discordgo.InteractionCreate, name string) (string, bool) {
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == name && opt.Type == discordgo.ApplicationCommandOptionUser {
			if uid, ok := opt.Value.(string); ok {
				return uid, true
			}
		}
	}
	return "", false
}

// getStringOption extracts a string option from the interaction.
func getStringOption(i *discordgo.InteractionCreate, name string) (string, bool) {
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == name && opt.Type == discordgo.ApplicationCommandOptionString {
			if val, ok := opt.Value.(string); ok {
				return val, true
			}
		}
	}
	return "", false
}
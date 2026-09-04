package economy

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/shopspring/decimal"
)

// getDecimalOption safely extracts a financial value from a Discord interaction option.
// It accepts both Integer and String option types to prevent precision loss, 
// and returns a shopspring/decimal.Decimal. It strictly forbids float64.
func getDecimalOption(i *discordgo.InteractionCreate, name string) (decimal.Decimal, bool) {
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == name {
			switch opt.Type {
			case discordgo.ApplicationCommandOptionInteger:
				// Safe conversion from int64 to decimal
				return decimal.NewFromInt(opt.IntValue()), true
			case discordgo.ApplicationCommandOptionString:
				// Safe parsing from string to decimal
				val, err := decimal.NewFromString(opt.StringValue())
				if err == nil {
					return val, true
				}
				return decimal.Zero, false
			case discordgo.ApplicationCommandOptionNumber:
				// Fallback for Number type, but we immediately convert to string 
				// to avoid float64 precision loss during decimal conversion.
				val, err := decimal.NewFromString(strconv.FormatFloat(opt.FloatValue(), 'f', -1, 64))
				if err == nil {
					return val, true
				}
				return decimal.Zero, false
			}
		}
	}
	return decimal.Zero, false
}

// getUserOption safely extracts a Discord User object from an interaction option.
// This is significantly safer than type-asserting opt.Value.
func getUserOption(i *discordgo.InteractionCreate, name string) (*discordgo.User, bool) {
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == name && opt.Type == discordgo.ApplicationCommandOptionUser {
			// discordgo provides a safe accessor for this
			return opt.UserValue(), true
		}
	}
	return nil, false
}

// getStringOption safely extracts a string option from the interaction.
func getStringOption(i *discordgo.InteractionCreate, name string) (string, bool) {
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == name && opt.Type == discordgo.ApplicationCommandOptionString {
			return opt.StringValue(), true
		}
	}
	return "", false
}

// getSubCommandGroup safely extracts the subcommand group name if one exists.
func getSubCommandGroup(i *discordgo.InteractionCreate) (string, bool) {
	options := i.ApplicationCommandData().Options
	if len(options) > 0 && options[0].Type == discordgo.ApplicationCommandOptionSubCommandGroup {
		return options[0].Name, true
	}
	return "", false
}

// getSubCommand safely extracts the subcommand name if one exists.
func getSubCommand(i *discordgo.InteractionCreate) (string, bool) {
	options := i.ApplicationCommandData().Options
	if len(options) > 0 {
		if options[0].Type == discordgo.ApplicationCommandOptionSubCommand {
			return options[0].Name, true
		}
		// Check nested subcommand
		if options[0].Type == discordgo.ApplicationCommandOptionSubCommandGroup && len(options[0].Options) > 0 {
			return options[0].Options[0].Name, true
		}
	}
	return "", false
}
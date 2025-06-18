package bot

import "github.com/bwmarrin/discordgo"

const (
	Campaign string = "campaign"
)

var Commands map[string]*discordgo.ApplicationCommand = map[string]*discordgo.ApplicationCommand{
	Campaign: {
		Name:        Campaign,
		Description: "Create or update a campaign to track",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "title",
				Description: "The title of your campaign.",
				Required:    true,
			},
		},
	},
}

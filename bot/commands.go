package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/taberiv/gopher-dm/sqlite"
)

const (
	Campaign      string = "campaign"
	ListCampaigns string = "list-campaigns"
)

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        Campaign,
		Description: "Create or update a campaign to track.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "title",
				Description: "The title of your campaign.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "referee",
				Description: "The Referee of the campaign.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionRole,
				Name:        "player-role",
				Description: "The role assigned to players in the campaign.",
				Required:    false,
			},
		},
	},
	{
		Name:        ListCampaigns,
		Description: "List all campaigns.",
	},
}

var CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	Campaign: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		options := i.Interaction.ApplicationCommandData().Options

		optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
		for _, opt := range options {
			optionMap[opt.Name] = opt
		}

		title := optionMap["title"].StringValue()

		var referee *discordgo.User
		var refereeID string
		if optionMap["referee"] != nil {
			referee = optionMap["referee"].UserValue(s)
			refereeID = referee.ID
		}

		var playerRole *discordgo.Role
		var playerRoleID string
		if optionMap["player-role"] != nil {
			playerRole = optionMap["player-role"].RoleValue(s, i.GuildID)
			playerRoleID = playerRole.ID
		}

		campaign := sqlite.Campaign{
			Title:        title,
			RefereeID:    refereeID,
			PlayerRoleID: playerRoleID,
		}
		campaign, err := sqlite.UpdateCampaign(campaign)
		var res string
		if err != nil {
			res = "Error creating campaign"
		} else {
			res = fmt.Sprintf("Created campaign %s!", title)

			if referee, err = s.User(campaign.RefereeID); referee != nil && err == nil {
				res = fmt.Sprintf("%s It is run by %s.", res, referee.Mention())
			}
			if playerRole, err = s.State.Role(i.GuildID, campaign.PlayerRoleID); playerRole != nil && err == nil {
				res = fmt.Sprintf("%s Players in it have the %s role.", res, playerRole.Mention())
			}
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: res,
			},
		})
	},
	ListCampaigns: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		campaigns, err := sqlite.GetAllCampaigns()
		if err != nil {
			log.Printf("Error getting campaigns: %v", err)
		}

		var res string
		for _, campaign := range campaigns {
			res += fmt.Sprintf("%d. %s\n", campaign.Id, campaign.Title)
			if campaign.RefereeID != "" {
				referee, err := s.User(campaign.RefereeID)
				if err != nil {
					log.Printf("Error getting referee: %v", err)
				} else {
					res += fmt.Sprintf("    - Referee: %s\n", referee.Mention())
				}
			}
			if campaign.PlayerRoleID != "" {
				playerRole, err := s.State.Role(i.GuildID, campaign.PlayerRoleID)
				if err != nil {
					log.Printf("Error getting player role: %v", err)
				} else {
					res += fmt.Sprintf("    - Player Role: %s\n", playerRole.Mention())
				}
			}
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: res,
			},
		})
	},
}

func RegisterCommands(s *discordgo.Session, guildID string, debug bool) []*discordgo.ApplicationCommand {
	registeredCommands := make([]*discordgo.ApplicationCommand, len(Commands))

	for i, v := range Commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	if debug {
		registeredCommands = RegisterDebugCommands(guildID, s, registeredCommands)
	}

	return registeredCommands
}

func DeleteCommands(s *discordgo.Session, guildID string, registeredCommands []*discordgo.ApplicationCommand) error {
	log.Println("Deleting commands...")
	for _, cmd := range registeredCommands {
		err := s.ApplicationCommandDelete(s.State.User.ID, guildID, cmd.ID)
		if err != nil {
			log.Printf("Cannot delete command %s: ", cmd.Name)
		}
	}
	return nil
}

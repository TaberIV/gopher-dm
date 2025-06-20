// Debug Commands
package bot

import (
	"errors"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

const (
	DebugClear string = "clear"
)

var DebugCommands map[string]*discordgo.ApplicationCommand = map[string]*discordgo.ApplicationCommand{
	DebugClear: {
		Name:        DebugClear,
		Description: "Delete all messages in a channel (DEBUG USE ONLY CANNOT BE UNDONE!)",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "channel",
				Description: "Channel to clear (defaults to current channel).",
				Type:        discordgo.ApplicationCommandOptionChannel,
				Required:    false,
			},
		},
	},
}

var DebugHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	DebugClear: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		var channelId string
		if len(i.Interaction.ApplicationCommandData().Options) != 0 {
			channelId = i.Interaction.ApplicationCommandData().Options[0].ChannelValue(s).ID
		} else {
			channelId = i.Interaction.ChannelID
		}

		messages, err := s.ChannelMessages(channelId, 100, "", "", "")
		if err != nil {
			log.Println("Cannot retrieve messages: ", err)
		}

		messageIds := make([]string, len(messages))
		for i, v := range messages {
			messageIds[i] = v.ID
		}
		err = s.ChannelMessagesBulkDelete(channelId, messageIds)

		var content string
		if err != nil {
			var RESTerror *discordgo.RESTError
			if errors.As(err, RESTerror) && RESTerror.Message.Code == 50034 {
				content = "Cannot delete messages older than 14 days. Limit your request to delete more recent messages."
			} else {
				content = fmt.Sprintf("Unknown error deleting messages: %v", err)
			}
		} else {
			content = fmt.Sprintf("Deleted %d messages", len(messageIds))
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: content,
			},
		})
	},
}

func RegisterDebugCommands(guildID string, s *discordgo.Session, registeredCommands []*discordgo.ApplicationCommand) []*discordgo.ApplicationCommand {
	debugCommands := make([]*discordgo.ApplicationCommand, len(DebugCommands))

	i := 0
	for _, v := range DebugCommands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, v)
		if err != nil {
			log.Printf("Cannot create '%v' debug command: %v", v.Name, err)
		}
		debugCommands[i] = cmd
		i++
	}

	return append(registeredCommands, debugCommands...)
}

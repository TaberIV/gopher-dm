package bot

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/taberiv/gopher-dm/sqlite"
)

var BotToken, AppID string
var GuildID = "659809913194807306"

func checkNilErr(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func Run() {

	// Create session
	discord, err := discordgo.New("Bot " + BotToken)
	checkNilErr(err)

	// Add Command Handler
	discord.AddHandler(onInteractionCreate)

	// Start Discord Session
	err = discord.Open()
	checkNilErr(err)

	log.Println("Gopher DM is starting a session...")
	defer discord.Close()

	// Register Commands
	user, err := discord.User("@me")
	checkNilErr(err)
	AppID = user.ID
	_, err = discord.ApplicationCommandCreate(AppID, GuildID, Commands["campaign"])
	checkNilErr(err)

	// Close on OS Interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println("Scheduling, no!")
}

func onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case Campaign:
			title := i.Interaction.ApplicationCommandData().Options[0]
			log.Printf("Create campaign: %s\n", title.StringValue())
			err := sqlite.CreateCampaign(title.StringValue())
			var res string
			if err != nil {
				res = "Failed. Campaign already exists!"
			} else {
				res = "Created campaign " + title.StringValue() + "!"
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: res,
				},
			})
		}
	}
}

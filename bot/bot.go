package bot

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

func checkNilErrMsg(e error, msg string) {
	if e != nil {
		if msg != "" {
			log.Fatalf("%s: %v", msg, e)
		} else {
			log.Fatal(e)
		}
	}
}

func Run(botToken, guildID string, debug bool) {
	// Create session
	discord, err := discordgo.New("Bot " + botToken)
	checkNilErrMsg(err, "Error creating bot, check token")

	// Add Command Handler
	discord.AddHandler(onInteractionCreate)

	// Start Discord Session
	err = discord.Open()
	checkNilErrMsg(err, "Error opening session")

	log.Println("Gopher DM is starting the session...")
	defer discord.Close()

	log.Println("Registering commands...")
	registeredCommands := RegisterCommands(discord, guildID, debug)
	if debug {
		defer DeleteCommands(discord, guildID, registeredCommands)
	}

	log.Println("The adventure begins!")

	// Close on OS Interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println("Ending session...")
}

func onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if h, ok := CommandHandlers[i.ApplicationCommandData().Name]; ok {
		h(s, i)
	} else if true {
		if h, ok := DebugHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	}
}

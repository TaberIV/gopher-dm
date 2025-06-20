package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	bot "github.com/taberiv/gopher-dm/bot"
	"github.com/taberiv/gopher-dm/sqlite"
	_ "github.com/taberiv/gopher-dm/sqlite"
)

func main() {
	// Load environment variables
	err := godotenv.Load("local.env")
	if err != nil {
		log.Fatal(err)
	}

	// Process command line arguments
	argMap, err := makeArgsMap(os.Args)
	if err != nil {
		log.Print(err)
	}

	debug, err := strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		log.Print("Error parsing DEBUG environment variable")
	}
	if debugStr, ok := argMap["debug"]; ok {
		debugArg, err := strconv.ParseBool(debugStr)
		if err != nil {
			log.Printf("Invalid value for debug: %s", debugStr)
		} else {
			debug = debugArg
		}
	}
	if debug {
		log.Printf("Starting in Debug mode...")
	}

	guildID := argMap["guild"]
	if guildID == "" && debug { // We don't want to set a GuildID in production
		guildID = os.Getenv("GUILD_ID")
	}

	token := argMap["token"]
	if token == "" {
		token = os.Getenv("BOT_TOKEN")
		if token == "" {
			log.Fatal("No token provided. Please provide a token using the --token flag or set the BOT_TOKEN environment variable.")
		}
	}

	// Initialize Database
	db, err := sqlite.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Start Discord Bot
	bot.Run(token, guildID, debug)
}

func makeArgsMap(args []string) (map[string]string, error) {
	ret := make(map[string]string)

	for i := 0; i < len(args); i++ {
		var name string
		if args[i][:2] == "--" {
			name = args[i][2:]
		} else if args[i] == "-d" {
			name = "debug"
		}

		if name != "" {
			ret[name] = args[i+1]
			i++
		}
	}

	return ret, nil
}

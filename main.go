package main

import (
	"log"
	"os"

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

	// Initialize Database
	db, err := sqlite.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Start Discord Bot
	bot.BotToken = os.Getenv("BOT_TOKEN")
	bot.Run()
}

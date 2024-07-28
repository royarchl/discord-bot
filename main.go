package main

import (
	"log"
	"os"

	bot "example.com/discord-bot/bot"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(".env file couldn't be loaded")
	}

	bot.BotToken = os.Getenv("BOT_TOKEN")
	bot.Run()
}

package bot

import (
	"log"
	"os"
	"os/signal"

	"example.com/discord-bot/handlers/commands"
	"example.com/discord-bot/handlers/errors"
	"example.com/discord-bot/handlers/settings"
	"example.com/discord-bot/handlers/text"
	"example.com/discord-bot/handlers/voice"

	"github.com/bwmarrin/discordgo"
)

var (
	// Global Variables
	BotToken string
)

func Run() {
	// Open database connection
	log.Println("Opening database connection...")
	db, err := settings.InitDatabase()
	errors.CheckNilErr(err)
	defer db.Close()

	// create a session
	log.Println("Establishing connection to Discord...")
	session, err := discordgo.New("Bot " + BotToken)
	errors.CheckNilErr(err)

	// add event handlers
	log.Println("Adding handlers...")
	// Triggers when bot joins or reconnects to a guild
	session.AddHandler(commands.OnGuildJoin)
	session.AddHandler(text.NewMessage)
	session.AddHandler(voice.VoiceStateUpdate)
	session.AddHandler(commands.SetCommands)
	// Triggers when bot leaves or disconnects from a guild
	session.AddHandler(commands.OnGuildLeave)

	err = session.Open()
	errors.CheckNilErr(err)
	defer session.Close()

	// keep bot running while no OS interruption
	log.Println("Bot running... (Press Ctrl+C to exit)")
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)
	<-channel

	log.Println("Removing commands...")
	for _, guild := range session.State.Guilds {
		log.Printf("[GLD] %v", guild.Name)
		commands.RemoveCommands(session, guild.ID)
	}

	log.Println("Terminating connection to Discord...")
	log.Println("Closing database connection...")
	log.Println("Gracefully shutting down.")
}

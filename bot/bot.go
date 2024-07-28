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
	BotToken string
)

func Run() {
	db, err := settings.InitDatabase()
	errors.CheckNilErr(err)
	defer db.Close()

	session, err := discordgo.New("Bot " + BotToken)
	errors.CheckNilErr(err)

	session.AddHandler(commands.OnGuildJoin)
	session.AddHandler(text.NewMessage)
	session.AddHandler(voice.VoiceStateUpdate)
	session.AddHandler(commands.SetCommands)
	session.AddHandler(commands.OnGuildLeave)

	err = session.Open()
	errors.CheckNilErr(err)
	defer session.Close()

	// keep bot running while no OS interruption
	log.Println("Bot running... (Press Ctrl+C to exit)")
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)
	<-channel

  log.Println("Shutting down gracefully")

	for _, guild := range session.State.Guilds {
		commands.RemoveCommands(session, guild.ID)
	}
}

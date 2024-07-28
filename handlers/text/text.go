package text

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func NewMessage(discord *discordgo.Session, message *discordgo.MessageCreate) {
	/* prevent bot from responding to its own message. this is achieved by looking into the message author id. if message.author.id is same as bot.author.id then just return */
	if message.Author.ID == discord.State.User.ID {
		return
	}

	// respond to user message if it contains `!help` or `!bye`
	switch {
	case strings.Contains(message.Content, "!help"):
		discord.ChannelMessageSend(message.ChannelID, "Hello world!")
	case strings.Contains(message.Content, "!bye"):
		discord.ChannelMessageSend(message.ChannelID, "Goodbye!")
	}
}

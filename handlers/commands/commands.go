package commands

import (
	"fmt"
	"log"
	"time"

	"example.com/discord-bot/handlers/settings"

	"github.com/bwmarrin/discordgo"
)

var (
	dmPermission                   = false
	defaultMemberPermissions int64 = discordgo.PermissionManageServer

	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "about",
			Description: "A link to the source code of the application",
			Type:        discordgo.ChatApplicationCommand,
		},
		{
			Name:        "ping",
			Description: "Check the bot's latency",
			Type:        discordgo.ChatApplicationCommand,
		},
		{
			Name:                     "set",
			Description:              "Managing application (bot) settings",
			DefaultMemberPermissions: &defaultMemberPermissions,
			DMPermission:             &dmPermission,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "channel",
					Description: "Modify generic channel creation settings",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "channel",
							Description: "The voice channel to set",
							Type:        discordgo.ApplicationCommandOptionChannel,
							Required:    true,
							ChannelTypes: []discordgo.ChannelType{
								discordgo.ChannelTypeGuildVoice,
							},
						},
						{
							Name:        "category",
							Description: "The category to create the channels in",
							Type:        discordgo.ApplicationCommandOptionChannel,
							Required:    false,
							ChannelTypes: []discordgo.ChannelType{
								discordgo.ChannelTypeGuildCategory,
							},
						},
						{
							Name:        "template",
							Description: "Set the base name of generated VCs. Default: `VC`",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    false,
						},
					},
				},
				{
					Name:        "activation",
					Description: "Enable/disable vc-creating functionality",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "bool",
							Description: "True or false",
							Type:        discordgo.ApplicationCommandOptionBoolean,
							Required:    true,
						},
					},
				},
				{
					Name:        "remove-on-offline",
					Description: "Remove bot commands when it goes offline",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "bool",
							Description: "True or false",
							Type:        discordgo.ApplicationCommandOptionBoolean,
							Required:    true,
						},
					},
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"about": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "https://github.com/royarchl/",
				},
			})
		},
		"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			start := time.Now()

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Pinging...",
				},
			})

			latency := time.Since(start).Milliseconds()

			responseContent := fmt.Sprintf("Latency is %dms.", latency)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &responseContent,
			})
		},
		"set": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			subCommand := i.ApplicationCommandData().Options[0]
			content := ""

			currentSetting := settings.QueryGuildSetting(i.GuildID)
			if currentSetting == nil {
				currentSetting = &settings.Setting{
					GuildID:           i.GuildID,
					VoiceTemplateName: "VC",
				}
			}

			switch subCommand.Name {
			case "channel":
				// currentSetting := settings.QueryGuildSetting(i.GuildID)

				for _, option := range subCommand.Options {

					// settings.ModifyGuildSetting(i.GuildID, )

					switch option.Name {
					case "channel":
						// currentSetting.VoiceID = option.ChannelValue(s).ID
						// if currentSetting.CategoryID == "" {
						// 	currentSetting.CategoryID = option.ChannelValue(s).ParentID
						// }
						channel := option.ChannelValue(s)

						settings.ModifyGuildSetting(i.GuildID, settings.WithVoiceID(channel.ID))
						if settings.QueryGuildSetting(i.GuildID).CategoryID == "" {
							settings.ModifyGuildSetting(i.GuildID, settings.WithCategoryID(channel.ParentID))
						}
					case "category":
						// currentSetting.CategoryID = option.ChannelValue(s).ID
						category := option.ChannelValue(s)

						settings.ModifyGuildSetting(i.GuildID, settings.WithCategoryID(category.ID))
					case "template":
						// currentSetting.VoiceTemplateName = option.StringValue()
						template := option.StringValue()

						settings.ModifyGuildSetting(i.GuildID, settings.WithVoiceTemplateName(template))
					}
				}

				content = fmt.Sprintf(
					"Channel: <#%v>\n"+
						"Category: <#%v>\n"+
						"Template: `%v`",
					currentSetting.VoiceID, currentSetting.CategoryID, currentSetting.VoiceTemplateName,
				)

				// err := settings.UpsertGuildSetting(
				// 	currentSetting.GuildID,
				// 	true, // RemoveCommands
				// 	currentSetting.VoiceID,
				// 	currentSetting.CategoryID,
				// 	currentSetting.VoiceTemplateName,
				// 	true, // IsEnabled
				// )
				// if err != nil {
				// 	content = "Failed to update settings: " + err.Error()
				// } else {
				// 	content = fmt.Sprintf(
				// 		"Channel: <#%v>\n"+
				// 			"Category: <#%v>\n"+
				// 			"Template: `%v`",
				// 		currentSetting.VoiceID, currentSetting.CategoryID, currentSetting.VoiceTemplateName,
				// 	)
				// }

				log.Println("Updating guild settings cache...")
				settings.PrintCache()
			case "activation":
				activationValue := subCommand.Options[0].BoolValue()

				settings.ModifyGuildSetting(i.GuildID, settings.WithIsEnabled(activationValue))
				content = fmt.Sprintf("activation: %v", activationValue)

				// currentSetting := settings.QueryGuildSetting(i.GuildID)
				// if currentSetting != nil {
				// 	settings.ModifyGuildSetting(i.GuildID, settings.WithIsEnabled(activationValue))
				// 	content = "Activation successfully toggled."
				// } else {
				// 	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				// 		Type: discordgo.InteractionResponseChannelMessageWithSource,
				// 		Data: &discordgo.InteractionResponseData{
				// 			Content: "### Error! `❌`\n >>> There is no channel set yet. Use `/set channel`.",
				// 			Flags:   discordgo.MessageFlagsEphemeral, // "Only you can see this"
				// 		},
				// 	})
				// 	return
				// }
			case "remove-on-offline":
				activationValue := subCommand.Options[0].BoolValue()
				log.Printf("remove_on_offline: %v", activationValue)

				settings.ModifyGuildSetting(i.GuildID, settings.WithRemoveCommands(activationValue))
				content = fmt.Sprintf("remove_on_offline: %v", activationValue)
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf(
						"### Settings updated successfully. `✅`\n"+
							">>> %v",
						content,
					),
					Flags: discordgo.MessageFlagsEphemeral, // "Only you can see this"
				},
			})
		},
	}
)

func SetCommands(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
		h(s, i)
	}
}

// Might require another a closure / variatic function
func RegisterCommands(s *discordgo.Session, guildID string) {
	// ERROR:
	// Doesn't check for new subcommands

	existingCommands, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		log.Printf("Cannot fetch commands for guild '%v': %v", guildID, err)
	}

	existingCommandNames := make(map[string]bool)
	for _, cmd := range existingCommands {
		existingCommandNames[cmd.Name] = true
	}

	for _, v := range commands {
		if _, exists := existingCommandNames[v.Name]; exists {
			// Command already exists
			log.Printf("[SKP] %v (%v)", v.Name, v.ID)
			continue
		}

		// Add command to guild
		log.Printf("[ADD] %v (%v)", v.Name, v.ID)
		_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
	}
}

func RemoveCommands(s *discordgo.Session, guildID string) {
	if !settings.QueryGuildSetting(guildID).RemoveCommands {
		log.Printf("[SKP] RemoveCommands set to FALSE")
		return
	}

	registeredCommands, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		log.Panicf("Cannot fetch commands for guild '%v': %v", guildID, err)
	}

	for _, cmd := range registeredCommands {
		err := s.ApplicationCommandDelete(s.State.User.ID, guildID, cmd.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", cmd.Name, err)
		} else {
			log.Printf("[DEL] %v (%v)", cmd.Name, cmd.ID)
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func OnGuildJoin(s *discordgo.Session, g *discordgo.GuildCreate) {
	RegisterCommands(s, g.ID)
}
func OnGuildLeave(s *discordgo.Session, g *discordgo.GuildDelete) {
	RemoveCommands(s, g.ID)
}
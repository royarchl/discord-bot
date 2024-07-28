package voice

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"

	"example.com/discord-bot/handlers/errors"
	"example.com/discord-bot/handlers/settings"

	"github.com/bwmarrin/discordgo"
)

var (
	voiceChannelUsers = make(map[string]int)
	mu                sync.RWMutex
	// FunctionalityEnabled = false
)

func VoiceStateUpdate(session *discordgo.Session, voiceState *discordgo.VoiceStateUpdate) {
	// Allows concurrency (threads)
	mu.Lock()
	defer mu.Unlock()

	// if !FunctionalityEnabled {
	// 	return
	// }

	guildSettings := settings.QueryGuildSetting(voiceState.GuildID)
	if guildSettings == nil || !guildSettings.IsEnabled {
		return
	}

	// Modify these functions, and all implementations, to not need the guildSettings argument
	if voiceState.BeforeUpdate == nil {
		handleUserJoin(session, voiceState, guildSettings)
	} else if voiceState.ChannelID == "" {
		handleUserLeave(session, voiceState, guildSettings)
	} else {
		handleUserSwitch(session, voiceState, guildSettings)
	}
}

func handleUserJoin(session *discordgo.Session, voiceState *discordgo.VoiceStateUpdate, settings *settings.Setting) {
	if voiceState.ChannelID == settings.VoiceID {
		createVoiceChannelAndMoveUser(session, voiceState, settings)
	} else {
		incrementUserCount(voiceState.ChannelID)
	}
}

func handleUserLeave(session *discordgo.Session, voiceState *discordgo.VoiceStateUpdate, settings *settings.Setting) {
	if voiceState.BeforeUpdate.ChannelID == settings.VoiceID {
		return
	}
	decrementUserCount(session, voiceState.BeforeUpdate.ChannelID)
}

func handleUserSwitch(session *discordgo.Session, voiceState *discordgo.VoiceStateUpdate, settings *settings.Setting) {
	if voiceState.ChannelID == settings.VoiceID {
		decrementUserCount(session, voiceState.BeforeUpdate.ChannelID)
		createVoiceChannelAndMoveUser(session, voiceState, settings)
	} else {
		incrementUserCount(voiceState.ChannelID)
		decrementUserCount(session, voiceState.BeforeUpdate.ChannelID)
	}
}

func incrementUserCount(channelID string) error {
	if _, ok := voiceChannelUsers[channelID]; !ok {
		return fmt.Errorf("channel %s does not exist in the map", channelID)
	}
	voiceChannelUsers[channelID]++
	return nil
}

func decrementUserCount(session *discordgo.Session, channelID string) error {
	if _, ok := voiceChannelUsers[channelID]; !ok {
		return fmt.Errorf("channel %s does not exist in the map", channelID)
	}

	voiceChannelUsers[channelID]--
	if voiceChannelUsers[channelID] < 1 {
		_, err := session.ChannelDelete(channelID)
		errors.CheckNilErr(err)
		delete(voiceChannelUsers, channelID)
	}
	return nil
}

func createVoiceChannelAndMoveUser(session *discordgo.Session, voiceState *discordgo.VoiceStateUpdate, settings *settings.Setting) {
	channelData := &discordgo.GuildChannelCreateData{
		Name:     fmt.Sprintf("%v #%v", settings.VoiceTemplateName, strconv.Itoa(rand.Intn(100))),
		Type:     2, // ChannelTypeGuildVoice
		ParentID: settings.CategoryID,
	}

	newChannel, err := session.GuildChannelCreateComplex(voiceState.GuildID, *channelData)
	errors.CheckNilErr(err)

	voiceChannelUsers[newChannel.ID] = 0

	err = session.GuildMemberMove(voiceState.GuildID, voiceState.UserID, &newChannel.ID)
	errors.CheckNilErr(err)
}

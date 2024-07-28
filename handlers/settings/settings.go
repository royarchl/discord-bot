package settings

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"

	"example.com/discord-bot/handlers/errors"
	_ "github.com/mattn/go-sqlite3"
)

const dbPath = "./settings.db"

type Setting struct {
	GuildID           string
	RemoveCommands    bool
	VoiceID           string
	CategoryID        string
	VoiceTemplateName string
	IsEnabled         bool
}

var (
	settingsCache map[string]*Setting
	Database      *sql.DB
	cacheMutex    sync.RWMutex
)

func InitDatabase() (*sql.DB, error) {
	db, err := openDatabase()
	errors.CheckNilErr(err)

	err = createTable(db)
	errors.CheckNilErr(err)

	log.Println("Caching all guild settings...")
	cacheAllGuildSettings(db)

	Database = db
	return db, err
}

func openDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func createTable(db *sql.DB) error {
	createTableSQL := `CREATE TABLE IF NOT EXISTS settings (
		"guild_id" TEXT NOT NULL PRIMARY KEY,
		"remove_commands" BOOLEAN,
		"voice_id" TEXT,
		"category_id" TEXT,
		"voice_template_name" TEXT,
		"is_enabled" BOOLEAN
	);`
	_, err := db.Exec(createTableSQL)
	return err
}

func cacheAllGuildSettings(db *sql.DB) {
	settingsCache = make(map[string]*Setting)

	rows, err := db.Query("SELECT guild_id, remove_commands, voice_id, category_id, voice_template_name, is_enabled FROM settings")
	errors.CheckNilErr(err)
	defer rows.Close()

	for rows.Next() {
		var setting Setting
		// maps db values to the provided variables
		err = rows.Scan(&setting.GuildID, &setting.RemoveCommands, &setting.VoiceID, &setting.CategoryID, &setting.VoiceTemplateName, &setting.IsEnabled)
		errors.CheckNilErr(err)
		settingsCache[setting.GuildID] = &setting
	}

	PrintCache()
}

// DEPRECATED!
func upsertGuildSetting(guildID string, removeCommands bool, voiceID, categoryID, voiceTemplateName string, isEnabled bool) error {
	// I would just like to state that I AM INTENTIONALLY OVERWRITING THE WHOLE STRUCT because going through the hassle of checking
	// for a change in every property without being able to overload struct comparators like in C++ is not worth the headache for
	// a minimal improvement in performance for my already shoddy code. Thank you but no.

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if _, exists := settingsCache[guildID]; exists {
		// update
		updateSQL := `UPDATE settings SET remove_commands = ?, voice_id = ?, category_id = ?, voice_template_name = ?, is_enabled = ? WHERE guild_id = ?`
		_, err := Database.Exec(updateSQL, removeCommands, voiceID, categoryID, voiceTemplateName, isEnabled, guildID)
		if err != nil {
			return err
		}
	} else {
		// insert
		insertSQL := `INSERT INTO settings (guild_id, remove_commands, voice_id, category_id, voice_template_name, is_enabled) VALUES (?, ?, ?, ?, ?, ?)`
		_, err := Database.Exec(insertSQL, guildID, removeCommands, voiceID, categoryID, voiceTemplateName, isEnabled)
		if err != nil {
			return err
		}
	}

	settingsCache[guildID] = &Setting{
		GuildID:           guildID,
		RemoveCommands:    removeCommands,
		VoiceID:           voiceID,
		CategoryID:        categoryID,
		VoiceTemplateName: voiceTemplateName,
		IsEnabled:         isEnabled,
	}

	return nil
}

// Closure time babyyy
type SettingOption func(*Setting) error

func ModifyGuildSetting(guildID string, options ...SettingOption) (*Setting, error) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	setting, exists := settingsCache[guildID]
	if !exists {
		setting = &Setting{GuildID: guildID}
	}

	for _, option := range options {
		if err := option(setting); err != nil {
			return nil, err
		}
	}

	if exists {
		// Update
		updateFields := []string{}
		updateValues := []interface{}{}

		if setting.VoiceID != "" {
			updateFields = append(updateFields, "voice_id = ?")
			updateValues = append(updateValues, setting.VoiceID)
		}
		if setting.CategoryID != "" {
			updateFields = append(updateFields, "category_id = ?")
			updateValues = append(updateValues, setting.CategoryID)
		}
		if setting.VoiceTemplateName != "" {
			updateFields = append(updateFields, "voice_template_name = ?")
			updateValues = append(updateValues, setting.VoiceTemplateName)
		}
		updateFields = append(updateFields, "remove_commands = ?", "is_enabled = ?")
		updateValues = append(updateValues, setting.RemoveCommands, setting.IsEnabled)

		updateValues = append(updateValues, guildID)

		updateSQL := fmt.Sprintf("UPDATE settings SET %s WHERE guild_id = ?", strings.Join(updateFields, ", "))
		_, err := Database.Exec(updateSQL, updateValues...)
		if err != nil {
			return nil, err
		}
	} else {
		// Insert
		insertSQL := `INSERT INTO settings (guild_id, remove_commands, voice_id, category_id, voice_template_name, is_enabled) VALUES (?, ?, ?, ?, ?, ?)`
		_, err := Database.Exec(insertSQL, guildID, setting.RemoveCommands, setting.VoiceID,
			setting.CategoryID, setting.VoiceTemplateName, setting.IsEnabled)
		if err != nil {
			return nil, err
		}
	}

	settingsCache[guildID] = setting

	return setting, nil
}

func WithRemoveCommands(removeCommands bool) SettingOption {
	return func(s *Setting) error {
		s.RemoveCommands = removeCommands
		return nil
	}
}

func WithVoiceID(voiceID string) SettingOption {
	return func(s *Setting) error {
		s.VoiceID = voiceID
		return nil
	}
}

func WithCategoryID(categoryID string) SettingOption {
	return func(s *Setting) error {
		s.CategoryID = categoryID
		return nil
	}
}

func WithVoiceTemplateName(voiceTemplateName string) SettingOption {
	return func(s *Setting) error {
		s.VoiceTemplateName = voiceTemplateName
		return nil
	}
}

func WithIsEnabled(isEnabled bool) SettingOption {
	return func(s *Setting) error {
		s.IsEnabled = isEnabled
		return nil
	}
}

func QueryGuildSetting(guildID string) *Setting {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	return settingsCache[guildID]
}

func PrintCache() {
	log.Printf("[CCH] %v", settingsCache)
}

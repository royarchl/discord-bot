# Discord bot in Go
A personal project to recreate the functionality of creating temporary VCs, which get automatically deleted when no users remain in the call.

### Features
- Slash commands
- Ability to create temporary voice channels
- Settings being stored on a per-guild basis
- Persistent storage of settings between sessions (bot going offline)

> [!NOTE]
> Go to the **settings.go** file and read through the ModifyGuildSettings() function. Look at that sick use of closures and variatic functions. You can see how those functions are used in the **commands.go** file.

### Dependencies
- [discordgo](https://github.com/bwmarrin/discordgo)
- [godotenv](https://github.com/joho/godotenv)
- [go-sqlite3](https://github.com/mattn/go-sqlite3)

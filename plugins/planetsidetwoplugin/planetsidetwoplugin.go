package planetsidetwoplugin

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/lampjaw/mutterblack.discord"
)

const CENSUS_IMAGEBASE_URI = "http://census.daybreakgames.com/files/ps2/images/static/"
const VOIDWELL_URI = "https://voidwell.com/"

type planetsidetwoPlugin struct {
	sync.RWMutex
}

type PlanetsideCharacter struct {
	CharacterId          string  `json:"id"`
	World                string  `json:"world"`
	Name                 string  `json:"name"`
	LastSaved            string  `json:"lastSaved"`
	FactionId            int     `json:"factionId"`
	FactionName          string  `json:"factionName"`
	FactionImageId       int     `json:"factionImageId"`
	BattleRank           int     `json:"battleRank"`
	OutfitAlias          string  `json:"outfitAlias"`
	OutfitName           string  `json:"outfitName"`
	Kills                int     `json:"kills"`
	Deaths               int     `json:"deaths"`
	PlayTime             int     `json:"playTime"`
	TotalPlayTimeMinutes int     `json:"totalPlayTimeMinutes"`
	Score                int     `json:"score"`
	KillDeathRatio       float32 `json:"killDeathRatio"`
	HeadshotRatio        float32 `json:"headshotRatio"`
	KillsPerHour         float32 `json:"killsPerHour"`
	TotalKillsPerHour    float32 `json:"totalKillsPerHour"`
	SiegeLevel           float32 `json:"siegeLevel"`
	IVIScore             int     `json:"iviScore"`
	IVIKillDeathRatio    float32 `json:"iviKillDeathRatio"`
	Prestige             int     `json:"prestige"`
}

type PlanetsideCharacterWeapon struct {
	CharacterId         string  `json:"characterId"`
	CharacterName       string  `json:"characterName"`
	ItemId              int     `json:"itemId"`
	WeaponName          string  `json:"weaponName"`
	WeaponImageId       int     `json:"weaponImageId"`
	Kills               int     `json:"kills"`
	Deaths              int     `json:"deaths"`
	PlayTime            int     `json:"playTime"`
	Score               int     `json:"score"`
	Headshots           int     `json:"headshots"`
	KillDeathRatio      float32 `json:"killDeathRatio"`
	HeadshotRatio       float32 `json:"headshotRatio"`
	KillsPerHour        float32 `json:"killsPerHour"`
	Accuracy            float32 `json:"accuracy"`
	KillDeathRatioGrade string  `json:"killDeathRatioGrade"`
	HeadshotRatioGrade  string  `json:"headshotRatioGrade"`
	KillsPerHourGrade   string  `json:"killsPerHourGrade"`
	AccuracyGrade       string  `json:"accuracyGrade"`
}

type PlanetsideOutfit struct {
	OutfitId       string `json:"outfitId"`
	Name           string `json:"name"`
	Alias          string `json:"alias"`
	FactionName    string `json:"factionName"`
	FactionImageId int    `json:"factionImageId"`
	WorldName      string `json:"worldName"`
	LeaderName     string `json:"leaderName"`
	MemberCount    int    `json:"memberCount"`
	Activity7Days  int    `json:"activity7Days"`
	Activity30Days int    `json:"activity30Days"`
	Activity90Days int    `json:"activity90Days"`
}

func (p *planetsidetwoPlugin) Commands() []mutterblack.CommandDefinition {
	return []mutterblack.CommandDefinition{
		mutterblack.CommandDefinition{
			CommandGroup: p.Name(),
			CommandID:    "ps2-character",
			Triggers: []string{
				"ps2c",
			},
			Arguments: []mutterblack.CommandDefinitionArgument{
				mutterblack.CommandDefinitionArgument{
					Pattern: "[a-zA-Z0-9]*",
					Alias:   "characterName",
				},
			},
			Description: "Get stats for a player.",
			Callback:    p.runCharacterStatsCommand,
		},
		mutterblack.CommandDefinition{
			CommandGroup: p.Name(),
			CommandID:    "ps2-character-weapons",
			Triggers: []string{
				"ps2c",
			},
			Arguments: []mutterblack.CommandDefinitionArgument{
				mutterblack.CommandDefinitionArgument{
					Pattern: "[a-zA-Z0-9]*",
					Alias:   "characterName",
				},
				mutterblack.CommandDefinitionArgument{
					Pattern: ".*",
					Alias:   "weaponName",
				},
			},
			Description: "Get weapon stats for a player.",
			Callback:    p.runCharacterWeaponStatsCommand,
		},
		mutterblack.CommandDefinition{
			CommandGroup: p.Name(),
			CommandID:    "ps2-outfit",
			Triggers: []string{
				"ps2o",
			},
			Arguments: []mutterblack.CommandDefinitionArgument{
				mutterblack.CommandDefinitionArgument{
					Pattern: "[a-zA-Z0-9]{2,4}",
					Alias:   "outfitAlias",
				},
			},
			Description: "Get outfit stats by outfit tag.",
			Callback:    p.runOutfitStatsCommand,
		},
	}
}

func (p *planetsidetwoPlugin) Name() string {
	return "PS2Stats"
}

func (p *planetsidetwoPlugin) Load(bot *mutterblack.Bot, client *mutterblack.Discord, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
		}
	}

	return nil
}

func (p *planetsidetwoPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

func (p *planetsidetwoPlugin) Help(bot *mutterblack.Bot, client *mutterblack.Discord, message mutterblack.Message, detailed bool) []string {
	return []string{
		mutterblack.CommandHelp(client, "ps2c", "<character name>", "Get stats for a player.")[0],
		mutterblack.CommandHelp(client, "ps2c", "<character name> <weapon name>", "Get weapon stats for a player.")[0],
		mutterblack.CommandHelp(client, "ps2o", "<outfit name>", "Get outfit stats")[0],
	}
}

func (p *planetsidetwoPlugin) Message(bot *mutterblack.Bot, client *mutterblack.Discord, message mutterblack.Message) {

}

func (p *planetsidetwoPlugin) Stats(bot *mutterblack.Bot, client *mutterblack.Discord, message mutterblack.Message) []string {
	return nil
}

func New() mutterblack.Plugin {
	return &planetsidetwoPlugin{}
}

func (p *planetsidetwoPlugin) runCharacterStatsCommand(bot *mutterblack.Bot, client *mutterblack.Discord, message mutterblack.Message, args map[string]string) {
	resp, err := mutterblack.SendCoreCommand("planetside2", "character", args)

	if err != nil {
		p.RLock()
		client.SendMessage(message.Channel(), fmt.Sprintf("%s", err))
		p.RUnlock()
		return
	}

	var character PlanetsideCharacter
	json.Unmarshal(resp, &character)

	lastSaved, _ := time.Parse(time.RFC3339, character.LastSaved)

	fields := []*discordgo.MessageEmbedField{
		&discordgo.MessageEmbedField{
			Name:   "Last Seen",
			Value:  fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d UTC", lastSaved.Year(), lastSaved.Month(), lastSaved.Day(), lastSaved.Hour(), lastSaved.Minute(), lastSaved.Second()),
			Inline: false,
		},
		&discordgo.MessageEmbedField{
			Name:   "Server",
			Value:  character.World,
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "Battle Rank",
			Value:  fmt.Sprintf("%d", character.BattleRank),
			Inline: false,
		},
		&discordgo.MessageEmbedField{
			Name:   "Kills",
			Value:  fmt.Sprintf("%d", character.Kills),
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "Play Time",
			Value:  fmt.Sprintf("%0.1f (%0.1f) Hours", float32(character.PlayTime)/3600.0, float32(character.TotalPlayTimeMinutes)/60.0),
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "KDR",
			Value:  fmt.Sprintf("%0.2f", character.KillDeathRatio),
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "HSR",
			Value:  fmt.Sprintf("%0.2f%%", character.HeadshotRatio*100),
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "KpH",
			Value:  fmt.Sprintf("%0.2f (%0.2f)", character.KillsPerHour, character.TotalKillsPerHour),
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "Siege Level",
			Value:  fmt.Sprintf("%0.1f", character.SiegeLevel),
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "IVI Score",
			Value:  fmt.Sprintf("%d", character.IVIScore),
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "IVI KDR",
			Value:  fmt.Sprintf("%0.2f", character.IVIKillDeathRatio),
			Inline: true,
		},
	}

	if len(character.OutfitName) > 0 {
		outfitValue := character.OutfitName
		if len(character.OutfitAlias) > 0 {
			outfitValue = "[" + character.OutfitAlias + "] " + character.OutfitName
		}

		outfitField := &discordgo.MessageEmbedField{
			Name:   "Outfit",
			Value:  outfitValue,
			Inline: true,
		}

		fields = insertSlice(fields, outfitField, 2)
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: character.Name,
		},
		Title: "Click here for full stats",
		URL:   VOIDWELL_URI + "ps2/player/" + character.CharacterId,
		Color: 0x070707,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: createCensusImageURI(character.FactionImageId),
		},
		Fields: fields,
	}

	p.RLock()
	client.SendEmbedMessage(message.Channel(), embed)
	p.RUnlock()
}

func (p *planetsidetwoPlugin) runCharacterWeaponStatsCommand(bot *mutterblack.Bot, client *mutterblack.Discord, message mutterblack.Message, args map[string]string) {
	resp, err := mutterblack.SendCoreCommand("planetside2", "character-weapon", args)

	if err != nil {
		p.RLock()
		client.SendMessage(message.Channel(), fmt.Sprintf("%s", err))
		p.RUnlock()
		return
	}

	var weapon PlanetsideCharacterWeapon
	json.Unmarshal(resp, &weapon)

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: weapon.CharacterName + " [" + weapon.WeaponName + "]",
		},
		Title: "Click here for full stats",
		URL:   VOIDWELL_URI + "ps2/player/" + weapon.CharacterId,
		Color: 0x070707,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: createCensusImageURI(weapon.WeaponImageId),
		},
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "Kills",
				Value:  fmt.Sprintf("%d", weapon.Kills),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Deaths",
				Value:  fmt.Sprintf("%d", weapon.Deaths),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Play Time",
				Value:  fmt.Sprintf("%d Minutes", weapon.PlayTime/60),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Score",
				Value:  fmt.Sprintf("%d", weapon.Score),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "KpH",
				Value:  fmt.Sprintf("%0.2f", weapon.KillsPerHour),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Δ",
				Value:  weapon.KillsPerHourGrade,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "KDR",
				Value:  fmt.Sprintf("%0.2f", weapon.KillDeathRatio),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Δ",
				Value:  weapon.KillDeathRatioGrade,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "HSR",
				Value:  fmt.Sprintf("%0.2f%%", weapon.HeadshotRatio*100),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Δ",
				Value:  weapon.HeadshotRatioGrade,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Accuracy",
				Value:  fmt.Sprintf("%0.2f%%", weapon.Accuracy*100),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Δ",
				Value:  weapon.AccuracyGrade,
				Inline: true,
			},
		},
	}

	p.RLock()
	client.SendEmbedMessage(message.Channel(), embed)
	p.RUnlock()
}

func (p *planetsidetwoPlugin) runOutfitStatsCommand(bot *mutterblack.Bot, client *mutterblack.Discord, message mutterblack.Message, args map[string]string) {
	resp, err := mutterblack.SendCoreCommand("planetside2", "outfit", args)

	if err != nil {
		p.RLock()
		client.SendMessage(message.Channel(), fmt.Sprintf("%s", err))
		p.RUnlock()
		return
	}

	var outfit PlanetsideOutfit
	json.Unmarshal(resp, &outfit)

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: "[" + outfit.Alias + "] " + outfit.Name,
		},
		Title: "Click here for full stats",
		URL:   VOIDWELL_URI + "ps2/outfit/" + outfit.OutfitId,
		Color: 0x070707,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: createCensusImageURI(outfit.FactionImageId),
		},
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "Server",
				Value:  outfit.WorldName,
				Inline: false,
			},
			&discordgo.MessageEmbedField{
				Name:   "Leader",
				Value:  outfit.LeaderName,
				Inline: false,
			},
			&discordgo.MessageEmbedField{
				Name:   "Member Count",
				Value:  fmt.Sprintf("%d", outfit.MemberCount),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Activity 7 Days",
				Value:  fmt.Sprintf("%d", outfit.Activity7Days),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Activity 30 Days",
				Value:  fmt.Sprintf("%d", outfit.Activity30Days),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Activity 90 Days",
				Value:  fmt.Sprintf("%d", outfit.Activity90Days),
				Inline: true,
			},
		},
	}

	p.RLock()
	client.SendEmbedMessage(message.Channel(), embed)
	p.RUnlock()
}

func createCensusImageURI(imageId int) string {
	return CENSUS_IMAGEBASE_URI + fmt.Sprintf("%v", imageId) + ".png"
}

func insertSlice(arr []*discordgo.MessageEmbedField, value *discordgo.MessageEmbedField, index int) []*discordgo.MessageEmbedField {
	return append(arr[:index], append([]*discordgo.MessageEmbedField{value}, arr[index:]...)...)
}

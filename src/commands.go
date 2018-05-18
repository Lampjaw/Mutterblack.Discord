package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

const COMMAND_PREFIX = "?"

var COMMANDS_CONFIG = []CommandConfig{
	CommandConfig{
		Name: "ps2-character",
		Triggers: []string{
			"ps2character",
			"ps2c",
		},
		Description:        COMMAND_PREFIX + "ps2c <characterName>",
		CommandGroup:       "planetside2",
		CommandGroupAction: "character",
		Arguments: []CommandConfigArgument{
			CommandConfigArgument{
				Pattern:   "[a-zA-Z0-9]*",
				CoreAlias: "characterName",
			},
		},
		Process: func(s *discordgo.Session, m *discordgo.MessageCreate, resp json.RawMessage) {
			var character PlanetsideCharacter
			json.Unmarshal(resp, &character)

			lastSaved, _ := time.Parse(time.RFC3339, character.LastSaved)

			embed := &discordgo.MessageEmbed{
				Author: &discordgo.MessageEmbedAuthor{
					Name: character.Name,
				},
				Title: "Click here for full stats",
				URL:   VOIDWELL_URI + "ps2/player/" + character.CharacterId,
				Color: 0x070707,
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: createCensusImageUri(character.FactionImageId),
				},
				Fields: []*discordgo.MessageEmbedField{
					&discordgo.MessageEmbedField{
						Name:   "Last Seen",
						Value:  fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d UTC", lastSaved.Year(), lastSaved.Month(), lastSaved.Day(), lastSaved.Hour(), lastSaved.Minute(), lastSaved.Second()),
						Inline: false,
					},
					&discordgo.MessageEmbedField{
						Name:   "Server",
						Value:  character.World,
						Inline: false,
					},
					&discordgo.MessageEmbedField{
						Name:   "Battle Rank",
						Value:  fmt.Sprintf("%d", character.BattleRank),
						Inline: false,
					},
					&discordgo.MessageEmbedField{
						Name:   "Outfit",
						Value:  "[" + character.OutfitAlias + "] " + character.OutfitName,
						Inline: false,
					},
					&discordgo.MessageEmbedField{
						Name:   "Play Time",
						Value:  fmt.Sprintf("%d Hours", character.PlayTime/3600),
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
						Value:  fmt.Sprintf("%0.2f", character.KillsPerHour),
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
				},
			}

			s.ChannelMessageSendEmbed(m.ChannelID, embed)
		},
	},
	CommandConfig{
		Name: "ps2-character-weapon",
		Triggers: []string{
			"ps2character",
			"ps2c",
		},
		Description:        COMMAND_PREFIX + "ps2c <characterName> <weaponName>",
		CommandGroup:       "planetside2",
		CommandGroupAction: "character-weapon",
		Arguments: []CommandConfigArgument{
			CommandConfigArgument{
				Pattern:   "[a-zA-Z0-9]*",
				CoreAlias: "characterName",
			},
			CommandConfigArgument{
				Pattern:   ".*",
				CoreAlias: "weaponName",
			},
		},
		Process: func(s *discordgo.Session, m *discordgo.MessageCreate, resp json.RawMessage) {
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
					URL: createCensusImageUri(weapon.WeaponImageId),
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

			s.ChannelMessageSendEmbed(m.ChannelID, embed)
		},
	},
	CommandConfig{
		Name: "ps2-outfit",
		Triggers: []string{
			"ps2outfit",
			"ps2o",
		},
		Description:        COMMAND_PREFIX + "ps2o <outfitAlias>",
		CommandGroup:       "planetside2",
		CommandGroupAction: "outfit",
		Arguments: []CommandConfigArgument{
			CommandConfigArgument{
				Pattern:   "[a-zA-Z0-9]{2,4}",
				CoreAlias: "outfitAlias",
			},
		},
		Process: func(s *discordgo.Session, m *discordgo.MessageCreate, resp json.RawMessage) {
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
					URL: createCensusImageUri(outfit.FactionImageId),
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

			s.ChannelMessageSendEmbed(m.ChannelID, embed)
		},
	},
	CommandConfig{
		Name: "weather-forecast",
		Triggers: []string{
			"weather",
			"w",
		},
		Description:        COMMAND_PREFIX + "w <location>",
		CommandGroup:       "weather",
		CommandGroupAction: "forecast",
		Arguments: []CommandConfigArgument{
			CommandConfigArgument{
				Pattern:   ".*",
				CoreAlias: "location",
			},
		},
		Process: func(s *discordgo.Session, m *discordgo.MessageCreate, resp json.RawMessage) {
			var weather CurrentWeather
			json.Unmarshal(resp, &weather)

			var messageFields []*discordgo.MessageEmbedField
			for i := 0; i < 5; i++ {
				var field = &discordgo.MessageEmbedField{
					Name:   weather.Forecast[i].Date,
					Value:  createWeatherDay(weather.Forecast[i]),
					Inline: false,
				}
				messageFields = append(messageFields, field)
			}

			embed := &discordgo.MessageEmbed{
				Author: &discordgo.MessageEmbedAuthor{
					Name: weather.City + ", " + weather.Region + " " + weather.Country,
				},
				Color:  0x070707,
				Fields: messageFields,
			}

			s.ChannelMessageSendEmbed(m.ChannelID, embed)
		},
	},
}

func createCensusImageUri(imageId int) string {
	return CENSUS_IMAGEBASE_URI + fmt.Sprintf("%v", imageId) + ".png"
}

func createWeatherDay(d WeatherDay) string {
	var highTempCelsius = (float32(d.High) - 32) / 1.8
	var lowTempCelsius = (float32(d.Low) - 32) / 1.8

	var temperatureHigh = fmt.Sprintf("%d °F (%d °C)", d.High, int32(highTempCelsius))
	var temperatureLow = fmt.Sprintf("%d °F (%d °C)", d.Low, int32(lowTempCelsius))
	return fmt.Sprintf("%s: %s / %s - %s", d.Day, temperatureHigh, temperatureLow, d.Condition)
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

const MUTTERBLACK_CORE_URI = "http://mutterblack:5000/"
const CENSUS_IMAGEBASE_URI = "http://census.daybreakgames.com/files/ps2/images/static/"
const VOIDWELL_URI = "https://voidwell.com/"

func init() {
	token = os.Getenv("TOKEN")
}

var token string
var buffer = make([][]byte, 0)

func main() {
	if token == "" {
		fmt.Println("No token provided. Please run: mutterblack -t <bot token>")
		return
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	dg.AddHandler(messageCreate)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Mutterblack is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(m.Content, "!void") {
		pat, _ := regexp.Compile(`!void (.*?) (.*)`)
		args := pat.FindAllStringSubmatch(m.Content, -1)[0]

		var commandGroup = args[1]
		var commandArgs = strings.Split(args[2], " ")

		switch commandGroup {
		case "ps2character":
			handlePlanetsideCharacter(s, m, commandArgs)
		case "ps2outfit":
			handlePlanetsideOutfit(s, m, commandArgs)
		case "weather":
			handleWeather(s, m, commandArgs)
		}
	}
}

func sendCoreCommand(commandGroup string, commandAction string, args map[string]string) (resp *http.Response, err error) {
	content, _ := json.Marshal(args)

	var commandUri = MUTTERBLACK_CORE_URI + "command/" + commandGroup + "/" + commandAction

	return http.Post(commandUri, "application/json", bytes.NewBuffer(content))
}

func handleCoreCommand(s *discordgo.Session, m *discordgo.MessageCreate, commandGroup string, commandAction string, args map[string]string) json.RawMessage {
	resp, err := sendCoreCommand(commandGroup, commandAction, args)

	if err != nil {
		log.Println(err)
		s.ChannelMessageSend(m.ChannelID, InterProcessCommunicationFailure)
		return nil
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println(fmt.Sprintf("Failed to get body for %v - %v: %v", commandGroup, commandAction, err))
		s.ChannelMessageSend(m.ChannelID, "Something went wrong :(")
		return nil
	}

	var commandResponse CommandResponse
	err = json.Unmarshal(body, &commandResponse)

	if err != nil {
		log.Println(fmt.Sprintf("Failed to unmarshal for %v - %v: %v", commandGroup, commandAction, err))
		s.ChannelMessageSend(m.ChannelID, "Something went wrong :(")
		return nil
	}

	if commandResponse.Error != "" {
		s.ChannelMessageSend(m.ChannelID, commandResponse.Error)
		return nil
	}

	return commandResponse.Result
}

func handlePlanetsideCharacter(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 1 {
		messagePlanetsideCharacter(s, m, args[0])
	} else {
		characterName, args := args[0], args[1:]
		var weaponName = strings.Join(args[:], " ")
		messagePlanetsideCharacterWeapon(s, m, characterName, weaponName)
	}
}

func handlePlanetsideOutfit(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 1 {
		messagePlanetsideOutfit(s, m, args[0])
	}
}

func handleWeather(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	messageCurrentWeather(s, m, args[0])
}

func messagePlanetsideCharacter(s *discordgo.Session, m *discordgo.MessageCreate, characterName string) {
	values := map[string]string{"characterName": characterName}
	resp := handleCoreCommand(s, m, "planetside2", "character", values)
	if resp == nil {
		return
	}

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
}

func messagePlanetsideCharacterWeapon(s *discordgo.Session, m *discordgo.MessageCreate, characterName string, weaponName string) {
	values := map[string]string{"characterName": characterName, "weaponName": weaponName}
	resp := handleCoreCommand(s, m, "planetside2", "character-weapon", values)
	if resp == nil {
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

	message, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	_ = message

	if err != nil {
		log.Println(err)
	}
}

func messagePlanetsideOutfit(s *discordgo.Session, m *discordgo.MessageCreate, outfitAlias string) {
	values := map[string]string{"outfitAlias": outfitAlias}
	resp := handleCoreCommand(s, m, "planetside2", "outfit", values)
	if resp == nil {
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
}

func messageCurrentWeather(s *discordgo.Session, m *discordgo.MessageCreate, location string) {
	values := map[string]string{"location": location}
	resp := handleCoreCommand(s, m, "weather", "forecast", values)
	if resp == nil {
		return
	}

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

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

	dg.AddHandler(ready)
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

func ready(s *discordgo.Session, event *discordgo.Ready) {

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
			handlePlanetsideCharacter(s, m.ChannelID, commandArgs)
		case "ps2outfit":
			handlePlanetsideOutfit(s, m.ChannelID, commandArgs)
		case "weather":
			handleWeather(s, m.ChannelID, commandArgs)
		}
	}
}

func handlePlanetsideCharacter(s *discordgo.Session, channelId string, args []string) {
	if len(args) == 1 {
		messagePlanetsideCharacter(s, channelId, args[0])
	} else if len(args) == 2 {
		messagePlanetsideCharacterWeapon(s, channelId, args[0], args[1])
	}
}

func messagePlanetsideCharacter(s *discordgo.Session, channelId string, characterName string) {
	values := map[string]string{"characterName": characterName}
	content, _ := json.Marshal(values)
	resp, err := http.Post("http://mutterblack/command/planetside2/character", "application/json", bytes.NewBuffer(content))
	if err != nil {
		s.ChannelMessageSend(channelId, "Failed to retrieve data :(")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var character PlanetsideCharacter
	json.Unmarshal(body, &character)

	lastSaved, _ := time.Parse(time.RFC3339, character.LastSaved)

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: character.Name,
		},
		Color: 0x070707,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: fmt.Sprintf("http://census.daybreakgames.com/files/ps2/images/static/%d.png", character.FactionImageId),
		},
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "Last Seen",
				Value:  fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d UTC", lastSaved.Year(), lastSaved.Month(), lastSaved.Day(), lastSaved.Hour(), lastSaved.Minute(), lastSaved.Second()),
				Inline: false,
			},
			&discordgo.MessageEmbedField{
				Name:   "Battle Rank",
				Value:  fmt.Sprintf("%d", character.BattleRank),
				Inline: false,
			},
			&discordgo.MessageEmbedField{
				Name:   "Outfit",
				Value:  "[" + character.OutfitAlias + "]" + character.OutfitName,
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
		},
	}

	s.ChannelMessageSendEmbed(channelId, embed)
}

func messagePlanetsideCharacterWeapon(s *discordgo.Session, channelId string, characterName string, weaponName string) {
	values := map[string]string{"characterName": characterName, "weaponName": weaponName}
	content, _ := json.Marshal(values)
	resp, err := http.Post("http://mutterblack/command/planetside2/character-weapon", "application/json", bytes.NewBuffer(content))
	if err != nil {
		s.ChannelMessageSend(channelId, "Failed to retrieve data :(")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var weapon PlanetsideCharacterWeapon
	json.Unmarshal(body, &weapon)

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: characterName + " [" + weapon.WeaponName + "]",
		},
		Color: 0x070707,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: fmt.Sprintf("http://census.daybreakgames.com/files/ps2/images/static/%d.png", weapon.WeaponImageId),
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

	message, err := s.ChannelMessageSendEmbed(channelId, embed)
	_ = message

	if err != nil {
		log.Println(err)
	}
}

func handlePlanetsideOutfit(s *discordgo.Session, channelId string, args []string) {
	if len(args) == 1 {
		messagePlanetsideOutfit(s, channelId, args[0])
	}
}

func messagePlanetsideOutfit(s *discordgo.Session, channelId string, outfitAlias string) {
	values := map[string]string{"outfitAlias": outfitAlias}
	content, _ := json.Marshal(values)
	resp, err := http.Post("http://mutterblack/command/planetside2/outfit", "application/json", bytes.NewBuffer(content))
	if err != nil {
		s.ChannelMessageSend(channelId, "Failed to retrieve data :(")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var outfit PlanetsideOutfit
	json.Unmarshal(body, &outfit)

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: "[" + outfit.Alias + "] " + outfit.Name,
		},
		Color: 0x070707,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: fmt.Sprintf("http://census.daybreakgames.com/files/ps2/images/static/%d.png", outfit.FactionImageId),
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

	s.ChannelMessageSendEmbed(channelId, embed)
}

func handleWeather(s *discordgo.Session, channelId string, args []string) {
	messageCurrentWeather(s, channelId, args[0])
}

func messageCurrentWeather(s *discordgo.Session, channelId string, location string) {
	values := map[string]string{"location": location}
	content, _ := json.Marshal(values)
	resp, err := http.Post("http://mutterblack/command/weather/current", "application/json", bytes.NewBuffer(content))
	if err != nil {
		s.ChannelMessageSend(channelId, "Failed to retrieve data :(")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var weather CurrentWeather
	json.Unmarshal(body, &weather)

	var tempCelsius = float32(weather.Temperature)/1.8 - 32

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: weather.City + ", " + weather.Region + " " + weather.Country,
		},
		Color: 0x070707,
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "Temperature",
				Value:  fmt.Sprintf("%d °F (%d °C)", weather.Temperature, int32(tempCelsius)),
				Inline: false,
			},
			&discordgo.MessageEmbedField{
				Name:   "Condition",
				Value:  weather.Condition,
				Inline: false,
			},
		},
	}

	s.ChannelMessageSendEmbed(channelId, embed)
}

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

	if strings.HasPrefix(m.Content, COMMAND_PREFIX+"commands") {
		var commandDescriptions []string
		for i := 0; i < len(COMMANDS_CONFIG); i++ {
			cmd := COMMANDS_CONFIG[i]
			commandDescriptions = append(commandDescriptions, fmt.Sprintf("`%v`", cmd.Description))
		}

		s.ChannelMessageSend(m.ChannelID, strings.Join(commandDescriptions, ", "))
		return
	}

	findCommand(s, m)
}

func findCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	for i := 0; i < len(COMMANDS_CONFIG); i++ {
		var cmd = COMMANDS_CONFIG[i]

		for k := 0; k < len(cmd.Triggers); k++ {
			var trig = COMMAND_PREFIX + cmd.Triggers[k]

			if strings.HasPrefix(m.Content, trig) {
				if cmd.Arguments == nil {
					processCommand(s, m, cmd, nil)
					return
				}

				var pattern = "^" + cmd.Triggers[k]
				for i := 0; i < len(cmd.Arguments); i++ {
					pattern += " (" + cmd.Arguments[i].Pattern + ")"
				}
				pattern += "$"

				var trimmedContent = trimLeftChar(m.Content)
				pat, _ := regexp.Compile(pattern)
				argsMatch := pat.FindAllStringSubmatch(trimmedContent, -1)

				if argsMatch != nil {
					args := argsMatch[0]

					if len(args)-1 == len(cmd.Arguments) {
						_, args := args[0], args[1:]
						processCommand(s, m, cmd, args)
						return
					}
				}
			}
		}
	}
}

func trimLeftChar(s string) string {
	for i := range s {
		if i > 0 {
			return s[i:]
		}
	}
	return s[:0]
}

func processCommand(s *discordgo.Session, m *discordgo.MessageCreate, c CommandConfig, args []string) {
	var values = make(map[string]string)

	for i := 0; i < len(c.Arguments); i++ {
		values[c.Arguments[i].CoreAlias] = args[i]
	}

	resp := handleCoreCommand(s, m, c.CommandGroup, c.CommandGroupAction, values)
	if resp == nil {
		return
	}

	c.Process(s, m, resp)
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

func sendCoreCommand(commandGroup string, commandAction string, args map[string]string) (resp *http.Response, err error) {
	content, _ := json.Marshal(args)

	var commandUri = MUTTERBLACK_CORE_URI + "command/" + commandGroup + "/" + commandAction

	return http.Post(commandUri, "application/json", bytes.NewBuffer(content))
}

package mutterblack

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"runtime/debug"
	"strings"
)

const VersionString string = "2.0.0"

type Bot struct {
	Client          *Discord
	Plugins         map[string]Plugin
	messageChannels []chan Message
}

func MessageRecover() {
	if r := recover(); r != nil {
		log.Println("Recovered:", string(debug.Stack()))
	}
}

func NewBot(token string, clientId string, ownerUserId string) *Bot {
	if token == "" {
		fmt.Println("No token provided. Please run: mutterblack -t <bot token>")
		return nil
	}

	bot := &Bot{
		Plugins: make(map[string]Plugin, 0),
		Client:  NewDiscord("Bot " + token),
	}

	bot.Client.ApplicationClientID = clientId
	bot.Client.OwnerUserID = ownerUserId

	return bot
}

func (b *Bot) getData(plugin Plugin) []byte {
	if b, err := ioutil.ReadFile("data/" + plugin.Name()); err == nil {
		return b
	}
	return nil
}

func (b *Bot) RegisterPlugin(plugin Plugin) {
	if b.Plugins[plugin.Name()] != nil {
		log.Println("Plugin with that name already registered", plugin.Name())
	}
	b.Plugins[plugin.Name()] = plugin
}

func (b *Bot) listen(messageChan <-chan Message) {
	for {
		message := <-messageChan
		plugins := b.Plugins
		for _, plugin := range plugins {
			go plugin.Message(b, b.Client, message)
			if !b.Client.IsMe(message) {
				go findCommandMatch(b, plugin, message)
			}
		}
	}
}

func findCommandMatch(b *Bot, plugin Plugin, message Message) {
	defer MessageRecover()

	if plugin.Commands() == nil || message.Message() == "" {
		return
	}

	for _, commandDefinition := range plugin.Commands() {
		for _, trigger := range commandDefinition.Triggers {
			var trig = b.Client.CommandPrefix() + trigger
			var parts = strings.Split(message.Message(), " ")

			if parts[0] == trig {
				log.Printf("<%s> %s: %s\n", message.Channel(), message.UserName(), message.Message())

				if commandDefinition.Arguments == nil {
					commandDefinition.Callback(b, b.Client, message, nil)
					return
				}

				parsedArgs := extractCommandArguments(message, trig, commandDefinition.Arguments)

				if parsedArgs != nil {
					commandDefinition.Callback(b, b.Client, message, parsedArgs)
					return
				}
			}
		}
	}
}

func (b *Bot) Open() {
	if messageChan, err := b.Client.Open(); err == nil {
		for _, plugin := range b.Plugins {
			plugin.Load(b, b.Client, b.getData(plugin))
		}
		go b.listen(messageChan)
	} else {
		log.Printf("Error creating discord service: %v\n", err)
	}
}

func (b *Bot) Save() {
	if err := os.Mkdir("data", os.ModePerm); err != nil {
		if !os.IsExist(err) {
			log.Println("Error creating service directory.")
		}
	}
	for _, plugin := range b.Plugins {
		if data, err := plugin.Save(); err != nil {
			log.Printf("Error saving plugin %s. %v", plugin.Name(), err)
		} else if data != nil {
			if err := ioutil.WriteFile("data/"+plugin.Name(), data, os.ModePerm); err != nil {
				log.Printf("Error saving plugin %s. %v", plugin.Name(), err)
			}
		}
	}
}

func extractCommandArguments(message Message, trigger string, arguments []CommandDefinitionArgument) map[string]string {
	var argPatterns []string
	for _, argument := range arguments {
		argPatterns = append(argPatterns, fmt.Sprintf("(?P<%s>%s)", argument.Alias, argument.Pattern))
	}
	var pattern = fmt.Sprintf("^%s$", strings.Join(argPatterns, " "))

	var trimmedContent = strings.TrimPrefix(message.Message(), fmt.Sprintf("%s ", trigger))
	pat := regexp.MustCompile(pattern)
	argsMatch := pat.FindStringSubmatch(trimmedContent)

	parsedArgs := make(map[string]string)

	if argsMatch == nil || len(argsMatch) == 1 {
		return nil
	}

	for i := 1; i < len(argsMatch); i++ {
		parsedArgs[pat.SubexpNames()[i]] = argsMatch[i]
	}

	if len(parsedArgs) != len(arguments) {
		return nil
	}

	return parsedArgs
}

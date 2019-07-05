package uwutranslatorplugin

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/lampjaw/mutterblack.discord"
)

type uwutranslatorPlugin struct {
	sync.RWMutex
}

func (p *uwutranslatorPlugin) Commands() []mutterblack.CommandDefinition {
	return []mutterblack.CommandDefinition{
		mutterblack.CommandDefinition{
			CommandGroup: p.Name(),
			CommandID:    "translate-uwu",
			Triggers: []string{
				"twanswate",
			},
			Arguments:   nil,
			Description: "Get stats for a player.",
			Callback:    p.runTranslateCommand,
		},
	}
}

func (p *uwutranslatorPlugin) Name() string {
	return "uwuTranslator"
}

func (p *uwutranslatorPlugin) Load(bot *mutterblack.Bot, client *mutterblack.Discord, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
		}
	}

	return nil
}

func (p *uwutranslatorPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

func (p *uwutranslatorPlugin) Help(bot *mutterblack.Bot, client *mutterblack.Discord, message mutterblack.Message, detailed bool) []string {
	return []string{
		mutterblack.CommandHelp(client, "twanswate", "", "Translate the previous message UwU.")[0],
	}
}

func (p *uwutranslatorPlugin) Message(bot *mutterblack.Bot, client *mutterblack.Discord, message mutterblack.Message) {

}

func (p *uwutranslatorPlugin) Stats(bot *mutterblack.Bot, client *mutterblack.Discord, message mutterblack.Message) []string {
	return nil
}

func New() mutterblack.Plugin {
	return &uwutranslatorPlugin{}
}

func (p *uwutranslatorPlugin) runTranslateCommand(bot *mutterblack.Bot, client *mutterblack.Discord, message mutterblack.Message, args map[string]string, trigger string) {
	previousMessages, err := client.GetMessages(message.Channel(), 1, message.MessageID())
	if err != nil {
		p.RLock()
		client.SendMessage(message.Channel(), fmt.Sprintf("%s", err))
		p.RUnlock()
		return
	}

	if previousMessages == nil || len(previousMessages) == 0 {
		p.RLock()
		client.SendMessage(message.Channel(), "Unable to find a message to translate.")
		p.RUnlock()
		return
	}

	if client.IsMe(previousMessages[0]) {
		return
	}

	textArg := make(map[string]string)
	textArg["text"] = previousMessages[0].Message()

	resp, err := mutterblack.SendCoreCommand("uwutranslator", "translate", textArg)

	if err != nil {
		p.RLock()
		client.SendMessage(message.Channel(), fmt.Sprintf("%s", err))
		p.RUnlock()
		return
	}

	var translatedText string
	json.Unmarshal(resp, &translatedText)

	p.RLock()
	client.SendMessage(message.Channel(), translatedText)
	p.RUnlock()
}

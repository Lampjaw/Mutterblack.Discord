package mutterblack

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
)

type helpPlugin struct {
	Private map[string]bool
}

func (p *helpPlugin) Name() string {
	return "Help"
}

func (p *helpPlugin) Commands() []CommandDefinition {
	return nil
}

// Help returns a list of help strings that are printed when the user requests them.
func (p *helpPlugin) Help(bot *Bot, client *Discord, message Message, detailed bool) []string {
	privs := !client.IsPrivate(message) && client.IsModerator(message)
	if detailed && !privs {
		return nil
	}

	commands := []string{}

	for _, plugin := range bot.Plugins {
		hasDetailed := false

		if plugin == p {
			hasDetailed = privs
		} else {
			t := plugin.Help(bot, client, message, true)
			hasDetailed = t != nil && len(t) > 0
		}

		if hasDetailed {
			commands = append(commands, strings.ToLower(plugin.Name()))
		}
	}

	sort.Strings(commands)

	help := []string{}

	if len(commands) > 0 {
		help = append(help, CommandHelp(client, "help", "[topic]", fmt.Sprintf("Returns help for a specific topic. Available topics: `%s`", strings.Join(commands, ", ")))[0])
	}

	if detailed {
		help = append(help, []string{
			CommandHelp(client, "setprivatehelp", "", "Sets help text to be sent through private messages in this channel.")[0],
			CommandHelp(client, "setpublichelp", "", "Sets the default help behavior for this channel.")[0],
		}...)
	}

	return help
}

func (p *helpPlugin) Message(bot *Bot, client *Discord, message Message) {
	if !client.IsMe(message) {
		if MatchesCommand(client, "help", message) || MatchesCommand(client, "command", message) || MatchesCommand(client, "commands", message) {
			_, parts := ParseCommand(client, message)

			help := []string{}

			for _, plugin := range bot.Plugins {
				var h []string
				if len(parts) == 0 {
					if plugin.Commands() == nil {
						h = plugin.Help(bot, client, message, false)
					} else {
						for _, commandDefinition := range plugin.Commands() {
							h = append(h, commandDefinition.Help(client))
						}
					}
				} else if len(parts) == 1 && strings.ToLower(parts[0]) == strings.ToLower(plugin.Name()) {
					if plugin.Commands() == nil {
						h = plugin.Help(bot, client, message, true)
					} else {
						for _, commandDefinition := range plugin.Commands() {
							h = append(h, commandDefinition.Help(client))
						}
					}
				}
				if h != nil && len(h) > 0 {
					help = append(help, h...)
				}
			}

			if len(parts) == 0 {
				sort.Strings(help)
				help = append([]string{fmt.Sprintf("All commands can be used in private messages without the `%s` prefix.", client.CommandPrefix())}, help...)
			}

			if len(parts) != 0 && len(help) == 0 {
				help = []string{fmt.Sprintf("Unknown topic: %s", parts[0])}
			}

			if p.Private[message.Channel()] {
				client.SendMessage(message.Channel(), "Help has been sent via private message.")
				client.PrivateMessage(message.UserID(), strings.Join(help, "\n"))
			} else {
				client.SendMessage(message.Channel(), strings.Join(help, "\n"))
			}
		} else if MatchesCommand(client, "setprivatehelp", message) && !client.IsPrivate(message) {
			if !client.IsModerator(message) {
				return
			}

			p.Private[message.Channel()] = true

			client.PrivateMessage(message.UserID(), fmt.Sprintf("Help text in <#%s> will be sent through private messages.", message.Channel()))
		} else if MatchesCommand(client, "setpublichelp", message) && !client.IsPrivate(message) {
			if !client.IsModerator(message) {
				return
			}

			p.Private[message.Channel()] = false

			client.PrivateMessage(message.UserID(), fmt.Sprintf("Help text in <#%s> will be sent publically.", message.Channel()))
		}
	}
}

// Load will load plugin state from a byte array.
func (p *helpPlugin) Load(bot *Bot, client *Discord, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
		}
	}
	return nil
}

// Save will save plugin state to a byte array.
func (p *helpPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

// Stats will return the stats for a plugin.
func (p *helpPlugin) Stats(bot *Bot, client *Discord, message Message) []string {
	return nil
}

// NeHelpPlugin will create a new help plugin.
func NewHelpPlugin() Plugin {
	p := &helpPlugin{
		Private: make(map[string]bool),
	}
	return p
}

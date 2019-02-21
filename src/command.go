package mutterblack

import (
	"fmt"
)

type CommandDefinition struct {
	Description string
	Triggers    []string
	Arguments   []CommandDefinitionArgument
	Callback    func(bot *Bot, client *Discord, message Message, args map[string]string)
}

type CommandDefinitionArgument struct {
	Optional bool
	Pattern  string
	Alias    string
}

func (c *CommandDefinition) Help(client *Discord) string {
	commandString := fmt.Sprintf("%s%s", client.CommandPrefix(), c.Triggers[0])

	if len(c.Arguments) > 0 {
		for _, argument := range c.Arguments {
			commandString = fmt.Sprintf("%s <%s>", commandString, argument.Alias)
		}
	}

	return fmt.Sprintf("`%s` - %s", commandString, c.Description)
}

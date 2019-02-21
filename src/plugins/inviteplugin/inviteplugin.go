package inviteplugin

import (
	"fmt"
	"log"
	"strings"

	"github.com/lampjaw/mutterblack.discord/src"
)

func discordInviteID(id string) string {
	id = strings.Replace(id, "://discordapp.com/invite/", "://discord.gg/", -1)
	id = strings.Replace(id, "https://discord.gg/", "", -1)
	id = strings.Replace(id, "http://discord.gg/", "", -1)
	return id
}

func InviteHelp(bot *mutterblack.Bot, client *mutterblack.Discord, message mutterblack.Message) (string, string) {
	if client.ApplicationClientID != "" {
		return "", fmt.Sprintf("Returns a URL to add %s to your server.", client.UserName())
	}
	return "<discordinvite>", "Joins the provided Discord server."
}

func InviteCommand(bot *mutterblack.Bot, client *mutterblack.Discord, message mutterblack.Message, command string, parts []string) {
	if client.ApplicationClientID != "" {
		client.SendMessage(message.Channel(), fmt.Sprintf("Please visit <https://discordapp.com/oauth2/authorize?client_id=%s&scope=bot> to add %s to your server.", client.ApplicationClientID, client.UserName()))
		return
	}

	if len(parts) == 1 {
		join := parts[0]
		join = discordInviteID(join)
		if err := client.Join(join); err != nil {
			if err == mutterblack.ErrAlreadyJoined {
				client.PrivateMessage(message.UserID(), "I have already joined that server.")
				return
			}
			log.Println("Error joining discord %v", err)
		} else {
			client.PrivateMessage(message.UserID(), "I have joined that server.")
		}
	}
}

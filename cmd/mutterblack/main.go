package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/lampjaw/mutterblack.discord"
	"github.com/lampjaw/mutterblack.discord/plugins/inviteplugin"
	"github.com/lampjaw/mutterblack.discord/plugins/planetsidetwoplugin"
	"github.com/lampjaw/mutterblack.discord/plugins/statsplugin"
	"github.com/lampjaw/mutterblack.discord/plugins/uwutranslatorplugin"
	"github.com/lampjaw/mutterblack.discord/plugins/weatherplugin"
)

func init() {
	token = os.Getenv("TOKEN")
	clientID = os.Getenv("CLIENT_ID")
	ownerUserID = os.Getenv("OWNER_USER_ID")
}

var token string
var clientID string
var ownerUserID string
var buffer = make([][]byte, 0)

func main() {
	q := make(chan bool)

	if token == "" {
		fmt.Println("No token provided. Please run: mutterblack -t <bot token>")
		return
	}

	bot := mutterblack.NewBot(token, clientID, ownerUserID)

	commandPlugin := mutterblack.NewCommandPlugin()
	commandPlugin.AddCommand("invite", inviteplugin.InviteCommand, inviteplugin.InviteHelp)
	commandPlugin.AddCommand("join", inviteplugin.InviteCommand, nil)
	commandPlugin.AddCommand("stats", statsplugin.StatsCommand, statsplugin.StatsHelp)
	commandPlugin.AddCommand("info", statsplugin.StatsCommand, nil)
	commandPlugin.AddCommand("stat", statsplugin.StatsCommand, nil)
	commandPlugin.AddCommand("quit", func(bot *mutterblack.Bot, client *mutterblack.Discord, message mutterblack.Message, args string, parts []string) {
		if client.IsBotOwner(message) {
			q <- true
		}
	}, nil)

	bot.RegisterPlugin(commandPlugin)
	bot.RegisterPlugin(mutterblack.NewHelpPlugin())
	bot.RegisterPlugin(weatherplugin.New())
	bot.RegisterPlugin(planetsidetwoplugin.New())
	bot.RegisterPlugin(uwutranslatorplugin.New())

	bot.Open()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	t := time.Tick(1 * time.Minute)

out:
	for {
		select {
		case <-q:
			break out
		case <-c:
			break out
		case <-t:
			bot.Save()
		}
	}

	bot.Save()
}

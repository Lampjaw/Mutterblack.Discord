package statsplugin

import (
	"bytes"
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/lampjaw/mutterblack.discord"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
)

var statsStartTime = time.Now()

func getDurationString(duration time.Duration) string {
	return fmt.Sprintf(
		"%0.2d:%02d:%02d",
		int(duration.Hours()),
		int(duration.Minutes())%60,
		int(duration.Seconds())%60,
	)
}

// StatsCommand returns bot statistics.
func StatsCommand(bot *mutterblack.Bot, client *mutterblack.Discord, message mutterblack.Message, command string, parts []string) {
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)

	w := &tabwriter.Writer{}
	buf := &bytes.Buffer{}

	w.Init(buf, 0, 4, 0, ' ', 0)
	fmt.Fprintf(w, "```\n")
	fmt.Fprintf(w, "mutterblack: \t%s\n", mutterblack.VersionString)
	fmt.Fprintf(w, "Discordgo: \t%s\n", discordgo.VERSION)
	fmt.Fprintf(w, "Go: \t%s\n", runtime.Version())
	fmt.Fprintf(w, "Uptime: \t%s\n", getDurationString(time.Now().Sub(statsStartTime)))
	fmt.Fprintf(w, "Memory used: \t%s / %s (%s garbage collected)\n", humanize.Bytes(stats.Alloc), humanize.Bytes(stats.Sys), humanize.Bytes(stats.TotalAlloc))
	fmt.Fprintf(w, "Concurrent tasks: \t%d\n", runtime.NumGoroutine())

	fmt.Fprintf(w, "Connected servers: \t%d\n", client.ChannelCount())
	if len(client.Sessions) > 1 {
		shards := 0
		for _, s := range client.Sessions {
			if s.DataReady {
				shards++
			}
		}
		if shards == len(client.Sessions) {
			fmt.Fprintf(w, "Shards: \t%d\n", shards)
		} else {
			fmt.Fprintf(w, "Shards: \t%d (%d connected)\n", len(client.Sessions), shards)
		}
		guild, err := client.Channel(message.Channel())
		if err == nil {
			id, err := strconv.Atoi(guild.ID)
			if err == nil {
				fmt.Fprintf(w, "Current shard: \t%d\n", ((id>>22)%len(client.Sessions) + 1))
			}
		}
	}

	plugins := bot.Plugins
	names := []string{}
	for _, plugin := range plugins {
		names = append(names, plugin.Name())
		sort.Strings(names)
	}

	for _, name := range names {
		stats := plugins[name].Stats(bot, client, message)
		for _, stat := range stats {
			fmt.Fprint(w, stat)
		}
	}

	fmt.Fprintf(w, "\n```")

	w.Flush()
	out := buf.String()

	end := ""

	if end != "" {
		out += "\n" + end
	}
	client.SendMessage(message.Channel(), out)
}

// StatsHelp is the help for the stats command.
var StatsHelp = mutterblack.NewCommandHelp("", "Lists bot statistics.")

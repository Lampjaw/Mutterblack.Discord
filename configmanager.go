package mutterblack

import (
	"encoding/json"
	"fmt"
)

type GuildConfiguration struct {
	Platform              string
	GuildID               string `json:"key"`
	Prefix                string
	AllowedChannels       []string
	AllowedRoles          []string
	CommandConfigurations map[string]*GuildCommandConfiguration
}

type GuildCommandConfiguration struct {
	CommandID       string
	AllowedChannels []string
	AllowedRoles    []string
	Enabled         bool
}

func newGuildConfiguration(guildID string) *GuildConfiguration {
	return &GuildConfiguration{
		Platform:              "Discord",
		GuildID:               guildID,
		Prefix:                "?",
		AllowedChannels:       make([]string, 0),
		AllowedRoles:          make([]string, 0),
		CommandConfigurations: make(map[string]*GuildCommandConfiguration),
	}
}

func newGuildCommandConfiguration(commandID string) *GuildCommandConfiguration {
	return &GuildCommandConfiguration{
		CommandID:       commandID,
		AllowedChannels: make([]string, 0),
		AllowedRoles:    make([]string, 0),
		Enabled:         true,
	}
}

func findGuildConfiguration(guildID string) *GuildConfiguration {
	var path = fmt.Sprintf("configuration/discord/%s", guildID)
	resp, err := SendCoreGet(path)
	if err != nil {
		return nil
	}

	var configuration *GuildConfiguration
	json.Unmarshal(resp, &configuration)

	if configuration == nil {
		return createGuildConfiguration(guildID)
	}

	return configuration
}

func createGuildConfiguration(guildID string) *GuildConfiguration {
	config := newGuildConfiguration(guildID)

	var path = fmt.Sprintf("configuration/discord/%s", guildID)
	resp, err := SendCorePost(path, config)
	if err != nil {
		return nil
	}

	var configuration *GuildConfiguration
	json.Unmarshal(resp, &configuration)

	return configuration
}

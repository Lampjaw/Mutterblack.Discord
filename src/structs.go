package main

import (
	"encoding/json"
)

type CommandResponse struct {
	Error  string          `json:"error"`
	Result json.RawMessage `json:"result"`
}

type PlanetsideCharacter struct {
	CharacterId    string  `json:"id"`
	World          string  `json:"world"`
	Name           string  `json:"name"`
	LastSaved      string  `json:"lastSaved"`
	FactionId      int     `json:"factionId"`
	FactionName    string  `json:"factionName"`
	FactionImageId int     `json:"factionImageId"`
	BattleRank     int     `json:"battleRank"`
	OutfitAlias    string  `json:"outfitAlias"`
	OutfitName     string  `json:"outfitName"`
	Kills          int     `json:"kills"`
	Deaths         int     `json:"deaths"`
	PlayTime       int     `json:"playTime"`
	Score          int     `json:"score"`
	KillDeathRatio float32 `json:"killDeathRatio"`
	HeadshotRatio  float32 `json:"headshotRatio"`
	KillsPerHour   float32 `json:"killsPerHour"`
	SiegeLevel     float32 `json:"siegeLevel"`
}

type PlanetsideCharacterWeapon struct {
	CharacterId         string  `json:"characterId"`
	CharacterName       string  `json:"characterName"`
	ItemId              int     `json:"itemId"`
	WeaponName          string  `json:"weaponName"`
	WeaponImageId       int     `json:"weaponImageId"`
	Kills               int     `json:"kills"`
	Deaths              int     `json:"deaths"`
	PlayTime            int     `json:"playTime"`
	Score               int     `json:"score"`
	Headshots           int     `json:"headshots"`
	KillDeathRatio      float32 `json:"killDeathRatio"`
	HeadshotRatio       float32 `json:"headshotRatio"`
	KillsPerHour        float32 `json:"killsPerHour"`
	Accuracy            float32 `json:"accuracy"`
	KillDeathRatioGrade string  `json:"killDeathRatioGrade"`
	HeadshotRatioGrade  string  `json:"headshotRatioGrade"`
	KillsPerHourGrade   string  `json:"killsPerHourGrade"`
	AccuracyGrade       string  `json:"accuracyGrade"`
}

type PlanetsideOutfit struct {
	OutfitId       string `json:"outfitId"`
	Name           string `json:"name"`
	Alias          string `json:"alias"`
	FactionName    string `json:"factionName"`
	FactionImageId int    `json:"factionImageId"`
	WorldName      string `json:"worldName"`
	LeaderName     string `json:"leaderName"`
	MemberCount    int    `json:"memberCount"`
	Activity7Days  int    `json:"activity7Days"`
	Activity30Days int    `json:"activity30Days"`
	Activity90Days int    `json:"activity90Days"`
}

type CurrentWeather struct {
	City     string       `json:"city"`
	Country  string       `json:"country"`
	Region   string       `json:"region"`
	Forecast []WeatherDay `json:"forecast"`
}

type WeatherDay struct {
	Date      string `json:"date"`
	Day       string `json:"day"`
	High      int    `json:"high"`
	Low       int    `json:"low"`
	Condition string `json:"text"`
}

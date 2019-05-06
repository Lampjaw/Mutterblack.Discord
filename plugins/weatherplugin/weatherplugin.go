package weatherplugin

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/lampjaw/mutterblack.discord"
)

type weatherPlugin struct {
	sync.RWMutex
}

type CurrentWeather struct {
	City         string  `json:"city"`
	Country      string  `json:"country"`
	Region       string  `json:"region"`
	Condition    string  `json:"condition"`
	Temperature  int     `json:"temperature"`
	Humidity     int     `json:"humidity"`
	WindChill    int     `json:"windChill"`
	WindSpeed    float32 `json:"windSpeed"`
	ForecastHigh int     `json:"forecastHigh"`
	ForecastLow  int     `json:"forecastLow"`
}

type ForecastWeather struct {
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

func (p *weatherPlugin) Commands() []mutterblack.CommandDefinition {
	return []mutterblack.CommandDefinition{
		mutterblack.CommandDefinition{
			CommandGroup: p.Name(),
			CommandID:    "weather-current",
			Triggers: []string{
				"w",
				"weather",
			},
			Arguments: []mutterblack.CommandDefinitionArgument{
				mutterblack.CommandDefinitionArgument{
					Pattern: ".*",
					Alias:   "location",
				},
			},
			Description: "Get the current weather condition.",
			Callback:    p.runCurrentWeatherCommand,
		},
		mutterblack.CommandDefinition{
			CommandGroup: p.Name(),
			CommandID:    "weather-forecast",
			Triggers: []string{
				"wf",
				"forecast",
			},
			Arguments: []mutterblack.CommandDefinitionArgument{
				mutterblack.CommandDefinitionArgument{
					Pattern: ".*",
					Alias:   "location",
				},
			},
			Description: "Get the forecasted weather conditions.",
			Callback:    p.runForecastWeatherCommand,
		},
	}
}

func (p *weatherPlugin) Name() string {
	return "Weather"
}

func (p *weatherPlugin) Load(bot *mutterblack.Bot, client *mutterblack.Discord, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
		}
	}

	return nil
}

func (p *weatherPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

func (p *weatherPlugin) Help(bot *mutterblack.Bot, client *mutterblack.Discord, message mutterblack.Message, detailed bool) []string {
	return []string{
		mutterblack.CommandHelp(client, "w", "<location>", "Returns the current weather.")[0],
		mutterblack.CommandHelp(client, "wf", "<location>", "Returns a 5 day forecast.")[0],
	}
}

func (p *weatherPlugin) Message(bot *mutterblack.Bot, client *mutterblack.Discord, message mutterblack.Message) {

}

func (p *weatherPlugin) Stats(bot *mutterblack.Bot, client *mutterblack.Discord, message mutterblack.Message) []string {
	return nil
}

func New() mutterblack.Plugin {
	return &weatherPlugin{}
}

func (p *weatherPlugin) runCurrentWeatherCommand(bot *mutterblack.Bot, client *mutterblack.Discord, message mutterblack.Message, args map[string]string) {
	resp, err := mutterblack.SendCoreCommand("weather", "current", args)

	if err != nil {
		p.RLock()
		client.SendMessage(message.Channel(), fmt.Sprintf("%s", err))
		p.RUnlock()
		return
	}

	var weather CurrentWeather
	json.Unmarshal(resp, &weather)

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: weather.City + ", " + weather.Region + " - " + weather.Country,
		},
		Color:       0x070707,
		Description: fmt.Sprintf("Currently %s and %s with a high of %s and a low of %s.", convertToTempString(weather.Temperature), weather.Condition, convertToTempString(weather.ForecastHigh), convertToTempString(weather.ForecastLow)),
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "Wind Speed",
				Value:  fmt.Sprintf("%0.1f MpH", weather.WindSpeed),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Wind Chill",
				Value:  convertToTempString(weather.WindChill),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Humidity",
				Value:  fmt.Sprintf("%d%%", weather.Humidity),
				Inline: true,
			},
		},
	}

	p.RLock()
	client.SendEmbedMessage(message.Channel(), embed)
	p.RUnlock()
}

func (p *weatherPlugin) runForecastWeatherCommand(bot *mutterblack.Bot, client *mutterblack.Discord, message mutterblack.Message, args map[string]string) {
	resp, err := mutterblack.SendCoreCommand("weather", "forecast", args)

	if err != nil {
		p.RLock()
		client.SendMessage(message.Channel(), fmt.Sprintf("%s", err))
		p.RUnlock()
		return
	}

	var weather ForecastWeather
	json.Unmarshal(resp, &weather)

	var messageFields []*discordgo.MessageEmbedField
	for i := 0; i < 5; i++ {
		var field = &discordgo.MessageEmbedField{
			Name:   weather.Forecast[i].Date,
			Value:  createWeatherDay(weather.Forecast[i]),
			Inline: false,
		}
		messageFields = append(messageFields, field)
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: weather.City + ", " + weather.Region + " - " + weather.Country,
		},
		Color:  0x070707,
		Fields: messageFields,
	}

	p.RLock()
	client.SendEmbedMessage(message.Channel(), embed)
	p.RUnlock()
}

func createWeatherDay(d WeatherDay) string {
	var temperatureHigh = convertToTempString(d.High)
	var temperatureLow = convertToTempString(d.Low)
	return fmt.Sprintf("%s: %s / %s - %s", d.Day, temperatureHigh, temperatureLow, d.Condition)
}

func convertToTempString(temp int) string {
	var tempCelsius = convertToCelsius(temp)
	return fmt.Sprintf("%d °F (%d °C)", temp, int32(tempCelsius))
}

func convertToCelsius(temp int) float32 {
	return (float32(temp) - 32) / 1.8
}

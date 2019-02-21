package mutterblack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const MUTTERBLACK_CORE_URI = "http://mutterblack:5000/"
//const MUTTERBLACK_CORE_URI = "http://localhost:8080/"

type CommandResponse struct {
	Error  string          `json:"error"`
	Result json.RawMessage `json:"result"`
}

func SendCoreCommand(commandGroup string, commandAction string, args map[string]string) (json.RawMessage, error) {
	resp, err := handleCoreCommand(commandGroup, commandAction, args)

	if err != nil {
		log.Println(err)
		return nil, errors.New(InterProcessCommunicationFailure)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var commandResponse CommandResponse
	err = json.Unmarshal(body, &commandResponse)

	if err != nil {
		log.Println(fmt.Sprintf("Failed to unmarshal for %v - %v: %v", commandGroup, commandAction, err))
		return nil, errors.New("Something went wrong :(")
	}

	if commandResponse.Error != "" {
		return nil, errors.New(commandResponse.Error)
	}

	return commandResponse.Result, nil
}

func handleCoreCommand(commandGroup string, commandAction string, args map[string]string) (resp *http.Response, err error) {
	content, _ := json.Marshal(args)

	var commandUri = MUTTERBLACK_CORE_URI + "command/" + commandGroup + "/" + commandAction

	return http.Post(commandUri, "application/json", bytes.NewBuffer(content))
}

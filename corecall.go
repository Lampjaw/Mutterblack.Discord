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
	var path = "command/" + commandGroup + "/" + commandAction
	resp, err := handleCoreRequestPost(path, args)
	return handleResponse(path, resp, err)
}

func SendCoreGet(path string) (json.RawMessage, error) {
	resp, err := handleCoreRequestGet(path)
	return handleResponse(path, resp, err)
}

func SendCorePost(path string, content interface{}) (json.RawMessage, error) {
	resp, err := handleCoreRequestPost(path, content)
	return handleResponse(path, resp, err)
}

func handleResponse(path string, resp *http.Response, err error) (json.RawMessage, error) {
	if err != nil {
		log.Println(err)
		return nil, errors.New(InterProcessCommunicationFailure)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var commandResponse CommandResponse
	err = json.Unmarshal(body, &commandResponse)

	if err != nil {
		log.Println(fmt.Sprintf("Failed to unmarshal for %v: %v", path, err))
		return nil, errors.New("Something went wrong :(")
	}

	if commandResponse.Error != "" {
		return nil, errors.New(commandResponse.Error)
	}

	return commandResponse.Result, nil
}

func handleCoreRequestGet(path string) (resp *http.Response, err error) {
	var commandURI = GetURI(path)
	return http.Get(commandURI)
}

func handleCoreRequestPost(path string, content interface{}) (resp *http.Response, err error) {
	var commandURI = GetURI(path)
	contentBytes, _ := json.Marshal(content)
	return http.Post(commandURI, "application/json", bytes.NewBuffer(contentBytes))
}

func GetURI(path string) string {
	return MUTTERBLACK_CORE_URI + path
}

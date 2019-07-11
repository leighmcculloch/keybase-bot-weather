package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/keybase/go-keybase-chat-bot/kbchat"
)

func main() {
	const kbLocation = "keybase"

	var err error

	var kbc *kbchat.API
	if kbc, err = kbchat.Start(kbchat.RunOptions{KeybaseLocation: kbLocation}); err != nil {
		fmt.Fprintf(os.Stderr, "error starting keybase chat: %v\n", err)
		return
	}

	var sub kbchat.NewSubscription
	if sub, err = kbc.ListenForNewTextMessages(); err != nil {
		fmt.Fprintf(os.Stderr, "error listening for new text messages: %v\n", err)
		return
	}

	for {
		var msg kbchat.SubscriptionMessage
		if msg, err = sub.Read(); err != nil {
			fmt.Fprintf(os.Stderr, "error reading message: %v\n", err)
			continue
		}
		fmt.Fprintf(os.Stderr, "message: %#v\n", msg)
		if msg.Message.Content.Type != "text" {
			continue
		}
		if msg.Message.Sender.Username == kbc.GetUsername() {
			continue
		}

		msgText := msg.Message.Content.Text.Body
		fmt.Fprintf(os.Stderr, "seeing message text: %s\n", msgText)
		if !strings.HasPrefix(msgText, "/weather ") {
			continue
		}
		locationName := strings.TrimPrefix(msgText, "/weather ")

		locationQueryVals := url.Values{}
		locationQueryVals.Add("query", locationName)
		locationResp, err := http.Get("https://www.metaweather.com/api/location/search/?" + locationQueryVals.Encode())
		if err != nil {
			fmt.Fprintf(os.Stderr, "error finding location %s: %v\n", locationName, err)
			continue
		}
		locationRespParsed := []struct {
			Woeid int `json:"woeid"`
		}{}
		if err = json.NewDecoder(locationResp.Body).Decode(&locationRespParsed); err != nil {
			fmt.Fprintf(os.Stderr, "error decoding location response for location %s: %v\n", locationName, err)
			continue
		}

		infoResp, err := http.Get("https://www.metaweather.com/api/location/" + strconv.Itoa(locationRespParsed[0].Woeid))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error getting weather for location %s: %v\n", locationName, err)
			continue
		}

		infoRespParsed := struct {
			ConsolidatedWeather []struct {
				ApplicableDate   string `json:"applicable_date"`
				WeatherStateName string `json:"weather_state_name"`
			} `json:"consolidated_weather"`
		}{}
		if err = json.NewDecoder(infoResp.Body).Decode(&infoRespParsed); err != nil {
			fmt.Fprintf(os.Stderr, "error decoding info response for location %d: %v\n", locationRespParsed[0].Woeid, err)
			continue
		}

		msgSend := fmt.Sprintf("Weather for %s:\n%s - %s", locationName, infoRespParsed.ConsolidatedWeather[0].ApplicableDate, infoRespParsed.ConsolidatedWeather[0].WeatherStateName)
		if err := kbc.SendMessage(msg.Message.Channel, msgSend); err != nil {
			fmt.Fprintf(os.Stderr, "error echoing message: %v\n", err)
			continue
		}
	}

}

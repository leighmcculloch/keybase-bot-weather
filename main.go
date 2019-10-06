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

func health() {
	ok200 := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	}
	mux := http.ServeMux{}
	mux.HandleFunc("/", ok200)
	svr := http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: &mux,
	}
	svr.ListenAndServe()
}

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

	fmt.Fprintf(os.Stderr, "bot started\n")
	defer fmt.Fprintf(os.Stderr, "bot shutting down\n")

	go health()

	for {
		var msg kbchat.SubscriptionMessage
		if msg, err = sub.Read(); err != nil {
			fmt.Fprintf(os.Stderr, "error reading message: %v\n", err)
			continue
		}
		if msg.Message.Content.Type != "text" {
			continue
		}
		if msg.Message.Sender.Username == kbc.GetUsername() {
			continue
		}

		msgText := msg.Message.Content.Text.Body
		if !strings.HasPrefix(msgText, "/weather ") {
			continue
		}
		fmt.Fprintf(os.Stderr, "channel: %s user: %s message: %s\n", msg.Message.Channel.Name, msg.Message.Sender.Username, msgText)
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

		if len(locationRespParsed) == 0 {
			msgSend := fmt.Sprintf("Location %q not found.", locationName)
			if err := kbc.SendMessage(msg.Message.Channel, msgSend); err != nil {
				fmt.Fprintf(os.Stderr, "error sending message: %v\n", err)
			}
			continue
		}

		infoResp, err := http.Get("https://www.metaweather.com/api/location/" + strconv.Itoa(locationRespParsed[0].Woeid))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error getting weather for location %s: %v\n", locationName, err)
			continue
		}

		infoRespParsed := struct {
			Title               string `json:"title"`
			Timezone            string `json:"timezone"`
			ConsolidatedWeather []struct {
				ApplicableDate   string  `json:"applicable_date"`
				TheTempC         float64 `json:"the_temp"`
				MinTempC         float64 `json:"min_temp"`
				MaxTempC         float64 `json:"max_temp"`
				WeatherStateName string  `json:"weather_state_name"`
			} `json:"consolidated_weather"`
		}{}
		if err = json.NewDecoder(infoResp.Body).Decode(&infoRespParsed); err != nil {
			fmt.Fprintf(os.Stderr, "error decoding info response for location %d: %v\n", locationRespParsed[0].Woeid, err)
			continue
		}
		if len(infoRespParsed.ConsolidatedWeather) == 0 {
			fmt.Fprintf(os.Stderr, "error no data for woeid %d\n", locationRespParsed[0].Woeid)
			continue
		}
		weather := infoRespParsed.ConsolidatedWeather[0]
		fahrenheit := strings.HasPrefix(infoRespParsed.Timezone, "US/")
		msgSend := ""
		if fahrenheit {
			msgSend = fmt.Sprintf(
				"%s today is %.1fF (Min: %.1fF, Max: %.1fF), %s.",
				infoRespParsed.Title,
				cToF(weather.TheTempC),
				cToF(weather.MinTempC),
				cToF(weather.MaxTempC),
				weather.WeatherStateName,
			)
		} else {
			msgSend = fmt.Sprintf(
				"%s today is %.1fC (Min: %.1fC, Max: %.1fC), %s.",
				infoRespParsed.Title,
				weather.TheTempC,
				weather.MinTempC,
				weather.MaxTempC,
				weather.WeatherStateName,
			)
		}
		if err := kbc.SendMessage(msg.Message.Channel, msgSend); err != nil {
			fmt.Fprintf(os.Stderr, "error sending message: %v\n", err)
			continue
		}
	}

}

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/nuuls/log"
)

const (
	pointsURL   = "https://api.streamelements.com/kappa/v1/points/%s/%d"
	chattersURL = "https://tmi.twitch.tv/group/user/%s/chatters"
)

type config struct {
	Channel  string `json:"channel"`
	JwtToken string `json:"jwtToken"`
	Interval uint   `json:"interval"`
	Points   int    `json:"points"`
}

var cfg *config

func main() {
	log.AddLogger(log.DefaultLogger)
	loadConfig()
	cfg.Interval = 1
	for _ = range time.Tick(time.Duration(cfg.Interval) * time.Minute) {
		chatters := getChatters()
		if chatters != nil {
			log.Info("updating points for", len(chatters), "chatters")
			updatePoints(chatters)
		}
	}
}

func getChatters() []string {
	res, err := http.Get(fmt.Sprintf(chattersURL, cfg.Channel))
	if err != nil {
		log.Error(err)
		return nil
	}
	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Critical(err)
		return nil
	}
	var s struct {
		Count    int                 `json:"chatter_count"`
		Chatters map[string][]string `json:"chatters"`
	}
	err = json.Unmarshal(bs, &s)
	if err != nil {
		log.Error(err)
		return nil
	}
	chatters := make([]string, 0, s.Count)
	for _, v := range s.Chatters {
		chatters = append(chatters, v...)
	}
	return chatters
}

func updatePoints(chatters []string) {
	for _, user := range chatters {
		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf(pointsURL, user, cfg.Points), nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+cfg.JwtToken)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Critical(err)
		}
		if res.StatusCode == 200 {
			log.Info("updated", user, "'s points")
		} else {
			log.Error("something went wrong while updating", user, "'s points", res.StatusCode, res.Status)
		}
	}
}

func loadConfig() {
	bs, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}
	cfg = &config{}
	err = json.Unmarshal(bs, cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("loaded config")
}

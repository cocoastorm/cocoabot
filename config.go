package main

import (
	"net/http"
	"os"

	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

type Config struct {
	BotToken   string
	YouTubeKey string
	Roles      []string
}

func initConfig(c *Config) {
	if c == nil {
		return
	}

	c.BotToken = os.Getenv("BOT_TOKEN")
	c.YouTubeKey = os.Getenv("YOUTUBE_KEY")
	c.Roles = []string{
		"music",
		"music lover",
	}
}

func (c Config) youtubeClient() (*youtube.Service, error) {
	httpClient := &http.Client{
		Transport: &transport.APIKey{Key: c.YouTubeKey},
	}

	return youtube.New(httpClient)
}

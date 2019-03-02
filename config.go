package main

import (
	"net/http"
	"os"
	"strings"

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

	// default allowed roles
	c.Roles = []string{
		"music",
		"music lover",
	}

	// additional read in roles
	input := strings.TrimSpace(os.Getenv("ALLOWED_ROLES"))
	roles := strings.Split(input, ",")

	for _, role := range roles {
		c.Roles = append(c.Roles, role)
	}
}

func (c Config) youtubeClient() (*youtube.Service, error) {
	httpClient := &http.Client{
		Transport: &transport.APIKey{Key: c.YouTubeKey},
	}

	return youtube.New(httpClient)
}

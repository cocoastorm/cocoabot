package main

import "os"

type Config struct {
	BotToken string
	Roles    []string
}

func initConfig(c *Config) {
	c.BotToken = os.Getenv("BOT_TOKEN")
	c.Roles = []string{
		"music",
		"musiclover",
	}
}

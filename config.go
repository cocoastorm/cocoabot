package main

import "os"

type Config struct {
	BotToken string
}

func initConfig(c *Config) {
	c = &Config{
		BotToken: os.Getenv("BOT_TOKEN"),
	}
}

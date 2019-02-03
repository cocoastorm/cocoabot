package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

var (
	token  string
	config Config
)

func init() {
	flag.StringVar(&token, "token", "", "Bot Token")
	flag.StringVar(&token, "t", "", "Bot Token (shorthand)")

	flag.Parse()
}

// This function will be called when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateStatus(0, "!cocoabot")
}

// This function will be called every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Check if the message is one of our commands
	if !strings.HasPrefix(m.Content, "!") {
		return
	}

	if strings.Contains(m.Content, "!decide") {
		decision := Decide(strings.TrimPrefix(m.Content, "!decide"))

		_, err := s.ChannelMessageSend(m.ChannelID, decision)
		if err != nil {
			log.Println(err)
		}
	}

	music := []string{
		"summon",
		"disconnect",
		"play",
		"skip",
		"stop",
	}

	for _, cmd := range music {
		if strings.Contains(m.Content, cmd) {
			err := musicHandler(s, m)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func main() {
	fmt.Println("cocoabot v1.0.0")

	initConfig(&config)

	if config.BotToken == "" && token == "" {
		fmt.Println("No token provided. Please run: cocoabot -t <bot token>")
		return
	}

	if token != "" {
		config.BotToken = token
	}

	dg, err := discordgo.New("Bot " + config.BotToken)
	if err != nil {
		log.Fatal(err)
	}

	// add handlers
	dg.AddHandler(ready)
	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		err = errors.Wrap(err, "Failed opening Discord session")
		log.Println(err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("cocoabot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// kill all active discord sessions
	for _, client := range clients {
		client.disconnect()
	}

	dg.Close()
}

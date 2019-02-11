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

// This function checks if the user is allowed to use the bot
// By cross referencing their guild roles with the configured allowed roles.
func isAllowed(session *discordgo.Session, guildID, userID string) bool {
	d := discord{session}

	for _, allowedRole := range config.Roles {
		if ok := d.hasRole(allowedRole, guildID, userID); ok {
			return true
		}
	}

	return false
}

// This function will be called every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message is one of our commands
	if !strings.HasPrefix(m.Content, "!") {
		return
	}

	d := discord{s}
	guild, _, err := d.getMessageOrigin(m)
	if err != nil {
		return
	}

	// check if the user is allowed to use the bot
	isServerOwner := guild.OwnerID == m.Author.ID
	hasAllowedRole := isAllowed(s, guild.ID, m.Author.ID)

	if !isServerOwner && !hasAllowedRole {
		msg := msgUserNotAllowed(m.Author)
		if _, err := s.ChannelMessageSend(m.ChannelID, msg); err != nil {
			log.Println(err)
		}

		log.Printf("%s is not allowed to use bot", m.Author.Username)
		return
	}

	// check if the message is the "decide" command
	if strings.Contains(m.Content, "!decide") {
		decision := Decide(strings.TrimPrefix(m.Content, "!decide"))

		_, err := s.ChannelMessageSend(m.ChannelID, decision)
		if err != nil {
			log.Println(err)
		}
	}

	// check if the message is one of our music commands
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

	// check if discord bot token exists
	if config.BotToken == "" && token == "" {
		fmt.Println("No token provided. Please run: cocoabot -t <bot token>")
		return
	}

	// check if token was given as a cmd argument instead
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
		err = errors.Wrap(err, "failed opening Discord session")
		log.Println(err)
	}

	defer dg.Close()

	// wait here until CTRL-C or other term signal is received.
	fmt.Println("cocoabot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// kill all active discord sessions
	for _, client := range clients {
		client.Disconnect()
	}
}

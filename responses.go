package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func msgUserNotAllowed(user *discordgo.User) string {
	msg := fmt.Sprintf("Sorry %s, looks like you don't have permissions. Please ask a server moderator.", user.Mention())

	return msg
}

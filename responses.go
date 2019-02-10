package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func msgUserNotAllowed(user *discordgo.User) string {
	msg := fmt.Sprintf("Sorry %s, looks like you don't have permissions. Please ask a server moderator.", user.Mention())

	return msg
}

func msgVoiceJoinFail(user *discordgo.User) string {
	msg := fmt.Sprintf("Sorry %s, failed to join the voice channel you're in. Am I allowed?", user)

	return msg
}

func msgQueueVideo(videoTitle string) string {
	msg := fmt.Sprintf("Got it! Adding \"%s\" to the queue.", videoTitle)

	return msg
}

func msgQueueVideoFail(originURL string) string {
	msg := fmt.Sprintf("Sorry! Failed to parse your link %s\n", originURL)

	return msg
}

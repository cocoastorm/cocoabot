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
	msg := fmt.Sprintf("Sorry %s, failed to join the voice channel you're in. Am I allowed?", user.Mention())

	return msg
}

func msgQueueVideo(videoTitle string) string {
	msg := fmt.Sprintf("Got it! Adding \"%s\" to the queue.", videoTitle)

	return msg
}

func msgNowPlaying(videoTitle string, user *discordgo.User) string {
	msg := fmt.Sprintf("Now Playing~! \"%s\" from %s", videoTitle, user.Mention())

	return msg
}

func msgNowPlayingAnon(videoTitle string) string {
	msg := fmt.Sprintf("Now Playing~! \"%s\"", videoTitle)

	return msg
}

func msgQueueVideoFail(originURL string) string {
	msg := fmt.Sprintf("Sorry! Failed to parse your link %s\n", originURL)

	return msg
}

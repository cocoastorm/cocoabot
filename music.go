package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

var clients = make(map[string]*VoiceClient)

func find(guildId string) (*VoiceClient, error) {
	client, ok := clients[guildId]
	if !ok {
		return nil, fmt.Errorf("failed to find voice client with guild id %s\n", guildId)
	}

	return client, nil
}

func findOrCreate(s *discordgo.Session, guildId, channelId string) (*VoiceClient, error) {
	client, err := find(guildId)
	if err != nil {
		client = newVoiceClient(s)
		clients[guildId] = client
	}

	err = client.connectVoice(guildId, channelId)

	return client, err
}

func stripMessage(prefix, msg string) string {
	msg = strings.TrimPrefix(msg, prefix)
	msg = strings.TrimSpace(msg)

	return msg
}

func musicHandler(s *discordgo.Session, m *discordgo.MessageCreate) error {
	guildId, channelId, err := derefMessageOrigin(s, m)

	if err != nil {
		return errors.Wrap(err, "failed to find origin of command")
	}

	if strings.Contains(m.Content, "summon") {
		guildId, channelId, err = derefUserVoiceChannel(s, m)
		_, err := findOrCreate(s, guildId, channelId)

		if err != nil {
			return errors.Wrap(err, "failed to connect to voice channel")
		}
	}

	if strings.Contains(m.Content, "disconnect") {
		client, ok := clients[guildId]
		if !ok {
			return nil
		}

		// cleanup
		client.disconnect()
		delete(clients, guildId)
	}

	if strings.Contains(m.Content, "play") {
		client, err := find(guildId)
		if err != nil {
			return errors.Wrap(err, "failed to add song to queue")
		}

		client.QueueVideo(stripMessage("!play", m.Content))

		if !client.stop {
			client.processQueue()
		}
	}

	if strings.Contains(m.Content, "resume") {
		client, err := find(guildId)
		if err != nil {
			return nil
		}

		client.ResumeVideo()
	}

	if strings.Contains(m.Content, "stop") {
		client, err := find(guildId)
		if err != nil {
			return nil
		}

		client.StopVideo()
	}

	if strings.Contains(m.Content, "skip") {
		client, err := find(guildId)
		if err != nil {
			return nil
		}

		client.SkipVideo()
	}

	return nil
}

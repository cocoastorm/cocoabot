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

func findOrCreate(d *discord, guild *discordgo.Guild, channel *discordgo.Channel) (*VoiceClient, error) {
	client, err := find(guild.ID)
	if err != nil {
		client = newVoiceClient(d)
		clients[guild.ID] = client
	}

	err = client.connectVoice(guild.ID, channel.ID)

	return client, err
}

func stripMessage(prefix, msg string) string {
	msg = strings.TrimPrefix(msg, prefix)
	msg = strings.TrimSpace(msg)

	return msg
}

func musicHandler(s *discordgo.Session, m *discordgo.MessageCreate) error {
	discord := &discord{s}
	guild, channel, err := discord.getMessageOrigin(m)

	if err != nil {
		return errors.Wrap(err, "failed to find origin of command")
	}

	if strings.Contains(m.Content, "summon") {
		guild, channel, err = discord.getUserVoiceChannel(m)
		_, err := findOrCreate(discord, guild, channel)

		if err != nil {
			return errors.Wrap(err, "failed to connect to voice channel")
		}
	}

	if strings.Contains(m.Content, "disconnect") {
		client, ok := clients[guild.ID]
		if !ok {
			return nil
		}

		// cleanup
		client.disconnect()
		delete(clients, guild.ID)
	}

	if strings.Contains(m.Content, "play") {
		client, err := find(guild.ID)
		if err != nil {
			return errors.Wrap(err, "failed to add song to queue")
		}

		client.QueueVideo(stripMessage("!play", m.Content))
	}

	if strings.Contains(m.Content, "resume") {
		client, err := find(guild.ID)
		if err != nil {
			return nil
		}

		client.ResumeVideo()
	}

	if strings.Contains(m.Content, "stop") {
		client, err := find(guild.ID)
		if err != nil {
			return nil
		}

		client.StopVideo()
	}

	if strings.Contains(m.Content, "skip") {
		client, err := find(guild.ID)
		if err != nil {
			return nil
		}

		client.SkipVideo()
	}

	return nil
}

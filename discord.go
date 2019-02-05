package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func derefMessageOrigin(s *discordgo.Session, m *discordgo.MessageCreate) (string, string, error) {
	var (
		err     error
		channel *discordgo.Channel
		guild   *discordgo.Guild
	)

	channel, err = s.State.Channel(m.ChannelID)
	if err != nil {
		return "", "", err
	}

	guild, err = s.State.Guild(channel.GuildID)
	if err != nil {
		return "", "", err
	}

	return guild.ID, channel.ID, nil
}

func derefUserVoiceChannel(s *discordgo.Session, m *discordgo.MessageCreate) (string, string, error) {
	var (
		err     error
		channel *discordgo.Channel
		guild   *discordgo.Guild
	)

	channel, err = s.State.Channel(m.ChannelID)
	if err != nil {
		return "", "", err
	}

	guild, err = s.State.Guild(channel.GuildID)
	if err != nil {
		return "", "", err
	}

	for _, vs := range guild.VoiceStates {
		if vs.UserID == m.Author.ID {
			return guild.ID, vs.ChannelID, nil
		}
	}

	return "", "", errors.New("user is not in a voice channel")
}

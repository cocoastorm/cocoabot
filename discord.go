package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

type discord struct {
	*discordgo.Session
}

func (d *discord) getMessageOrigin(msg *discordgo.MessageCreate) (*discordgo.Guild, *discordgo.Channel, error) {
	channel, err := d.Session.State.Channel(msg.ChannelID)

	if err != nil {
		return nil, nil, err
	}

	guild, err := d.Session.State.Guild(channel.GuildID)

	return guild, channel, err
}

func (d *discord) getUserVoiceChannel(msg *discordgo.MessageCreate) (*discordgo.Guild, *discordgo.Channel, error) {
	if guild, channel, err := d.getMessageOrigin(msg); err == nil {
		for _, vs := range guild.VoiceStates {
			if vs.UserID == msg.Author.ID {
				channel, err = d.Session.Channel(vs.ChannelID)
				return guild, channel, err
			}
		}
	} else {
		return nil, nil, err
	}

	return nil, nil, errors.New("user is not in any voice channel")
}

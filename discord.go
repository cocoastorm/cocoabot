package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

type discord struct {
	*discordgo.Session
}

func (d *discord) getGuildUserRoles(guildID, userID string) ([]*discordgo.Role, error) {
	var roles []*discordgo.Role

	member, err := d.Session.GuildMember(guildID, userID)
	if err != nil {
		return roles, err
	}

	guild, err := d.Session.Guild(guildID)
	if err != nil {
		return roles, err
	}

	for _, guildRole := range guild.Roles {
		for _, roleID := range member.Roles {
			if guildRole.ID == roleID {
				roles = append(roles, guildRole)
			}
		}
	}

	return roles, err
}

func (d *discord) hasRole(role, guildID, userID string) bool {
	if roles, err := d.getGuildUserRoles(guildID, userID); err == nil {
		for _, r := range roles {
			if strings.ToLower(r.Name) == strings.ToLower(role) {
				return true
			}
		}
	}

	return false
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

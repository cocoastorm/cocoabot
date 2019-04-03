package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

var clients = make(map[string]*VoiceClient)

// SongRequest represents a play query including metainfo of what user and channel.
type SongRequest struct {
	SongQuery string
	Title     string
	ChannelId string
	UserId    string
}

func find(guildId string) (*VoiceClient, error) {
	client, ok := clients[guildId]
	if !ok {
		return nil, fmt.Errorf("failed to find voice client with guild id %s", guildId)
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

	// !summon
	if strings.Contains(m.Content, "summon") {
		guild, channel, err = discord.getUserVoiceChannel(m)
		_, err := findOrCreate(discord, guild, channel)

		if err != nil {
			msg := msgVoiceJoinFail(m.Author)
			if _, err := s.ChannelMessageSend(m.ChannelID, msg); err != nil {
				return err
			}

			return errors.Wrap(err, "failed to connect to voice channel")
		}
	}

	// !disconnect
	if strings.Contains(m.Content, "disconnect") {
		client, ok := clients[guild.ID]
		if !ok {
			return nil
		}

		// cleanup
		client.Disconnect()
		delete(clients, guild.ID)
	}

	// !play
	if strings.Contains(m.Content, "play") {
		re := regexp.MustCompile(`^!play+\s(.*)`)
		if matchPlay := re.MatchString(strings.TrimSpace(m.Content)); !matchPlay {
			return fmt.Errorf("invalid command: %s", m.Content)
		}

		guild, channel, err = discord.getUserVoiceChannel(m)
		client, err := findOrCreate(discord, guild, channel)
		if err != nil {
			return errors.Wrap(err, "failed to add song to queue")
		}

		originURL := stripMessage("!play", m.Content)

		songRequest := SongRequest{
			SongQuery: originURL,
			UserId:    m.Author.ID,
			ChannelId: m.ChannelID,
		}

		titles, err := client.PlayQuery(songRequest)

		if err != nil {
			msg := msgQueueVideoFail(originURL)

			if _, err := s.ChannelMessageSend(m.ChannelID, msg); err != nil {
				return err
			}

			return errors.Wrap(err, "failed to add song(s) to queue")
		}

		for _, title := range titles {
			msg := msgQueueVideo(title)
			if _, err := s.ChannelMessageSend(m.ChannelID, msg); err != nil {
				log.Println(err)
			}
		}

		return nil
	}

	// !stop
	if strings.Contains(m.Content, "stop") {
		client, err := find(guild.ID)
		if err != nil {
			return err
		}

		client.StopVideo()
	}

	// !skip
	if strings.Contains(m.Content, "skip") {
		client, err := find(guild.ID)
		if err != nil {
			return err
		}

		client.SkipVideo()
	}

	return nil
}

package main

import (
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"layeh.com/gopus"
)

var (
	sendpcm bool
	recv    chan *discordgo.Packet
	mu      sync.Mutex
)

// max size of opus data
const maxBytes int = (frameSize * 2) * 2

// SendPCM will receive on the provied channel encode
// Received PCM data into Opus then send that to Discord
func SendPCM(v *discordgo.VoiceConnection, pcm <-chan []int16) {
	// prevent any other process from sending data
	mu.Lock()
	defer mu.Unlock()

	// pcm data has been consumed/finished
	// unlock the mutex
	if sendpcm || pcm == nil {
		return
	}

	defer func() {
		sendpcm = false
	}()

	encoder, err := gopus.NewEncoder(frameRate, channels, gopus.Audio)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		recv, ok := <-pcm
		if !ok {
			log.Println("PCM Channel is closed.")
			return
		}

		// try encoding pcm frame with opus
		opus, err := encoder.Encode(recv, frameSize, maxBytes)
		if err != nil {
			err = errors.Wrap(err, "Encoding Failed")
			log.Println(err)
			return
		}

		if v.Ready == false || v.OpusSend == nil {
			return
		}

		v.OpusSend <- opus
	}
}

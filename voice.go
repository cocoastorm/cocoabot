package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/oleiade/lane"
	"github.com/rylio/ytdl"
)

const (
	channels  int = 2
	frameRate int = 48000
	frameSize int = 960
)

type VoiceClient struct {
	discord    *discord
	voice      *discordgo.VoiceConnection
	history    *lane.Queue
	queue      *lane.Queue
	pcmChannel chan []int16
	serverId   string
	skip       bool
	stop       bool
	isPlaying  bool
}

func getYoutubeDownloadLink(url string) (*url.URL, error) {
	vid, err := ytdl.GetVideoInfo(url)
	if err != nil {
		return nil, err
	}

	vid.Formats.Sort(ytdl.FormatAudioEncodingKey, true)

	link, yterr := vid.GetDownloadURL(vid.Formats[0])

	return link, yterr
}

func newVoiceClient(d *discord) *VoiceClient {
	return &VoiceClient{
		discord:    d,
		history:    lane.NewQueue(),
		queue:      lane.NewQueue(),
		pcmChannel: make(chan []int16, 2),
	}
}

func (vc *VoiceClient) connectVoice(guildId, channelId string) error {
	voice, err := vc.discord.ChannelVoiceJoin(guildId, channelId, false, false)
	if err != nil {
		return err
	}

	vc.voice = voice

	go SendPCM(vc.voice, vc.pcmChannel)

	return nil
}

func (vc *VoiceClient) disconnect() {
	close(vc.pcmChannel)

	if vc.voice != nil {
		vc.voice.Close()
	}
}

func (vc *VoiceClient) ResumeVideo() {
	vc.stop = false

	link := vc.history.Dequeue()
	if link != nil {
		vc.playVideo(link.(string))
	}
}

func (vc *VoiceClient) StopVideo() {
	vc.stop = true
}

func (vc *VoiceClient) SkipVideo() {
	vc.skip = true
}

func (vc *VoiceClient) QueueVideo(youtubeLink string) {
	fmt.Printf("Queuing Video: %s\n", youtubeLink)

	link, err := getYoutubeDownloadLink(youtubeLink)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("[youtube] %s\n", youtubeLink)

	vc.queue.Enqueue(link.String())
	vc.processQueue()
}

func (vc *VoiceClient) playVideo(url string) {
	vc.isPlaying = true

	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("status non-200")
	}

	run := exec.Command("ffmpeg", "-i", "-", "-f", "s16le", "-ar", strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "pipe:1")
	run.Stdin = resp.Body

	stdout, err := run.StdoutPipe()
	if err != nil {
		fmt.Printf("ffmpeg failed to pipe out: %s\n", err.Error())
		return
	}

	err = run.Start()
	if err != nil {
		fmt.Printf("ffmpeg failed to start: %s\n", err.Error())
		return
	}

	audiobuf := make([]int16, frameSize*channels)

	vc.voice.Speaking(true)
	defer vc.voice.Speaking(false)

	for {
		// read data from ffmpeg
		err = binary.Read(stdout, binary.LittleEndian, &audiobuf)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			log.Println("oops, failed playing", err)
			break
		}

		if err != nil {
			log.Println("oops, failed playing", err)
			break
		}

		if vc.stop == true || vc.skip == true {
			run.Process.Kill()
			log.Println("stopped playing")
			break
		}

		vc.pcmChannel <- audiobuf
	}

	vc.isPlaying = false
}

func (vc *VoiceClient) processQueue() {
	if vc.isPlaying {
		return
	}

	for {
		vc.skip = false
		if link := vc.queue.Dequeue(); link != nil && !vc.stop {
			vc.history.Enqueue(link.(string))
			vc.playVideo(link.(string))
		} else {
			break
		}
	}
}

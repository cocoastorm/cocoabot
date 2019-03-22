package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/khoanguyen96/cocoabot/audio"
	"github.com/oleiade/lane"
	"github.com/pkg/errors"
	"github.com/rylio/ytdl"
)

const (
	channels  int = 2
	frameRate int = 48000
	frameSize int = 960

	userAgent string = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36"
)

type VoiceClient struct {
	discord    *discord
	voice      *discordgo.VoiceConnection
	queue      *lane.Queue
	mutex      sync.Mutex
	pcmChannel chan []int16
	serverId   string
	skip       bool
	stop       bool
	isPlaying  bool
}

func newVoiceClient(d *discord) *VoiceClient {
	return &VoiceClient{
		discord:    d,
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

func (vc *VoiceClient) Disconnect() {
	if vc.isPlaying {
		vc.StopVideo()

		// wait a little bit~
		time.Sleep(250 * time.Millisecond)
	}

	close(vc.pcmChannel)

	if vc.voice != nil {
		vc.voice.Disconnect()
	}
}

func (vc *VoiceClient) StopVideo() {
	vc.stop = true
}

func (vc *VoiceClient) SkipVideo() {
	vc.skip = true
}

func (vc *VoiceClient) queueVideo(sq SongRequest) {
	vc.queue.Enqueue(sq)
	go vc.processQueue()
}

func (vc *VoiceClient) PlayQuery(query SongRequest) ([]string, error) {
	// if the query is a youtube playlist link
	// fetch the videos from the youtube api
	if playlistId, err := getYouTubePlayListIdFromURL(query.SongQuery); err == nil {
		videos, err := playlistVideos(playlistId)
		if err != nil {
			return []string{}, err
		}

		return vc.playYoutubeList(videos, query)
	}

	if !isYouTubeLink(query.SongQuery) {
		// check if an API Key was configured
		// if it isn't searching can't be done, so quit early
		if config.YouTubeKey == "" {
			return []string{}, errors.New("youtube searching has not been configured, needs API key")
		}

		// if these are just words to search for
		// search with the youtube api
		resp, err := searchByKeywords(query.SongQuery)
		if err != nil {
			return []string{}, err
		}

		query.SongQuery = resp.VideoId
	}

	// if its just a regular youtube link
	// pass it along
	title, err := vc.playYoutubeWithId(query)
	return []string{title}, err
}

func (vc *VoiceClient) playYoutubeWithId(s SongRequest) (string, error) {
	info, err := ytdl.GetVideoInfo(s.SongQuery)
	if err != nil {
		return "", err
	}

	fmt.Printf("Queuing Video: %s [%s]\n", info.Title, s.SongQuery)

	audioLink, err := getSortYouTubeAudioLink(info)
	if err != nil {
		return "", err
	}

	s.Title = strings.TrimSpace(info.Title)
	s.SongQuery = audioLink.String()

	vc.queueVideo(s)

	return info.Title, nil
}

func (vc *VoiceClient) playYoutubeList(videos []string, sr SongRequest) ([]string, error) {
	var titleVideos []string

	for _, video := range videos {
		request := SongRequest{
			SongQuery: video,
			ChannelId: sr.ChannelId,
			UserId:    sr.UserId,
		}

		title, err := vc.playYoutubeWithId(request)
		if err != nil {
			log.Println(err)
			continue
		}

		titleVideos = append(titleVideos, title)
	}

	return titleVideos, nil
}

// func (vc *VoiceClient) playVideo(url string) {
// 	vc.isPlaying = true

// 	// pass music stream url to ffmpeg
// 	run := exec.Command("ffmpeg", "-i", url, "-headers", fmt.Sprintf("User-Agent: %s", userAgent), "-acodec", "pcm_s16le", "-f", "s16le", "-ar", strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "pipe:1")

// 	stdout, err := run.StdoutPipe()
// 	if err != nil {
// 		fmt.Printf("ffmpeg failed to pipe out: %s\n", err.Error())
// 		return
// 	}

// 	ffmpegbuf := bufio.NewReaderSize(stdout, 16384)

// 	err = run.Start()
// 	if err != nil {
// 		fmt.Printf("ffmpeg failed to start: %s\n", err.Error())
// 		return
// 	}

// 	defer run.Process.Kill()

// 	audiobuf := make([]int16, frameSize*channels)

// 	vc.voice.Speaking(true)
// 	defer vc.voice.Speaking(false)

// 	for {
// 		// read data from ffmpeg
// 		err = binary.Read(ffmpegbuf, binary.LittleEndian, &audiobuf)
// 		if err == io.EOF {
// 			log.Println("oops, encountered the end too early", err)
// 			break
// 		}

// 		if err == io.ErrUnexpectedEOF {
// 			log.Println("oops, connection was closed", err)
// 			break
// 		}

// 		if err != nil {
// 			log.Println("oops, failed playing", err)
// 			break
// 		}

// 		if vc.stop == true || vc.skip == true {
// 			log.Println("stopped playing")
// 			break
// 		}

// 		vc.pcmChannel <- audiobuf
// 	}

// 	vc.isPlaying = false
// }

func (vc *VoiceClient) playVideo(url string) {
	vc.isPlaying = true

	encoding := audio.Encode(url, audio.WithDefaults())

	vc.voice.Speaking(true)

	defer func() {
		vc.isPlaying = false
		vc.voice.Speaking(false)
	}()

	for {
		frame, err := encoding.OpusFrame()
		if err != nil {
			// log.Println(err)
			log.Fatal(err)
			break
		}

		if vc.stop || vc.skip {
			log.Println("stopped")
			break
		}

		vc.voice.OpusSend <- frame
	}
}

func (vc *VoiceClient) NowPlaying(sr SongRequest) {
	var msg string

	user, err := vc.discord.User(sr.UserId)
	if err != nil {
		msg = msgNowPlayingAnon(sr.Title)
	} else {
		msg = msgNowPlaying(sr.Title, user)
	}

	if _, err := vc.discord.ChannelMessageSend(sr.ChannelId, msg); err != nil {
		log.Println(err)
	}

	log.Println(msg)
}

func (vc *VoiceClient) processQueue() {
	// if music is currently playing
	// exit early, as another goroutine is (most likely) accessing the queue
	if vc.isPlaying {
		return
	}

	// if !stop was used sometime ago
	// reset it
	if vc.stop {
		vc.stop = false
	}

	// strictly allow one goroutine to dequeue
	vc.mutex.Lock()
	defer vc.mutex.Unlock()

	for {
		if songRequest := vc.queue.Dequeue(); songRequest != nil && !vc.stop {
			sr := songRequest.(SongRequest)

			// send a message that the next song is playing
			// to the original user who requested the song
			vc.NowPlaying(sr)

			// NOTE: this should be blocking
			// as we don't want multiple ffmpeg instances running for every song
			// in the damn queue
			vc.playVideo(sr.SongQuery)
		} else {
			break
		}

		vc.skip = false
	}
}

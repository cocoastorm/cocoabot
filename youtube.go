package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/rylio/ytdl"
)

type YouTubeResult struct {
	Title   string
	VideoId string
}

func isYouTubeLink(link string) bool {
	if strings.Contains(link, "youtu") || strings.ContainsAny(link, "\"?&/<%=") {
		matchers := []*regexp.Regexp{
			regexp.MustCompile(`(?:v|embed|watch\?v)(?:=|/)([^"&?/=%]{11})`),
			regexp.MustCompile(`(?:=|/)([^"&?/=%]{11})`),
			regexp.MustCompile(`([^"&?/=%]{11})`),
		}

		for _, re := range matchers {
			if isMatch := re.MatchString(link); isMatch {
				return true
			}
		}
	}

	return false
}

func searchByKeywords(keywords string) (YouTubeResult, error) {
	// get youtube client
	service, err := config.youtubeClient()
	if err != nil {
		err = errors.Wrap(err, "failed configuring youtube client")
		return YouTubeResult{}, err
	}

	// search for the youtube videos
	call := service.Search.List("snippet")
	call = call.Type("video")
	call = call.Q(keywords)

	response, err := call.Do()
	if err != nil {
		err = errors.Wrap(err, "failed getting results from youtube")
		return YouTubeResult{}, err
	}

	// TODO: better selection
	// for now return the first result
	item := response.Items[0]

	return YouTubeResult{
		Title:   item.Snippet.Title,
		VideoId: item.Id.VideoId,
	}, nil
}

func playlistVideos(playlistId string) ([]string, error) {
	var videos []string

	service, err := config.youtubeClient()
	if err != nil {
		err = errors.Wrap(err, "failed configuring youtube client")
		return videos, err
	}

	// fetch playlist items
	call := service.PlaylistItems.List("snippet")
	call = call.PlaylistId(playlistId)

	response, err := call.Do()
	if err != nil {
		err = errors.Wrap(err, "failed getting playlist videos from youtube")
		return videos, err
	}

	for _, item := range response.Items {
		videoId := item.Snippet.ResourceId.VideoId

		if videoId != "" {
			videos = append(videos, item.Snippet.ResourceId.VideoId)
		}
	}

	return videos, nil
}

func getYouTubePlayListIdFromURL(link string) (string, error) {
	if !strings.Contains(link, "youtu") && !strings.Contains(link, "playlist") {
		return "", fmt.Errorf("%s is not a valid youtube playlist URL", link)
	}

	playlistURL, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	values := playlistURL.Query()
	playlistId := values.Get("list")

	if playlistId == "" {
		return "", fmt.Errorf("%s has no valid youtube playlist Id", playlistURL.String())
	}

	return playlistId, nil
}

func getSortYouTubeAudioLink(info *ytdl.VideoInfo) (*url.URL, error) {
	if len(info.Formats) == 0 {
		return &url.URL{}, errors.New("failed to get info from youtube")
	}

	info.Formats.Sort(ytdl.FormatAudioEncodingKey, true)

	// TODO: better selection of which format/stream
	audioFormat := info.Formats[0]
	link, err := info.GetDownloadURL(audioFormat)

	return link, err
}

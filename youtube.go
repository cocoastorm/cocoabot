package main

import "github.com/pkg/errors"

type YouTubeResult struct {
	Title   string
	VideoId string
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

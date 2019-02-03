package main

import (
	"fmt"

	"github.com/rylio/ytdl"
)

func main() {
	vid, _ := ytdl.GetVideoInfo("https://www.youtube.com/watch?v=IG1U2FJOWXs")
	vid.Formats.Sort(ytdl.FormatAudioEncodingKey, true)
	link, _ := vid.GetDownloadURL(vid.Formats[0])

	fmt.Println(link)
}

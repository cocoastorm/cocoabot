package audio

import (
	"fmt"
	"strconv"
)

type AudioOptions struct {
	Bitrate           int
	Channels          int
	CompressionLevel  int
	FrameRate         int
	FrameDuration     int
	PacketLoss        int
	BufferedFrames    int
	VBR               bool
	WithSpoofedHeader bool
}

const userAgent string = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36"

func (opts *AudioOptions) PCMFrameLen() int {
	return 960 * opts.Channels * (opts.FrameDuration / 20)
}

func (opts *AudioOptions) FFmpegArgs() []string {
	args := []string{
		"-b:a", strconv.Itoa(opts.Bitrate * 1000),
		"-ac", strconv.Itoa(opts.Channels),
		"-compression_level", strconv.Itoa(opts.CompressionLevel),
		"-ar", strconv.Itoa(opts.FrameRate),
		"-frame_duration", strconv.Itoa(opts.FrameDuration),
		"-packet_loss", strconv.Itoa(opts.PacketLoss),
	}

	if opts.VBR {
		args = append(args, []string{"-vbr", "on"}...)
	} else {
		args = append(args, []string{"-vbr", "off"}...)
	}

	if opts.WithSpoofedHeader {
		args = append(args, []string{
			"-headers",
			fmt.Sprintf("User-Agent: %s", userAgent),
		}...)
	}

	return args
}

func WithDefaults() *AudioOptions {
	return &AudioOptions{
		Bitrate:           64,
		Channels:          2,
		CompressionLevel:  10,
		FrameRate:         48000,
		FrameDuration:     20,
		PacketLoss:        1,
		BufferedFrames:    100,
		VBR:               true,
		WithSpoofedHeader: true,
	}
}

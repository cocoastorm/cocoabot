package audio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/jonas747/ogg"
)

type Frame []byte

type Encoding struct {
	mu            sync.Mutex
	options       *AudioOptions
	input         string
	frameChannel  chan Frame
	ffmpegProcess *os.Process
	running       bool
	stop          chan bool
	lastframe     int
}

func (e *Encoding) encodeArgs() []string {
	args := []string{
		"-i", e.input,
		"-map", "0:a",
		"-acodec", "libopus",
		"-f", "ogg",
	}

	encodeArgs := e.options.FFmpegArgs()

	args = append(args, encodeArgs...)
	args = append(args, "pipe:1")

	return args
}

func (e *Encoding) start() {
	defer func() {
		e.running = false
		e.mu.Unlock()
	}()

	args := e.encodeArgs()
	run := exec.Command("ffmpeg", args...)

	log.Printf("ffmpeg args: %v", args)

	// ready? set? go!
	e.mu.Lock()
	e.running = true

	stdout, err := run.StdoutPipe()
	if err != nil {
		fmt.Printf("FFmpeg failed to pipe out: %s\n", err.Error())
		return
	}

	// debug stderr
	// stderr, err := run.StderrPipe()
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// bufstderr := new(bytes.Buffer)

	// ffmpegbuf := bufio.NewReaderSize(stdout, 16384)

	err = run.Start()
	defer run.Process.Kill()

	if err != nil {
		fmt.Printf("ffmpeg failed to start: %s\n", err.Error())
		return
	}

	// debug ffmpeg stderr
	// bufstderr.ReadFrom(stderr)
	// fmt.Println(bufstderr.String())

	decoder := ogg.NewPacketDecoder(ogg.NewDecoder(stdout))
	skip := 2

	for {
		audiobuf := new(bytes.Buffer)
		packet, _, err := decoder.Decode()

		if err != nil {
			log.Println(err)
			break
		}

		// skip the ogg metadata packets
		if skip > 0 {
			fmt.Printf("skipping packet: %d\n", skip)
			skip--
			continue
		}

		err = binary.Write(audiobuf, binary.LittleEndian, int16(len(packet)))
		if err != nil {
			log.Println(err)
			break
		}

		_, err = audiobuf.Write(packet)
		if err != nil {
			log.Println(err)
			break
		}

		e.frameChannel <- Frame(audiobuf.Bytes())

		// select {
		// case <-e.stop:
		// 	break
		// default:
		// 	continue
		// }

		// e.mu.Lock()
		// e.lastframe++
		// e.mu.Unlock()
	}

	// err = run.Wait()
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
}

func Encode(input string, opts *AudioOptions) *Encoding {
	e := &Encoding{
		input:        input,
		options:      opts,
		frameChannel: make(chan Frame, opts.BufferedFrames),
		stop:         make(chan bool),
	}

	go e.start()

	return e
}

func (e *Encoding) OpusFrame() (frame []byte, err error) {
	f := <-e.frameChannel
	if f == nil {
		return nil, io.EOF
	}

	if len(f) < 2 {
		return nil, errors.New("Bad Frame")
	}

	return f[2:], nil
}

func (e *Encoding) Stop() {
	e.stop <- true
}

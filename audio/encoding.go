package audio

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/jonas747/ogg"
)

type Frame []byte

type Encoding struct {
	mu            sync.Mutex
	options       AudioOptions
	input         string
	frameChannel  chan Frame
	ffmpegProcess *os.Process
	lastframe     int
	running       bool
}

func (e *Encoding) encodeArgs() []string {
	args := e.options.FFmpegArgs()
	args = append(args, []string{
		"-i", e.input,
		"-map", "0:a",
		"-acodec", "libopus",
		"-f", "ogg",
		"-pipe:1",
	}...)

	return args
}

func (e *Encoding) start() {
	e.mu.Lock()

	defer func() {
		e.running = false
		e.mu.Unlock()
	}()

	run := exec.Command("ffmpeg", e.encodeArgs()...)

	stdout, err := run.StdoutPipe()
	if err != nil {
		fmt.Printf("FFmpeg failed to pipe out: %s\n", err.Error())
		return
	}

	ffmpegbuf := bufio.NewReaderSize(stdout, 16384)

	err = run.Start()
	defer run.Process.Kill()

	if err != nil {
		fmt.Printf("ffmpeg failed to start: %s\n", err.Error())
		return
	}

	decoder := ogg.NewPacketDecoder(ogg.NewDecoder(ffmpegbuf))
	skip := 2

	for {
		buf := new(bytes.Buffer)
		packet, _, err := decoder.Decode()

		if skip > 0 {
			skip--
			continue
		}

		if err != nil {
			break
		}

		err = binary.Write(buf, binary.LittleEndian, int16(len(packet)))
		if err != nil {
			break
		}

		_, err = buf.Write(packet)
		if err != nil {
			break
		}

		e.frameChannel <- Frame(buf.Bytes())

		e.mu.Lock()
		e.lastframe++
		e.mu.Unlock()
	}
}

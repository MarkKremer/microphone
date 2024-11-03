// Package microphone provides a wrapper around the PortAudio microphone
// stream to make it compatible with the beep audio library.
package microphone

import (
	"errors"

	"github.com/gopxl/beep/v2"
	"github.com/gordonklaus/portaudio"
)

var (
	errInvalidAmountOfInputChannels = errors.New("invalid amount of inputChannels specified. microphone.OpenDefaultStream func expects exactly 2 or 1")
)

// Init initializes internal datastructures of PortAudio and
// the host APIs for use.
//
// This method exists as a convenient single point of contact
// if the client doesn't use any other PortAudio functionality.
// Otherwise, calling portaudio.Initialize() is recommended
// instead of calling this method.
func Init() error {
	return portaudio.Initialize()
}

// Terminate deallocates all resources allocated by PortAudio.
//
// This method exists as a convenient single point of contact
// if the client doesn't use any other PortAudio functionality.
// Otherwise, calling portaudio.Terminate() is recommended
// instead of calling this method.
//
// Terminate MUST be called before exiting a program which uses PortAudio.
// Failure to do so may result in serious resource leaks, such as audio devices
// not being available until the next reboot.
func Terminate() error {
	return portaudio.Terminate()
}

// bufferSize is the size of the internal buffer. It is currently
// set to be the same as the Append method of beep.Buffer which
// might make this the tiniest bit more efficient if used together.
const bufferSize = 512

// OpenDefaultStream opens the default input stream.
func OpenDefaultStream(sampleRate beep.SampleRate, inputChannels int) (s *Streamer, format beep.Format, err error) {
	if inputChannels > 2 || inputChannels == 0 {
		return nil, beep.Format{}, errInvalidAmountOfInputChannels
	}

	s = &Streamer{}
	s.buffer = make([][]float32, inputChannels)

	for i := range s.buffer {
		s.buffer[i] = make([]float32, bufferSize)
	}

	s.stream, err = portaudio.OpenDefaultStream(inputChannels, 0, float64(sampleRate), bufferSize, s.buffer)
	if err != nil {
		return nil, beep.Format{}, err
	}
	// Set the position to the end so that the Stream
	// method will fetch new data from the microphone.
	s.pos = bufferSize
	format = beep.Format{
		SampleRate:  sampleRate,
		NumChannels: inputChannels,
		// NOTE(m): I couldn't find how to obtain the actual precision
		// from the microphone. 3 bytes is the highest precision
		// supported by the beep library for saving WAV files.
		Precision: 3,
	}
	return
}

// portaudioStream is used to mock portaudio.Stream during
// testing.
type portaudioStream interface {
	Start() error
	Stop() error
	Read() error
	Close() error
}

// Streamer is an implementation of the beep.StreamCloser interface
// to provide access to the microphone through the PulseAudio library.
type Streamer struct {
	stream portaudioStream
	buffer [][]float32
	pos    int
	err    error
}

// Start commences audio processing.
func (s *Streamer) Start() error {
	return s.stream.Start()
}

// Stop terminates audio processing (but does not terminate the stream).
func (s *Streamer) Stop() error {
	return s.stream.Stop()
}

// Stream fills samples with the audio recorded with the microphone.
// Unless there is an error, this method will wait until samples
// is filled completely which may involve waiting for the OS to
// supply the data.
func (s *Streamer) Stream(samples [][2]float64) (int, bool) {
	if s.err != nil {
		return 0, false
	}

	// Fill samples with previously fetched data.
	n := bufferSize - s.pos
	if n > len(samples) {
		n = len(samples)
	}

	var i int
	for i = 0; i < n; i++ {
		samples[i] = convertBufferIntoSamples(s.buffer, s.pos+i)
	}

	if n == len(samples) {
		s.pos += n
		return len(samples), true
	}

	// Once buffer is drained, fetch new data.
	for {
		s.err = s.stream.Read()
		if s.err != nil {
			return 0, false
		}

		m := bufferSize
		if n+m > len(samples) {
			m = len(samples) - n
		}

		for i = 0; i < m; i++ {
			samples[n+i] = convertBufferIntoSamples(s.buffer, i)
		}

		n += m
		if n == len(samples) {
			s.pos = m
			return len(samples), true
		}
	}
}

// Err returns an error that occurred during streaming.
// If no error occurred, nil is returned.
func (s *Streamer) Err() error {
	return s.err
}

// Close terminates the stream.
func (s *Streamer) Close() error {
	return s.stream.Close()
}

func convertBufferIntoSamples(buffer [][]float32, bufferPos int) [2]float64 {
	var samples [2]float64

	if len(buffer) > 1 {
		samples[0] = float64(buffer[0][bufferPos])
		samples[1] = float64(buffer[1][bufferPos])

		return samples
	}

	samples[0] = float64(buffer[0][bufferPos])
	samples[1] = float64(buffer[0][bufferPos])

	return samples
}

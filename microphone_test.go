package microphone

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"testing"

	"github.com/gopxl/beep/v2/wav"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func ExampleOpenDefaultStream_recordWav() {
	if len(os.Args) < 2 {
		fmt.Println("missing required argument: output file name")
		return
	}
	fmt.Println("Recording. Press Ctrl-C to stop.")

	err := Init()
	if err != nil {
		log.Fatal(err)
	}
	defer Terminate()

	stream, format, err := OpenDefaultStream(44100, 1)
	if err != nil {
		log.Fatal(err)
	}
	// Close the stream at the end if it hasn't already been
	// closed explicitly.
	defer stream.Close()

	filename := os.Args[1]
	if !strings.HasSuffix(filename, ".wav") {
		filename += ".wav"
	}
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	// Stop the stream when the user tries to quit the program.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	go func() {
		<-sig
		stream.Stop()
		stream.Close()
	}()

	stream.Start()

	// Encode the stream. This is a blocking operation because
	// wav.Encode will try to drain the stream. However, this
	// doesn't happen until stream.Close() is called.
	err = wav.Encode(f, stream, format)
	if err != nil {
		log.Fatal(err)
	}
}

type mockNumberStreamer struct {
	buffer [][]float32
	index  int
}

func (m *mockNumberStreamer) Start() error {
	return nil
}

func (m *mockNumberStreamer) Stop() error {
	return nil
}

func (m *mockNumberStreamer) Read() error {
	// Stream a consecutive numbers.
	for i := 0; i < len(m.buffer[0]); i++ {
		val := float32(m.index + i)
		m.buffer[0][i] = val
		m.buffer[1][i] = -val
	}
	m.index += len(m.buffer[0])
	return nil
}

func (m *mockNumberStreamer) Close() error {
	return nil
}

func newTestableStreamer() (s Streamer) {
	s.buffer = make([][]float32, 2)
	s.buffer[0] = make([]float32, bufferSize)
	s.buffer[1] = make([]float32, bufferSize)
	s.pos = bufferSize
	return
}

func testStreamWithSampleSize(t *testing.T, numSamples int) {
	s := newTestableStreamer()
	s.stream = &mockNumberStreamer{
		buffer: s.buffer,
	}

	for i := 0; i < 3; i++ {
		samples := make([][2]float64, numSamples)
		n, ok := s.Stream(samples)
		assert.Equal(t, len(samples), n)
		assert.True(t, ok)

		for j, sample := range samples {
			assert.InDelta(t, float64(i*numSamples+j), sample[0], 0.001)
			assert.InDelta(t, float64(-(i*numSamples + j)), sample[1], 0.001)
		}
	}
}

func TestStreamer_Stream_withSameSampleSizeAsInternalBufferSize(t *testing.T) {
	testStreamWithSampleSize(t, bufferSize)
}

func TestStreamer_Stream_withSmallerSampleSizeAsInternalBufferSize(t *testing.T) {
	testStreamWithSampleSize(t, 100)
}

func TestStreamer_Stream_withBiggerSampleSizeAsInternalBufferSize(t *testing.T) {
	testStreamWithSampleSize(t, bufferSize+100)
}

func TestStreamer_Stream_withDoubleSampleSizeAsInternalBufferSize(t *testing.T) {
	testStreamWithSampleSize(t, bufferSize*2)
}

var ErrTest = errors.New("error for testing")

type mockErrorStreamer struct {
}

func (m *mockErrorStreamer) Start() error {
	return nil
}

func (m *mockErrorStreamer) Stop() error {
	return nil
}

func (m *mockErrorStreamer) Read() error {
	return ErrTest
}

func (m *mockErrorStreamer) Close() error {
	return nil
}

func TestStreamer_Stream_onError(t *testing.T) {
	s := newTestableStreamer()
	s.stream = &mockErrorStreamer{}

	// When an error occurs, s.Stream() must return a
	// 0, false response and the error must be accessible
	// through s.Err(). Any subsequent calls to s.Stream()
	// must return the same result. See the beep.Streamer
	// interface for more details.
	samples := make([][2]float64, 100)
	assert.Nil(t, s.Err())
	n, ok := s.Stream(samples)
	assert.Equal(t, 0, n)
	assert.False(t, ok)
	assert.Equal(t, ErrTest, s.Err())

	n, ok = s.Stream(samples)
	assert.Equal(t, 0, n)
	assert.False(t, ok)
	assert.Equal(t, ErrTest, s.Err())
}

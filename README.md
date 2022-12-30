# Microphone [![GoDoc](https://godoc.org/github.com/MarkKremer/microphone?status.svg)](https://godoc.org/github.com/MarkKremer/microphone) [![Go Report Card](https://goreportcard.com/badge/github.com/MarkKremer/microphone)](https://goreportcard.com/report/github.com/MarkKremer/microphone)

Microphone is a small library that takes [this Go PortAudio library](https://github.com/gordonklaus/portaudio)
and wraps its microphone stream in a beep.StreamCloser
so that it can be used with everything else in the [Beep library](https://github.com/faiface/beep).

```bash
go get -u github.com/MarkKremer/microphone
```

## Installation
This package requires that you have the PortAudio development headers and libraries installed.
On Ubuntu this can be done using:
```bash
apt-get install portaudio19-dev
```
See [the PortAudio library](https://github.com/gordonklaus/portaudio) for more information.

## Example

Here is a short example of creating a microphone stream and sending it to the speaker. 

```
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/MarkKremer/microphone"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

func main() {
	// First, some configuration
	sampleRate := beep.SampleRate(144100) // Choose sample rate
	numChannels := 1                      // 1 - mono, 2 - stereo

	microphone.Init() // without this you will get "PortAudio not initialized" error later

	// Create microphone stream
	micStream, format, err := microphone.OpenDefaultStream(sampleRate, numChannels)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%#v\n", format) // Sample rate and number of channels, in case you'll need it

	speaker.Init(sampleRate, sampleRate.N(time.Second/10)) // Initialize speaker, with a buffer for 0.1 sec

	micStream.Start()       // Start recording
	speaker.Play(micStream) // Start playing what you record

	select {} // Wait forever
}
```

## Licence

[MIT](https://github.com/MarkKremer/microphone/blob/master/LICENSE)

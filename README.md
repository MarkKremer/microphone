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

## Licence

[MIT](https://github.com/MarkKremer/microphone/blob/master/LICENSE)

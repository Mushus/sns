package main

import (
	"os"

	"github.com/rs/zerolog"
)

func main() {
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()
	s, err := createServer(&log)
	if err != nil {
		log.Fatal().Err(err).Send()
		return
	}
	if err := s.Start(); err != nil {
		log.Fatal().Err(err).Send()
		return
	}
}

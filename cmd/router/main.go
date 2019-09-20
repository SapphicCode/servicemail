package main

import (
	"os"

	"github.com/Pandentia/servicemail/servicemail/routing"
	"github.com/rs/zerolog"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	app := kingpin.New("router", "Routing middleware for Servicemail")

	AMQPURI := app.Flag("amqp-uri", "The AMQP URI to connect to").Envar("AMQP_URI").Short('u').Required().String()

	verbose := app.Flag("verbose", "Enables debug logging").Short('v').Bool()
	pretty := app.Flag("pretty", "Enables pretty logging").Short('p').Bool()

	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *pretty {
		logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	if *verbose {
		logger = logger.Level(zerolog.DebugLevel)
	} else {
		logger = logger.Level(zerolog.InfoLevel)
	}

	router := routing.Router{
		MQURI:  *AMQPURI,
		Logger: logger,
	}

	if err := router.New(); err != nil {
		logger.Fatal().Err(err).Msg("Error initializing router.")
	}
	if err := router.Run(); err != nil {
		logger.Fatal().Err(err).Msg("Error running router.")
	}
}

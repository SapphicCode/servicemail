package main

import (
	"os"

	"github.com/Pandentia/servicemail/servicemail/ingress/mailgun"
	"github.com/rs/zerolog"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	app := kingpin.New("ingress-mailgun", "Ingress middleware for Servicemail")

	AMQPURI := app.Flag("amqp-uri", "The AMQP URI to connect to").Envar("AMQP_URI").Short('u').Required().String()
	bind := app.Flag("bind", "The address to bind to").Default("[::]:8080").Short('b').String()

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

	ingest := &mailgun.API{
		Logger: logger,
	}

	if err := ingest.New(*AMQPURI); err != nil {
		logger.Fatal().Err(err).Msg("Error initializing.")
	}
	if err := ingest.Run(*bind); err != nil {
		logger.Fatal().Err(err).Msg("Error running ingress API.")
	}
}

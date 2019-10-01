package mailgun

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
)

// API describes this ingress API.
type API struct {
	Logger zerolog.Logger

	connnection *amqp.Connection
	channel     *amqp.Channel
}

// New initializes and connects a new API instance.
func (api *API) New(MQURI string) error {
	logger := api.Logger.With().Str("module", "initializer").Logger()

	// connect to the message broker
	conn, err := amqp.Dial(MQURI)
	if err != nil {
		return err
	}
	api.connnection = conn
	logger.Debug().Msg("Connection to message broker established")

	// create channel
	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	api.channel = channel
	logger.Debug().Msg("Channel created")

	// we have no queue or exchange fuss this time around, we're done :)
	return nil
}

// Run runs the API instance at a given bind address.
func (api *API) Run(bind string) error {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	r.POST("/mailgun", api.mailgunHandler)

	return r.Run(bind)
}

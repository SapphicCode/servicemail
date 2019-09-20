package rpc

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
	"github.com/vmihailenco/msgpack/v4"
)

// Server describes an RPC server, providing multiple RPC handlers.
type Server struct {
	Handlers map[string]func(interface{}) interface{}
	Logger   zerolog.Logger

	// Connection configuration
	Connection *amqp.Connection
	Exchange   string // Exchange is expected to be a direct or topic exchange
	RoutingKey string // routing key prefix for requests (eg. "rpc.request")

	channel *amqp.Channel
}

func (s *Server) runHandler(queueName, handlerName string) {
	logger := s.Logger.With().Str("module", "rpc-handler").Str("handler", handlerName).Logger()

	// create channel
	deliveryChannel, err := s.channel.Consume(
		queueName,
		"",
		true,  // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,
	)
	if err != nil {
		logger.Err(err).Msg("Error creating delivery channel.")
		return
	}

	// fetch handler
	handler := s.Handlers[handlerName]

	// process RPC requests
	for delivery := range deliveryChannel {
		var data interface{}

		// receive and parse data
		err := msgpack.Unmarshal(delivery.Body, &data)
		if err != nil {
			logger.Err(err).Msg("Error deserializing request body.")
			continue
		}

		// run handler
		output := handler(data)

		// serialize output
		response, err := msgpack.Marshal(output)
		if err != nil {
			logger.Err(err).Msg("Error serializing response body.")
			continue
		}

		// return output to sender
		err = s.channel.Publish(
			s.Exchange,
			delivery.ReplyTo, // use ReplyTo as routing key
			true,             // mandatory
			false,            // immediate
			amqp.Publishing{
				Body: response,
			},
		)
		if err != nil {
			logger.Err(err).Msg("Error publishing response.")
		}
	}
}

// Run runs the RPC server.
func (s *Server) Run(ctx context.Context) error {
	// init channel
	channel, err := s.Connection.Channel()
	if err != nil {
		return err
	}
	s.channel = channel

	// set prefetch
	err = channel.Qos(1, 0, false)
	if err != nil {
		return err
	}

	// spawn handler goroutines
	for handler := range s.Handlers {
		// declare RPC queue
		queue, err := channel.QueueDeclare(
			s.Exchange+"."+s.RoutingKey+"."+handler,
			false, // durable
			false, // autoDelete
			false, // exclusive
			false, // noWait
			nil,
		)
		if err != nil {
			return err
		}

		// bind RPC queue
		err = channel.QueueBind(
			queue.Name,
			s.RoutingKey+"."+handler,
			s.Exchange,
			false, // noWait
			nil,
		)
		if err != nil {
			return err
		}

		// spawn handler listener
		go s.runHandler(queue.Name, handler)
	}

	return nil
}

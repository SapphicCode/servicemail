package rpc

import (
	"errors"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
)

// Client describes an RPC client with the ability to call remote RPC servers.
type Client struct {
	Logger zerolog.Logger

	// Connection configuration.
	Connection         *amqp.Connection
	Exchange           string // Exchange to register our response queues against. Expected to be topic or direct.
	RequestRoutingKey  string // Routing key prefix for requests.
	ResponseRoutingKey string // Routing key prefix for responses (e.g. "rpc.response").

	channel *amqp.Channel
	replyTo string // assembled routing key for requests
	seq     uint64 // sequence number for request correlation
	callers map[string]chan<- []byte
}

func (c *Client) setup() error {
	// return if already set up
	if c.channel != nil {
		return nil
	}

	// set up caller map
	c.callers = make(map[string]chan<- []byte)

	// set up channel
	channel, err := c.Connection.Channel()
	if err != nil {
		return err
	}
	c.channel = channel

	// set prefetch
	err = channel.Qos(1, 0, false)
	if err != nil {
		return err
	}

	// set up queue
	queue, err := channel.QueueDeclare(
		"",    // name, let server pick
		false, // durable
		true,  // autoDelete
		true,  // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return err
	}
	c.replyTo = c.ResponseRoutingKey + "." + queue.Name
	err = channel.QueueBind(
		queue.Name,
		c.replyTo, // routing key
		c.Exchange,
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return err
	}

	// spawn queue consumer
	go c.consumer(queue.Name)

	return nil
}

func (c *Client) consumer(queueName string) {
	logger := c.Logger.With().Str("module", "rpc-consumer").Logger()

	// create delivery channel
	deliveries, err := c.channel.Consume(
		queueName,
		"",
		true,  // autoAck
		true,  // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		logger.Err(err).Msg("Unable to create delivery channel.")
	}

	// process deliveries
	for delivery := range deliveries {
		// send response out to caller
		if callback, ok := c.callers[delivery.CorrelationId]; ok {
			callback <- delivery.Body
		} else {
			logger.Error().Str("correlation_id", delivery.CorrelationId).Msg("Received response with no caller?")
		}
	}
}

// Call makes a RPC call, and initializes the client on first use.
func (c *Client) Call(callName string, arguments []byte, timeout time.Duration) ([]byte, error) {
	// init client
	if err := c.setup(); err != nil {
		return nil, err
	}

	// get next ID in sequence for our CorrelationID
	correlationID := strconv.FormatUint(atomic.AddUint64(&c.seq, 1), 10)

	// get logger
	logger := c.Logger.With().
		Str("module", "rpc-call").
		Str("call_name", callName).
		Str("correlation_id", correlationID).Logger()
	logger.Debug().Msg("Making RPC call.")

	// create correlation channel
	callback := make(chan []byte)
	c.callers[correlationID] = callback
	defer delete(c.callers, correlationID) // prevent a memory leak

	// send our request
	err := c.channel.Publish(
		c.Exchange,
		c.RequestRoutingKey+"."+callName,
		true,
		false,
		amqp.Publishing{
			CorrelationId: correlationID,
			Body:          arguments,
		},
	)
	if err != nil {
		return nil, err
	}

	// return callback
	select {
	case result := <-callback:
		return result, nil
	case <-time.After(timeout):
		return nil, errors.New("request timed out")
	}
}

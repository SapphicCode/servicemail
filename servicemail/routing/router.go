package routing

import (
	"github.com/Pandentia/servicemail/rpc"

	"github.com/Pandentia/servicemail/servicemail"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
)

// Router represents a Mail router.
type Router struct {
	MQURI  string // The AMQP message queue URL to dial.
	Logger zerolog.Logger

	conn    *amqp.Connection
	channel *amqp.Channel
	rpc     *rpc.Client
}

// New initializes the Router struct. It should only be called once.
func (r *Router) New() error {
	// create the connection
	conn, err := amqp.Dial(r.MQURI)
	if err != nil {
		// _ = conn.Close()
		return err
	}
	r.conn = conn
	r.Logger.Debug().Msg("Connection established")

	// create the channel
	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return err
	}
	r.channel = channel
	r.Logger.Debug().Msg("Channel established")

	// set prefetching
	err = r.channel.Qos(1, 0, false)
	if err != nil {
		return err
	}
	r.Logger.Debug().Msg("Prefetching set")

	// register the exchange
	err = channel.ExchangeDeclare(servicemail.Exchange, "topic", true, false, false, false, nil)
	if err != nil {
		_ = conn.Close()
		return err
	}
	r.Logger.Debug().Msg("Exchange registered")

	// register the routing queue
	queue, err := channel.QueueDeclare(servicemail.RoutingQueue, true, false, false, false, nil)
	if err != nil {
		_ = conn.Close()
		return err
	}
	r.Logger.Debug().Msg("Routing queue registered")
	// bind the routing queue to the exchange
	err = channel.QueueBind(queue.Name, servicemail.IngressRoutingKey+".*", servicemail.Exchange, false, nil)
	if err != nil {
		_ = conn.Close()
		return err
	}
	r.Logger.Debug().Msg("Routing queue bound to exchange")

	// create the RPC client
	rpc := &rpc.Client{
		Logger:             r.Logger,
		Connection:         conn,
		Exchange:           servicemail.Exchange,
		RequestRoutingKey:  servicemail.RPCRoutingKey,
		ResponseRoutingKey: servicemail.RPCResponseRoutingKey,
	}
	r.rpc = rpc
	r.Logger.Debug().Msg("RPC client created")

	return nil
}

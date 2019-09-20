package routing

import (
	"github.com/Pandentia/servicemail/servicemail"
	"github.com/streadway/amqp"
	"github.com/vmihailenco/msgpack/v4"
)

// Run starts the router. It will (ideally) block indefinitely.
func (r *Router) Run() error {
	logger := r.Logger.With().Str("module", "consumer").Logger()

	// enable prefetching
	err := r.channel.Qos(1, 0, false)
	if err != nil {
		return err
	}

	// begin consuming
	deliveries, err := r.channel.Consume(servicemail.RoutingQueue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	// begin processing
	for delivery := range deliveries {
		var mail servicemail.Mail

		// receive mail from ingress
		if err := msgpack.Unmarshal(delivery.Body, &mail); err != nil {
			logger.Err(err).Bytes("data", delivery.Body).Msg("Error deserializing. Rejecting and continuing.")
			delivery.Reject(false) // do *not* requeue, otherwise we'll just be stuck processing garbage
			continue
		}

		// management routing request
		resp, err := r.rpc.Call(servicemail.RoutingCall, mail, servicemail.DefaultRPCTimeout)
		if err != nil {
			logger.Err(err).Msg("Error sending RPC request. Requeuing delivery.")
			delivery.Reject(true)
			continue
		}
		envelopes := resp.([]servicemail.Envelope)

		// send envelopes to platform-specific delivery nodes
		// beyond this point we can't requeue, so we error gracefully
		delivery.Ack(false) // acknowledge delivery, as we are past the point of no return
		for _, envelope := range envelopes {
			// encode envelope
			data, err := msgpack.Marshal(envelope)
			if err != nil {
				logger.Err(err).Msg("Error serializing envelope.")
				continue
			}

			// publish envelope to delivery queues
			err = r.channel.Publish(
				servicemail.Exchange,
				servicemail.DeliveryRoutingKey+"."+envelope.Platform,
				true,  // mandatory
				false, // immediate
				amqp.Publishing{
					Body: data,
				},
			)
			if err != nil {
				logger.Err(err).Str("platform", envelope.Platform).Msg("Error publishing envelope.")
			}
		}
	}

	return nil
}

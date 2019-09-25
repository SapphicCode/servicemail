package routing

import (
	"github.com/Pandentia/servicemail/servicemail"
	"github.com/streadway/amqp"
)

func mustSerialize(v interface{}) []byte {
	if data, err := servicemail.Marshal(v); err != nil {
		panic(err)
	} else {
		return data
	}
}

// Run starts the router. It will (ideally) block indefinitely.
func (r *Router) Run() error {
	logger := r.Logger.With().Str("module", "consumer").Logger()
	logger.Info().Msg("Router started")

	// begin consuming
	deliveries, err := r.channel.Consume(servicemail.RoutingQueue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	// begin processing
	for delivery := range deliveries {
		var mail servicemail.Mail

		logger.Debug().Msg("Delivery received from ingress")

		// receive mail from ingress
		if err := servicemail.Unmarshal(delivery.Body, &mail); err != nil {
			logger.Err(err).Bytes("data", delivery.Body).Msg("Error deserializing. Rejecting and continuing.")
			delivery.Reject(false) // do *not* requeue, otherwise we'll just be stuck processing garbage
			continue
		}
		logger.Debug().Interface("mail", mail).Msg("Message successfully deserialized.")

		// I'm very aware we're deserializing and immediately re-serializing here. This is mostly for flexibility.
		// A possible example for this use-case: Periodically fetching a banned domains list and rejecting mail here.

		// management routing request
		serializedEnvelopes, err := r.rpc.Call(
			servicemail.RoutingCall, mustSerialize(mail), servicemail.DefaultRPCTimeout,
		)
		if err != nil {
			logger.Err(err).Msg("Error sending RPC request. Requeuing delivery.")
			delivery.Reject(true)
			continue
		}
		envelopes := make([]servicemail.Envelope, 1) // the vast majority of the time it's going to be one or zero
		if err := servicemail.Unmarshal(serializedEnvelopes, &envelopes); err != nil {
			logger.Err(err).Msg("Management returned garbage. Requeing delivery.")
			delivery.Reject(true)
			continue
		}
		logger.Debug().Interface("envelopes", envelopes).Msg("Envelopes successfully deserialized from management.")

		// send envelopes to platform-specific delivery nodes
		// beyond this point we can't requeue, so we error gracefully
		delivery.Ack(false) // acknowledge delivery, as we are past the point of no return
		for _, envelope := range envelopes {
			// encode envelope
			data, err := servicemail.Marshal(envelope)
			if err != nil {
				logger.Err(err).Msg("Error serializing envelope.")
				continue
			}

			// publish envelope to delivery queues
			err = r.channel.Publish(
				servicemail.Exchange,
				servicemail.DeliveryRoutingKey+"."+envelope.Platform,
				false, // mandatory
				false, // immediate
				amqp.Publishing{
					Body: data,
				},
			)
			if err != nil {
				logger.Err(err).Interface("envelope", envelope).Msg("Error publishing envelope.")
			}
		}
	}

	return nil
}

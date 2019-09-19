package routing

import (
	"github.com/Pandentia/servicemail/servicemail"
	"github.com/vmihailenco/msgpack/v4"
)

// Run starts the router. It will (ideally) block indefinitely.
func (r *Router) Run() error {
	logger := r.Logger.With().Str("module", "consumer").Logger()

	deliveries, err := r.channel.Consume(servicemail.RoutingQueue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for delivery := range deliveries {
		var mail servicemail.Mail
		if err := msgpack.Unmarshal(delivery.Body, &mail); err != nil {
			logger.Err(err).Bytes("data", delivery.Body).Msg("Error deserializing. Rejecting and continuing.")
			delivery.Reject(false) // do *not* requeue, otherwise we'll just be stuck processing garbage
			continue
		}

		// TODO: management request (REST?), where should this mail go (if anywhere)

		// TODO: package into Envelope, dispatch to relevant routing key (delivery.<platform>)
	}

	return nil
}

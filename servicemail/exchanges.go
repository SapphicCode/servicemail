package servicemail

// Exchange describes the RabbitMQ exchange name.
const Exchange = "servicemail"

// Message broker queue names.
const (
	RoutingQueue  = Exchange + ".routing"
	DeliveryQueue = Exchange + ".deliveries"
)

// Routing key prefixes.
const (
	IngressRoutingKey  = "ingress"
	DeliveryRoutingKey = "delivery"
)

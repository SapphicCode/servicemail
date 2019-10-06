package management

import (
	"github.com/Pandentia/servicemail/rpc"
	"github.com/Pandentia/servicemail/servicemail"

	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
)

// Server represents the management server.
type Server struct {
	Logger     zerolog.Logger
	Connection *amqp.Connection

	rpc *rpc.Server
}

// Run runs the Server.
func (server *Server) Run() error {
	logger := server.Logger.With().Str("module", "runner").Logger()

	handlers := make(map[string]func([]byte) []byte)
	handlers[servicemail.RoutingCall] = server.routingHandler
	logger.Debug().Msgf("%d handlers registered", len(handlers))

	rpcServer := &rpc.Server{
		Logger:   server.Logger,
		Handlers: handlers,

		Connection: server.Connection,
		Exchange:   servicemail.Exchange,
		RoutingKey: servicemail.RPCRoutingKey,
	}

	rpcServer.Run()
	select {} // sleep forever, the goroutines for the queues are running in the background
	// TODO: refactor to a WaitGroup?
}

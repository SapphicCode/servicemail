package management

import (
	"os"
	"os/signal"
	"sync"

	"github.com/rs/zerolog"
	"github.com/streadway/amqp"

	"github.com/Pandentia/servicemail/rpc"
	"github.com/Pandentia/servicemail/servicemail"
)

// Server represents the management server.
type Server struct {
	Logger     zerolog.Logger
	Connection *amqp.Connection

	rpc       *rpc.Server
	waitGroup *sync.WaitGroup
}

// Run runs the Server.
func (server *Server) Run() (err error) {
	logger := server.Logger.With().Str("module", "runner").Logger()

	handlers := make(map[string]func([]byte) []byte)
	handlers[servicemail.RoutingCall] = server.routingHandler
	logger.Debug().Msgf("%d handlers registered", len(handlers))

	rpcServer := &rpc.Server{
		Logger:    server.Logger,
		WaitGroup: &sync.WaitGroup{},
		Handlers:  handlers,

		Connection: server.Connection,
		Exchange:   servicemail.Exchange,
		RoutingKey: servicemail.RPCRoutingKey,
	}

	// set up signal handlers
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, os.Interrupt)
	go stopSignalHandler(rpcServer, signalChannel)

	// run server
	err = rpcServer.Run()
	if err != nil {
		rpcServer.WaitGroup.Wait()
	}
	return
}

func stopSignalHandler(server *rpc.Server, signalChannel <-chan os.Signal) {
	sig := <-signalChannel
	server.Logger.Info().Str("signal", sig.String()).Msg(
		"Stop signal received. Exiting gracefully. Send another signal to stop immediately.",
	)
	signal.Reset()

	_ = server.Connection.Close()
}

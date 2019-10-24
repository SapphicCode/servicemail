package management

import (
	"github.com/rs/zerolog"

	"github.com/Pandentia/servicemail/servicemail"
)

func (server *Server) routingHandler(data []byte) []byte {
	logger := server.getLogger("handler.routing")
	logger.Debug().Bytes("data", data).Msg("Received mail routing request.")

	var mail servicemail.Mail
	err := servicemail.Unmarshal(data, &mail)
	if err != nil {
		logger.Err(err).Msg("Error deserializing message.")
		return []byte("[]")
	}

	envelopes := server.processRoutingRequest(logger, mail)
	data, err = servicemail.Marshal(envelopes)
	if err != nil {
		logger.Err(err).Msg("Error serializing envelopes.")
	}
	logger.Debug().Bytes("data", data).Msg("Answered mail routing request.")
	return data
}

func (server *Server) processRoutingRequest(logger zerolog.Logger, mail servicemail.Mail) []servicemail.Envelope {
	logger.Debug().Interface("mail", mail).Msg("Processing mail routing request.")

	row, err := server.DB.Query("SELECT account_id FROM servicemail.addresses WHERE address = ?")
	if err != nil {
		logger.Err(err).Msg("Error making database query?")
	}
	
}

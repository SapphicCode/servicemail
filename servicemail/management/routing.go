package management

import "github.com/Pandentia/servicemail/servicemail"

func (server *Server) routingHandler(data []byte) []byte {
	var mail servicemail.Mail
	servicemail.Unmarshal(data, &mail)
}

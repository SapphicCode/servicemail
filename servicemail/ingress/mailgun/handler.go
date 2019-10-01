package mailgun

import (
	"regexp"
	"strings"

	"github.com/Pandentia/servicemail/servicemail"
	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

const routingKey = servicemail.IngressRoutingKey + ".mailgun"

var senderRegex = regexp.MustCompile(`(?P<name>.+) <(?P<email>\S+@\S+)>`)

func (api *API) mailgunHandler(c *gin.Context) {
	logger := api.Logger.With().Str("module", "handler").Logger()
	m := servicemail.Mail{}

	logger.Debug().Msg("Request received")

	// first thing's first, extracting sender info
	match := senderRegex.FindStringSubmatch(c.PostForm("from"))
	if match == nil {
		c.AbortWithStatus(400)
		logger.Error().Msg("Error parsing email sender")
		return
	}
	for i, name := range senderRegex.SubexpNames() {
		if i == 0 {
			continue
		}
		if name == "name" {
			m.SenderName = strings.TrimSpace(match[i])
		}
		if name == "email" {
			m.From = match[i]
		}
	}

	// extract recipient
	m.To = c.PostForm("recipient")

	// extract subject
	m.Subject = c.PostForm("subject")

	// extract body (but don't parse)
	m.Body = c.PostForm("body-html")

	logger.Debug().Interface("mail", m).Msg("Finished parsing mail")

	// encode Mail
	mailData, err := servicemail.Marshal(m)
	if err != nil {
		c.AbortWithStatus(500)
		logger.Err(err).Interface("mail", mailData).Msg("Error encoding Mail object")
	}

	// publish to router
	err = api.channel.Publish(
		servicemail.Exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			Body: mailData,
		},
	)
	if err != nil {
		c.AbortWithStatus(500)
		logger.Err(err).Msg("Error publishing to AMQP")
	}
	c.Status(204)
	logger.Debug().Msg("Request successfully parsed and published")
}

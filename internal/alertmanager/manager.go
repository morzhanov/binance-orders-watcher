package alertmanager

import (
	"github.com/mailjet/mailjet-apiv3-go"
)

type Manager interface {
	SendAlert(from, fromName, to, toName, text string) error
}

type manager struct {
	mailjetClient mailjet.Client
}

func New(mailjetApiKey, mailjetApiSecret string) Manager {
	mailjet.NewMailjetClient(mailjetApiKey, mailjetApiSecret)
	return &manager{}
}

func (m *manager) SendAlert(from, fromName, to, toName, text string) error {
	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: from,
				Name:  fromName,
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: to,
					Name:  toName,
				},
			},
			Subject:  "Binance Alert",
			TextPart: text,
		},
	}
	messages := mailjet.MessagesV31{Info: messagesInfo}
	_, err := m.mailjetClient.SendMailV31(&messages)
	return err
}

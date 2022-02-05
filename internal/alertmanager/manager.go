package alertmanager

import (
	"github.com/mailjet/mailjet-apiv3-go"
)

type Manager interface {
	SendAlert(toEmail, toName, text string) error
}

type manager struct {
	mailjetClient *mailjet.Client
	senderEmail   string
	senderName    string
}

func New(mailjetApiKey, mailjetApiSecret, senderName, senderEmail string) Manager {
	client := mailjet.NewMailjetClient(mailjetApiKey, mailjetApiSecret)
	return &manager{mailjetClient: client, senderName: senderName, senderEmail: senderEmail}
}

func (m *manager) SendAlert(toEmail, toName, text string) error {
	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: m.senderEmail,
				Name:  m.senderName,
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: toEmail,
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

package main

import (
	"fmt"
	"net/http"

	"gitlab.com/asmaloney/magsli/mailgun"
	"gitlab.com/asmaloney/magsli/slack"
)

func handler(w http.ResponseWriter, r *http.Request) {
	valid, _ := mailgun.VerifyMessage(mailGunAPIKey, r)

	if !valid {
		// don't log this as it's not useful and can be abused
		return
	}

	data, err := mailgun.NewMailGunData(r)

	if (data == mailgun.Data{}) {
		fmt.Printf("Could not decode message from MailGun: %v\n", err)
		return
	}

	msg := newSlackMessageFromMailGunData(data)

	err = msg.Send(slackWebhookURL)

	if err != nil {
		fmt.Printf("Could not send message to Slack: %v\n", err)
	}
}

func newSlackMessageFromMailGunData(d mailgun.Data) (msg slack.Message) {
	text := fmt.Sprintf("MailGun message for domain: %v", d.Domain)

	msg = slack.NewMessage(text)

	if d.IsErrorEvent() {
		msg.AddError("Event", d.EventType, true)
	} else {
		msg.AddData("Event", d.EventType, true)

	}
	msg.AddData("Message ID", d.MessageID, false)
	msg.AddData("Recipient", d.Recipient, true)
	msg.AddData("Subject", d.Subject, true)

	switch d.EventType {
	case "bounced":
		msg.AddError("SMTP Code", d.SMTPCode, true)
		msg.AddError("SMTP Error", d.SMTPError, false)
	case "dropped":
		msg.AddError("Reason", d.Reason, true)
		msg.AddError("ESP Code", d.ESPCode, true)
		msg.AddError("Description", d.Description, false)
	}

	return msg
}

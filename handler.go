package main

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"gitlab.com/asmaloney/magsli/mailgundata"
	"gitlab.com/asmaloney/magsli/slack"
)

func handler(w http.ResponseWriter, r *http.Request) {
	var payload mailgundata.Payload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.Error("could not decode message from MailGun", "error", err)
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	verified, err := mailgundata.VerifyWebhookSignature(mailGunAPIKey, payload.Signature)
	if err != nil {
		slog.Error("could not verify message from MailGun", "error", err)
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	if !verified {
		// Commented out to avoid spamming the logs
		//slog.Error("MailGun message failed verification", "request", *r)
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	data, err := mailgundata.NewMailGunData(payload.EventData)
	if (data == mailgundata.Data{}) {
		slog.Error("could not decode message from MailGun", "error", err)
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	msg := newSlackMessageFromMailGunData(data)

	err = msg.Send(slackWebhookURL)

	if err != nil {
		slog.Error("could not send message to Slack", "error", err)
	}
}

func newSlackMessageFromMailGunData(d mailgundata.Data) (msg slack.Message) {
	msg = slack.NewMessage("MailGun Error")

	msg.AddError("Event", d.EventType, true)

	switch d.EventType {
	case "rejected":
		msg.AddData("Message ID", d.RejectedEvent.ID, true)
		msg.AddData("Subject", d.RejectedEvent.Message.Headers.Subject, true)
		msg.AddData("To", d.RejectedEvent.Message.Headers.To, true)
		msg.AddData("Reason", d.RejectedEvent.Reject.Reason, false)
		msg.AddData("Description", d.RejectedEvent.Reject.Description, false)

	case "failed":
		msg.AddData("Message ID", d.FailedEvent.ID, true)
		msg.AddData("Recipient", d.FailedEvent.Recipient, true)
		msg.AddData("Subject", d.FailedEvent.Message.Headers.Subject, true)
		msg.AddData("Severity", d.FailedEvent.Severity, true)
		msg.AddData("DeliveryStatus", d.FailedEvent.DeliveryStatus.Message, false)
		msg.AddData("Reason", d.FailedEvent.Reason, false)
	}

	return msg
}

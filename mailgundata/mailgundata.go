package mailgundata

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	"github.com/mailgun/mailgun-go/v4/events"
)

var (
	ErrUnrecognizedEvent = fmt.Errorf("unrecognized MailGun event")
)

type ErrParseEvent struct {
	Name string
	Err  error
}

func (err ErrParseEvent) Error() string {
	return fmt.Sprintf("failed to parse MailGun event %q: %v", err.Name, err.Err)
}

// Signature is the signature data from the webhook POST
type Signature struct {
	TimeStamp string `json:"timestamp"`
	Token     string `json:"token"`
	Signature string `json:"signature"`
}

// Payload is the body of the webhook POST
type Payload struct {
	Signature Signature      `json:"signature"`
	EventData events.RawJSON `json:"event-data"`
}

// Data defines the fields from MailGun that we are going to use for our Slack message.
type Data struct {
	EventType string

	RejectedEvent *events.Rejected
	FailedEvent   *events.Failed
}

// VerifyWebhookSignature verifies the signature coming from our MailGun API.
// See https://github.com/mailgun/mailgun-go/
func VerifyWebhookSignature(apiKey string, sig Signature) (verified bool, err error) {
	h := hmac.New(sha256.New, []byte(apiKey))

	_, err = io.WriteString(h, sig.TimeStamp)
	if err != nil {
		return
	}
	_, err = io.WriteString(h, sig.Token)
	if err != nil {
		return
	}

	calculatedSignature := h.Sum(nil)
	signature, err := hex.DecodeString(sig.Signature)
	if err != nil {
		return false, err
	}

	if len(calculatedSignature) != len(signature) {
		return false, nil
	}

	return subtle.ConstantTimeCompare(signature, calculatedSignature) == 1, nil
}

// NewMailGunData converts raw bytes into a Data struct.
func NewMailGunData(raw []byte) (Data, error) {
	var eventName events.EventName
	if err := json.Unmarshal(raw, &eventName); err != nil {
		return Data{}, ErrUnrecognizedEvent
	}

	data := Data{EventType: eventName.Name}

	switch eventName.Name {
	case "rejected":
		var event events.Rejected
		if err := json.Unmarshal(raw, &event); err != nil {
			return Data{}, ErrParseEvent{
				Name: eventName.GetName(),
				Err:  err,
			}
		}

		data.RejectedEvent = &event

	case "failed":
		var event events.Failed
		if err := json.Unmarshal(raw, &event); err != nil {
			return Data{}, ErrParseEvent{
				Name: eventName.GetName(),
				Err:  err,
			}
		}

		data.FailedEvent = &event

	default:
	}

	return data, nil
}

package slack

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

const (
	dataAttachmentName  = "Data"
	errorAttachmentName = "Errors"

	notFound = -1
)

// Message is a message to be posted to the Slack webhook.
type Message struct {
	Name        string       `json:"username"`
	Message     string       `json:"text"`
	Emoji       string       `json:"icon_emoji"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Attachment is an attachment to a Message.
// This implements a small subset of the fields.
// https://api.slack.com/docs/attachments
type Attachment struct {
	Fallback string  `json:"fallback"`
	Color    string  `json:"color"` // 'good', 'warning', 'danger' or a HEX value
	Fields   []Field `json:"fields,omitempty"`
}

// Field is a field in an Attachment.
// https://api.slack.com/docs/attachments
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"` // is it short enough to display side-by-side?
}

// NewMessage returns a new Slack message with default values.
func NewMessage(text string) Message {
	return Message{
		Name:    "magsli",
		Message: text,
		Emoji:   ":moyai:",
	}
}

// AddData adds a new field to the "Data" Attachment.
func (m *Message) AddData(title string, value string, isShort bool) {
	_, _ = m.addDataToAttachment(dataAttachmentName, title, value, isShort)
}

// AddError adds a new field to the "Errors" Attachment and changes the message emoji.
func (m *Message) AddError(title string, value string, isShort bool) {
	a, err := m.addDataToAttachment(errorAttachmentName, title, value, isShort)

	if err != nil {
		return
	}

	a.Color = "danger"

	m.Emoji = ":rotating_light:"
}

func (m *Message) addDataToAttachment(attachmentName string, title string, value string, isShort bool) (*Attachment, error) {
	if value == "" {
		return nil, errors.New("no value in data")
	}

	a := m.findOrCreateAttachment(attachmentName)

	f := Field{title, value, isShort}
	a.Fields = append(a.Fields, f)

	return a, nil
}

func (m *Message) findAttachment(name string) int {
	for index, a := range m.Attachments {
		if a.Fallback == name {
			return index
		}
	}

	return notFound
}

func (m *Message) newAttachment() *Attachment {
	a := Attachment{}

	m.Attachments = append(m.Attachments, a)

	return &m.Attachments[len(m.Attachments)-1]
}

func (m *Message) findOrCreateAttachment(name string) *Attachment {
	i := m.findAttachment(name)

	if i == notFound {
		a := m.newAttachment()
		a.Fallback = name

		return a
	}

	return &m.Attachments[i]
}

// Send POSTs a message to the Slack webhook
func (m *Message) Send(webhookURL string) error {
	b, err := json.Marshal(m)

	if err != nil {
		return err
	}

	client := http.Client{Timeout: time.Duration(1) * time.Second}

	url := webhookURL

	resp, err := client.Post(url, "application/json", strings.NewReader(string(b)))

	if err != nil {
		return err
	}

	return resp.Body.Close()
}

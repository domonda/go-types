package email

import (
	"bytes"
	"fmt"

	"github.com/inbucket/html2text"
	"github.com/teamwork/tnef"

	"github.com/domonda/go-errs"
	"github.com/domonda/go-types/nullable"
)

// ParseTNEFMessageBytes parses a TNEF (Transport Neutral Encapsulation
// Format) email message from the passed bytes.
// If the plain text body is not longer than one byte but there is an
// HTML body, then the Body is set to the plain text converted from the HTML.
func ParseTNEFMessageBytes(messageBytes []byte) (msg *Message, err error) {
	defer errs.WrapWithFuncParams(&err, messageBytes)

	if len(messageBytes) < 4 {
		return nil, errs.New("message too short")
	}

	t, err := tnef.Decode(messageBytes)
	if err != nil {
		return nil, err
	}

	msg = &Message{
		Body:        string(t.Body),
		BodyHTML:    nullable.TrimmedStringFrom(string(t.BodyHTML)),
		ExtraHeader: make(Header),
	}
	if len(t.Body) <= 1 && len(t.BodyHTML) > len(t.Body) {
		text, err := html2text.FromReader(bytes.NewReader(t.BodyHTML))
		if err != nil {
			return nil, err
		}
		msg.Body = text
	}
	for i, attachment := range t.Attachments {
		msg.Attachments = append(msg.Attachments, &Attachment{
			PartID:   fmt.Sprintf("TNEF%d", i),
			Inline:   false, // ??
			Filename: attachment.Title,
			Content:  attachment.Data,
		})
	}

	return msg, nil
}

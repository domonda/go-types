package email

import (
	"bytes"
	"fmt"
	"io"

	"github.com/jhillyerd/enmime/v2"

	"github.com/domonda/go-errs"
	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/strutil"
)

func ParseMIMEMessage(reader io.Reader) (msg *Message, err error) {
	defer errs.WrapWithFuncParams(&err, reader)

	envelope, err := enmime.ReadEnvelope(reader)
	if err != nil {
		return nil, err
	}

	msg = &Message{
		// From:        Address(envelope.GetHeader("From")),
		// ReplyTo:     NullableAddress(strutil.TrimSpace(envelope.GetHeader("Reply-To"))),
		MessageID:   nullable.TrimmedStringFrom(envelope.GetHeader("Message-Id")),
		InReplyTo:   nullable.TrimmedStringFrom(envelope.GetHeader("In-Reply-To")),
		References:  nullable.TrimmedStringFrom(envelope.GetHeader("References")),
		Subject:     strutil.TrimSpace(envelope.GetHeader("Subject")),
		Body:        envelope.Text,
		BodyHTML:    nullable.TrimmedStringFrom(envelope.HTML),
		ExtraHeader: make(Header),
	}
	if date := envelope.GetHeader("Date"); date != "" {
		msg.Date, err = parseDate(date)
		if err != nil {
			return nil, err
		}
	}
	msg.From, err = NormalizedAddress(envelope.GetHeader("From"))
	if err != nil {
		return nil, fmt.Errorf("can't parse email header 'From': %w", err)
	}
	msg.ReplyTo, err = NormalizedNullableAddress(envelope.GetHeader("Reply-To"))
	if err != nil {
		return nil, fmt.Errorf("can't parse email header 'Reply-To': %w", err)
	}
	for _, to := range envelope.GetHeaderValues("To") {
		addrs, err := AddressList(to).Split()
		if err != nil {
			return nil, fmt.Errorf("can't parse email header 'To': %w", err)
		}
		msg.To = msg.To.Append(addrs...)
	}
	for _, deliveredTo := range envelope.GetHeaderValues("Delivered-To") {
		addrs, err := AddressList(deliveredTo).Split()
		if err != nil {
			return nil, fmt.Errorf("can't parse email header 'Delivered-To': %w", err)
		}
		msg.DeliveredTo = msg.DeliveredTo.Append(addrs...)
	}
	for _, cc := range envelope.GetHeaderValues("Cc") {
		addrs, err := AddressList(cc).Split()
		if err != nil {
			return nil, fmt.Errorf("can't parse email header 'Cc': %w", err)
		}
		msg.Cc = msg.Cc.Append(addrs...)
	}
	for _, bcc := range envelope.GetHeaderValues("Bcc") {
		addrs, err := AddressList(bcc).Split()
		if err != nil {
			return nil, fmt.Errorf("can't parse email header 'Bcc': %w", err)
		}
		msg.Bcc = msg.Bcc.Append(addrs...)
	}

	for key, values := range envelope.Root.Header {
		if IsExtraHeader(key) {
			for _, value := range values {
				msg.ExtraHeader.Add(key, value)
			}
		}
	}

	for _, part := range envelope.Attachments {
		msg.Attachments = append(msg.Attachments, &Attachment{
			PartID:      part.PartID,
			ContentID:   part.ContentID,
			ContentType: part.ContentType,
			Inline:      false,
			Filename:    part.FileName,
			Content:     part.Content,
		})
	}
	for _, part := range envelope.Inlines {
		msg.Attachments = append(msg.Attachments, &Attachment{
			PartID:      part.PartID,
			ContentID:   part.ContentID,
			ContentType: part.ContentType,
			Inline:      true,
			Filename:    part.FileName,
			Content:     part.Content,
		})
	}

	return msg, nil
}

func ParseMIMEMessageBytes(msgBytes []byte) (msg *Message, err error) {
	defer errs.WrapWithFuncParams(&err, msgBytes)

	return ParseMIMEMessage(bytes.NewReader(msgBytes))
}

package email

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/jhillyerd/enmime"
	"github.com/ungerik/go-fs"

	"github.com/domonda/go-errs"
	"github.com/domonda/go-types/nullable"
)

func ParseMIMEMessage(reader io.Reader) (msg *Message, err error) {
	defer errs.WrapWithFuncParams(&err, reader)

	envelope, err := enmime.ReadEnvelope(reader)
	if err != nil {
		return nil, err
	}

	msg = &Message{
		MessageID:   nullable.TrimmedStringFrom(envelope.GetHeader("Message-Id")),
		InReplyTo:   nullable.TrimmedStringFrom(envelope.GetHeader("In-Reply-To")),
		References:  nullable.TrimmedStringFrom(envelope.GetHeader("References")),
		Subject:     strings.TrimSpace(envelope.GetHeader("Subject")),
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

	for _, attachment := range envelope.Attachments {
		msg.Attachments = append(msg.Attachments, &Attachment{
			PartID:      attachment.PartID,
			ContentID:   attachment.ContentID,
			ContentType: attachment.ContentType,
			File: fs.MemFile{
				FileName: attachment.FileName,
				FileData: attachment.Content,
			},
		})
	}

	return msg, nil
}

func ParseMIMEMessageBytes(msgBytes []byte) (msg *Message, err error) {
	defer errs.WrapWithFuncParams(&err, msgBytes)

	return ParseMIMEMessage(bytes.NewReader(msgBytes))
}

func ParseMIMEMessageFile(file fs.FileReader) (msg *Message, err error) {
	defer errs.WrapWithFuncParams(&err, file)

	reader, err := file.OpenReader()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return ParseMIMEMessage(reader)
}

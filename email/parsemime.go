package email

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime"

	"github.com/jhillyerd/enmime/v2"
	"golang.org/x/net/html/charset"

	"github.com/domonda/go-errs"
	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/strutil"
)

// mimeHeaderDecoder decodes RFC 2047 encoded-words in email headers.
// The CharsetReader extends Go's stdlib mime support (us-ascii, utf-8,
// iso-8859-1) to the full IANA charset set (iso-8859-2…16, windows-125x,
// koi8, big5, shift_jis, …) so that headers from senders using other
// charsets are decoded instead of failing the whole message.
var mimeHeaderDecoder = &mime.WordDecoder{CharsetReader: charset.NewReaderLabel}

// ParseMIMEMessage parses a MIME email message read from the passed reader.
//
// As long as the underlying envelope can be read, a non-nil message with
// all usable data is returned even if individual headers (Date, From, To)
// can't be parsed. Such parsing errors are collected and returned joined
// via errors.Join alongside the message, so callers can use the partial
// result while still being able to inspect what went wrong.
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
	// parseErrs collects non-fatal header parsing errors so that a message
	// with usable data is still returned together with the joined errors.
	var parseErrs []error

	if date := envelope.GetHeader("Date"); date != "" {
		msg.Date, err = parseDate(date)
		if err != nil {
			parseErrs = append(parseErrs, fmt.Errorf("can't parse email header 'Date': %w", err))
			msg.Date = nil
		}
	}
	msg.From, err = NormalizedAddress(envelope.GetHeader("From"))
	if err != nil {
		parseErrs = append(parseErrs, fmt.Errorf("can't parse email header 'From': %w", err))
		msg.From = ""
	}
	msg.ReplyTo, err = NormalizedNullableAddress(envelope.GetHeader("Reply-To"))
	if err != nil {
		// intentionally ignoring parsing issues with Reply-To, unset the value.
		// we've seen weird and unparsable values
		msg.ReplyTo = NullableAddress("")
	}
	for _, to := range envelope.GetHeaderValues("To") {
		// Split returns all parsable addresses even when others fail,
		// so keep the usable ones and collect the error.
		addrs, err := AddressList(to).Split()
		if err != nil {
			parseErrs = append(parseErrs, fmt.Errorf("can't parse email header 'To': %w", err))
		}
		msg.To = msg.To.Append(addrs...)
	}
	for _, deliveredTo := range envelope.GetHeaderValues("Delivered-To") {
		// intentionally ignoring parsing issues with nullable lists
		// (we've seen weird and unparsable values) but keep the
		// addresses that could be parsed
		addrs, _ := AddressList(deliveredTo).Split()
		msg.DeliveredTo = msg.DeliveredTo.Append(addrs...)
	}
	for _, cc := range envelope.GetHeaderValues("Cc") {
		// intentionally ignoring parsing issues with nullable lists
		// (we've seen weird and unparsable values) but keep the
		// addresses that could be parsed
		addrs, _ := AddressList(cc).Split()
		msg.Cc = msg.Cc.Append(addrs...)
	}
	for _, bcc := range envelope.GetHeaderValues("Bcc") {
		// intentionally ignoring parsing issues with nullable lists
		// (we've seen weird and unparsable values) but keep the
		// addresses that could be parsed
		addrs, _ := AddressList(bcc).Split()
		msg.Bcc = msg.Bcc.Append(addrs...)
	}

	for key, values := range envelope.Root.Header {
		if IsExtraHeader(key) {
			for _, value := range values {
				decodedvalue, err := mimeHeaderDecoder.DecodeHeader(value)
				if err != nil {
					// ignore decoding issues, just use value
					msg.ExtraHeader.Add(key, value)
				} else {
					msg.ExtraHeader.Add(key, decodedvalue)
				}
			}
		}
	}

	for _, part := range envelope.Attachments {
		msg.Attachments = append(msg.Attachments, &Attachment{
			PartID:      part.PartID,
			ContentID:   part.ContentID,
			ContentType: part.ContentType,
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
	for _, part := range envelope.OtherParts {
		msg.Attachments = append(msg.Attachments, &Attachment{
			PartID:      part.PartID,
			ContentID:   part.ContentID,
			ContentType: part.ContentType,
			OtherPart:   true,
			Filename:    part.FileName,
			Content:     part.Content,
		})
	}

	return msg, errors.Join(parseErrs...)
}

// ParseMIMEMessageBytes parses a MIME email message from the passed bytes.
func ParseMIMEMessageBytes(msgBytes []byte) (msg *Message, err error) {
	defer errs.WrapWithFuncParams(&err, msgBytes)

	return ParseMIMEMessage(bytes.NewReader(msgBytes))
}

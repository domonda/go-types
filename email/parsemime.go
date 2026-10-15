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
	typescharset "github.com/domonda/go-types/charset"
	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/strutil"
)

// mimeHeaderDecoder decodes RFC 2047 encoded-words in email headers.
// The CharsetReader extends Go's stdlib mime support (us-ascii, utf-8,
// iso-8859-1) to the full IANA charset set (iso-8859-2…16, windows-125x,
// koi8, big5, shift_jis, …) so that headers from senders using other
// charsets are decoded instead of failing the whole message.
var mimeHeaderDecoder = &mime.WordDecoder{CharsetReader: headerCharsetReader}

// headerCharsetReader returns a reader decoding input from the charset
// label with the WHATWG labels of charset.NewReaderLabel, or else with the
// IANA labels of the charset package of this module, like "cp850".
func headerCharsetReader(label string, input io.Reader) (io.Reader, error) {
	reader, err := charset.NewReaderLabel(label, input)
	if err == nil {
		return reader, nil
	}
	enc, encErr := typescharset.GetEncoding(label)
	if encErr != nil {
		return nil, err
	}
	encoded, err := io.ReadAll(input)
	if err != nil {
		return nil, err
	}
	decoded, err := enc.Decode(encoded)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(decoded), nil
}

// addressHeader returns the raw value of an address header for parsing,
// so that only its display names are RFC 2047 decoded and an encoded
// comma or @ in a name can't split off another address.
// A value that is only encoded-words, which a MIME encoder writes when
// an address part is not ASCII, is decoded as a whole, because its raw
// form holds no address that could be read.
func addressHeader(raw string) string {
	if !encodedWordsRegexp.MatchString(raw) {
		return raw
	}
	decoded, err := mimeHeaderDecoder.DecodeHeader(raw)
	if err != nil {
		return raw
	}
	return decoded
}

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
	// The address headers are read raw instead of RFC 2047 decoded like by
	// envelope.GetHeader, so that the parser decodes only display names,
	// and an encoded comma or @ in a name can't split off another address
	header := envelope.Root.Header
	msg.From, err = NormalizedAddress(addressHeader(header.Get("From")))
	if err != nil {
		parseErrs = append(parseErrs, fmt.Errorf("can't parse email header 'From': %w", err))
		msg.From = ""
	}
	msg.ReplyTo, err = NormalizedNullableAddress(addressHeader(header.Get("Reply-To")))
	if err != nil {
		// intentionally ignoring parsing issues with Reply-To, unset the value.
		// we've seen weird and unparsable values
		msg.ReplyTo = NullableAddress("")
	}
	for _, to := range header.Values("To") {
		// Split returns all parsable addresses even when others fail,
		// so keep the usable ones and collect the error.
		addrs, err := AddressList(addressHeader(to)).SplitNormalized()
		if err != nil {
			parseErrs = append(parseErrs, fmt.Errorf("can't parse email header 'To': %w", err))
		}
		msg.To = msg.To.Append(addrs...)
	}
	for _, deliveredTo := range header.Values("Delivered-To") {
		// intentionally ignoring parsing issues with nullable lists
		// (we've seen weird and unparsable values) but keep the
		// addresses that could be parsed
		addrs, _ := AddressList(addressHeader(deliveredTo)).SplitNormalized()
		msg.DeliveredTo = msg.DeliveredTo.Append(addrs...)
	}
	for _, cc := range header.Values("Cc") {
		// intentionally ignoring parsing issues with nullable lists
		// (we've seen weird and unparsable values) but keep the
		// addresses that could be parsed
		addrs, _ := AddressList(addressHeader(cc)).SplitNormalized()
		msg.Cc = msg.Cc.Append(addrs...)
	}
	for _, bcc := range header.Values("Bcc") {
		// intentionally ignoring parsing issues with nullable lists
		// (we've seen weird and unparsable values) but keep the
		// addresses that could be parsed
		addrs, _ := AddressList(addressHeader(bcc)).SplitNormalized()
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

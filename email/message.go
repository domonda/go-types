package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/mail"
	"net/textproto"
	"strings"
	txttemplate "text/template"
	"time"

	"github.com/jhillyerd/enmime"
	"github.com/ungerik/go-fs"

	"github.com/domonda/go-errs"
	"github.com/domonda/go-types/nullable"
)

type Header = textproto.MIMEHeader

var parsedMessageHeaders = map[string]struct{}{
	"Message-Id":  {},
	"In-Reply-To": {},
	"References":  {},
	"Date":        {},
	"From":        {},
	"Reply-To":    {},
	"To":          {},
	"Cc":          {},
	"Bcc":         {},
	"Subject":     {},
}

func IsParsedHeader(key string) bool {
	_, is := parsedMessageHeaders[textproto.CanonicalMIMEHeaderKey(key)]
	return is
}

func IsExtraHeader(key string) bool {
	return !IsParsedHeader(key)
}

type Message struct {
	// ProviderID is the optional ID of the message
	// at the email provider like GMail that might be
	// different from the RFC 822 Message-ID.
	ProviderID nullable.TrimmedString `json:"providerID,omitempty"`

	// InReplyToProviderID is the ProviderID of the message
	// that this message is a reply to.
	InReplyToProviderID nullable.TrimmedString `json:"inReplyToProviderID,omitempty"`

	// ProviderLabels are optional labels from the email provider
	// like GMail that are not encoded in the essage itself.
	ProviderLabels []string `json:"providerLabels,omitempty"`

	// MessageID is the "Message-ID" header according to RFC 822/2822/5322.
	// Find in Gmail via filter: rfc822msgid:MessageID
	MessageID nullable.TrimmedString `json:"messageID,omitempty"`

	// In-Reply-To header
	InReplyTo nullable.TrimmedString `json:"inReplyTo,omitempty"`

	// References header
	References nullable.TrimmedString `json:"references,omitempty"`

	// Date header
	Date *time.Time `json:"date,omitempty"`

	From    Address             `json:"from,omitempty"`
	ReplyTo NullableAddress     `json:"replyTo,omitempty"`
	To      AddressList         `json:"to,omitempty"`
	Cc      NullableAddressList `json:"cc,omitempty"`
	Bcc     NullableAddressList `json:"bcc,omitempty"`

	// ExtraHeader can be used for additional header data
	// not covered by the other fields of the struct.
	ExtraHeader Header `json:"extraHeader,omitempty"`

	Subject string `json:"subject,omitempty"`

	// Body is the plaintext body of the email.
	Body string `json:"body,omitempty"`

	// BodyHTML returns the HTML body if available.
	BodyHTML nullable.TrimmedString `json:"bodyHTML,omitempty"`

	Attachments []*Attachment `json:"attachments,omitempty"`
}

// NewMessage returns a new message using the passed from, to, subject, body, and bodyHTML arguments.
func NewMessage(from Address, to AddressList, subject, body string, bodyHTML nullable.TrimmedString) *Message {
	return &Message{
		From:        from,
		To:          to,
		Subject:     subject,
		Body:        body,
		BodyHTML:    bodyHTML,
		ExtraHeader: make(Header),
	}
}

// Recipients returns the valid, normalized, name stripped,
// deduplicated addresses from the To, Cc, and Bcc fields.
func (msg *Message) Recipients() []string {
	var recipients []string
	exists := func(a *mail.Address) bool {
		for _, r := range recipients {
			if r == a.Address {
				return true
			}
		}
		return false
	}
	to, _ := msg.To.Parse()
	for _, a := range to {
		if !exists(a) {
			recipients = append(recipients, a.Address)
		}
	}
	cc, _ := msg.Cc.Parse()
	for _, a := range cc {
		if !exists(a) {
			recipients = append(recipients, a.Address)
		}
	}
	bcc, _ := msg.Bcc.Parse()
	for _, a := range bcc {
		if !exists(a) {
			recipients = append(recipients, a.Address)
		}
	}
	return recipients
}

// ReferencesMessageIDs returns the message IDs listed in the References header.
func (msg *Message) ReferencesMessageIDs() []string {
	if msg.References.IsNull() {
		return nil
	}
	var ids []string
	for _, id := range strings.Split(string(msg.References), ",") {
		if id = strings.TrimSpace(id); id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func ParseMessageFile(ctx context.Context, file fs.FileReader) (msg *Message, err error) {
	defer errs.WrapWithFuncParams(&err, ctx, file)

	data, err := file.ReadAllContext(ctx)
	if err != nil {
		return nil, err
	}

	return ParseMessage(data)
}

func ParseMessage(data []byte) (msg *Message, err error) {
	defer errs.WrapWithFuncParams(&err, data)

	if len(data) == 0 {
		return nil, errs.New("no message data")
	}

	// Fast check for JSON object
	if data[0] == '{' && data[len(data)-1] == '}' {
		err = json.Unmarshal(data, &msg)
		if err != nil {
			return nil, err
		}
		return msg, nil
	}

	// Fast check of first 4 bytes for TNEF signature
	tnefMessage, err := ParseTNEFMessageBytes(data)
	if err == nil {
		return tnefMessage, nil
	}

	return ParseMIMEMessageBytes(data)
}

// ReplyToAddress returns the ReplyTo address if available,
// else the From address.
func (msg *Message) ReplyToAddress() Address {
	if msg.ReplyTo.IsNull() {
		return msg.From
	}
	return msg.ReplyTo.Get()
}

// DeliveredTo returns the normalized address part
// of the "Delivered-To" header if available.
func (msg *Message) DeliveredTo() NullableAddress {
	addr, _ := NullableAddress(msg.ExtraHeader.Get("Delivered-To")).AddressPart()
	return addr
}

// Returns if the "Auto-Submitted" header is set
// and has a different value than "no".
// See RFC 3834: https://datatracker.ietf.org/doc/html/rfc3834
func (msg *Message) IsAutoSubmitted() bool {
	as := msg.ExtraHeader.Get("Auto-Submitted")
	return as != "" && as != "no"
}

// Returns if the X-Auto-Response-Suppress header
// contains any of the values "DR", "AutoReply", or "All".
// See https://docs.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxcmail/ced68690-498a-4567-9d14-5c01f974d8b1
func (msg *Message) AutoResponseSuppress() bool {
	values := msg.ExtraHeader.Values("X-Auto-Response-Suppress")
	for _, value := range values {
		switch value {
		case "DR", "AutoReply", "All":
			return true
		}
	}
	return false
}

// ListID returns the value of the "List-Id" header
// or an empty string if not available.
func (msg *Message) ListID() string {
	return msg.ExtraHeader.Get("List-Id")
}

// FeedbackID returns the value of the "Feedback-Id" header
// or an empty string if not available.
func (msg *Message) FeedbackID() string {
	return msg.ExtraHeader.Get("Feedback-Id")
}

func (msg *Message) String() string {
	return fmt.Sprintf(
		"Message{Subject: %q, From: %s, DeliveredTo: %s, MessageID: %s, ProviderID: %s, Labels: %q}",
		msg.Subject,
		msg.From,
		msg.DeliveredTo(),
		msg.MessageID,
		msg.ProviderID,
		msg.ProviderLabels,
	)
}

func (msg *Message) AddAttachment(filename string, content []byte) {
	msg.Attachments = append(msg.Attachments, NewAttachment(filename, content))
}

func (msg *Message) AddAttachmentReadFile(ctx context.Context, file fs.FileReader) error {
	attachment, err := NewAttachmentReadFile(ctx, file)
	if err != nil {
		return err
	}
	msg.Attachments = append(msg.Attachments, attachment)
	return nil
}

type ReplyTemplateData struct {
	Message
	Date      string
	BodyLines []string
	BodyHTML  template.HTML
	ReplyText string
	ReplyHTML template.HTML
}

// NewReplyMessage creates a reply based on an existing message.
// The textTempl and htmlTempl templates will called with
// a ReplyTemplateData struct instance as context
// and are responsible to render the passed replyText and replyHTML
// together with a quotation of the original message.
// The passed replyText and replyHTML are not interpreted as templates.
func (msg *Message) NewReplyMessage(from Address, replyText, replyHTML string, keepAttachments bool, textTempl, htmlTempl string) (re *Message, err error) {
	defer errs.WrapWithFuncParams(&err, from, replyText, replyHTML, keepAttachments, textTempl, htmlTempl)

	re = &Message{
		InReplyToProviderID: msg.ProviderID,
		InReplyTo:           msg.MessageID,
		References:          msg.MessageID,
		From:                from,
		Subject:             "Re: " + msg.Subject,
		ExtraHeader:         make(Header),
	}
	if msg.ReplyTo.IsNotNull() {
		re.To = msg.ReplyTo.Get().AsList()
	} else {
		re.To = msg.From.AsList()
	}
	if keepAttachments {
		re.Attachments = msg.Attachments
	}

	//#nosec G203 -- not escaped HTML OK
	data := &ReplyTemplateData{
		Message:   *msg,
		Date:      formatDate(msg.Date),
		BodyLines: strings.Split(msg.Body, "\n"),
		BodyHTML:  template.HTML(msg.BodyHTML),
		ReplyText: replyText,
		ReplyHTML: template.HTML(replyHTML),
	}
	if data.BodyHTML == "" {
		data.BodyHTML = plaintextToHTML(msg.Body)
	}
	if data.ReplyHTML == "" {
		data.ReplyHTML = plaintextToHTML(replyText)
	}

	var b strings.Builder
	textTemplate, err := txttemplate.New("text").Parse(textTempl)
	if err != nil {
		return nil, err
	}
	err = textTemplate.Execute(&b, data)
	if err != nil {
		return nil, err
	}
	re.Body = b.String()

	if msg.BodyHTML.IsNotNull() || replyHTML != "" {
		var b strings.Builder
		htmlTemplate, err := template.New("html").Parse(htmlTempl)
		if err != nil {
			return nil, err
		}
		err = htmlTemplate.Execute(&b, data)
		if err != nil {
			return nil, err
		}
		re.BodyHTML.Set(b.String())
	}

	return re, nil
}

func (msg *Message) BuildRawMessage() (raw []byte, err error) {
	defer errs.WrapWithFuncParams(&err)

	// Fully loaded structure; the presence of text, html, inlines, and attachments will determine
	// how much is necessary:
	//
	//  multipart/mixed
	//  |- multipart/related
	//  |  |- multipart/alternative
	//  |  |  |- text/plain
	//  |  |  `- text/html
	//  |  `- inlines..
	//  `- attachments..
	//
	// We build this tree starting at the leaves, re-rooting as needed.
	var (
		root *enmime.Part
		part *enmime.Part
	)
	if msg.Body != "" || msg.BodyHTML == "" {
		root = enmime.NewPart("text/plain")
		root.Content = []byte(msg.Body)
		root.Charset = "utf-8"
	}
	if msg.BodyHTML != "" {
		part = enmime.NewPart("text/html")
		part.Content = []byte(msg.BodyHTML)
		part.Charset = "utf-8"
		if root == nil {
			root = part
		} else {
			root.NextSibling = part
		}
	}
	if msg.Body != "" && msg.BodyHTML != "" {
		// Wrap Text & HTML bodies
		part = root
		root = enmime.NewPart("multipart/alternative")
		root.AddChild(part)
	}
	// if len(b.inlines) > 0 {
	// 	part = root
	// 	root = enmime.NewPart("multipart/related")
	// 	root.AddChild(part)
	// 	for _, ip := range b.inlines {
	// part := enmime.NewPart(contentType)
	// part.Content = content
	// part.FileName = fileName
	// part.Disposition = "inline"
	// part.ContentID = contentID
	// 		part.Header = make(textproto.MIMEHeader)
	// 		root.AddChild(part)
	// 	}
	// }
	if len(msg.Attachments) > 0 {
		part = root
		root = enmime.NewPart("multipart/mixed")
		root.AddChild(part)
		for _, att := range msg.Attachments {
			part := enmime.NewPart(att.ContentType)
			part.Content = att.File.FileData
			part.FileName = att.File.FileName
			part.Disposition = "attachment"
			root.AddChild(part)
		}
	}
	// Headers
	root.Header.Set("MIME-Version", "1.0")
	if msg.MessageID.IsNotNull() {
		root.Header.Set("Message-Id", msg.MessageID.Get())
	}
	if msg.InReplyTo.IsNotNull() {
		root.Header.Set("In-Reply-To", msg.InReplyTo.Get())
	}
	if msg.References.IsNotNull() {
		root.Header.Set("References", msg.References.Get())
	}
	root.Header.Set("From", string(msg.From))
	if msg.ReplyTo.IsNotNull() {
		root.Header.Set("Reply-To", string(msg.ReplyTo))
	}
	tos, err := msg.To.Split()
	if err != nil {
		return nil, err
	}
	for _, to := range tos {
		root.Header.Add("To", string(to))
	}
	ccs, err := msg.Cc.Split()
	if err != nil {
		return nil, err
	}
	for _, cc := range ccs {
		root.Header.Set("Cc", string(cc))
	}
	bccs, err := msg.Bcc.Split()
	if err != nil {
		return nil, err
	}
	for _, bcc := range bccs {
		root.Header.Set("Bcc", string(bcc))
	}
	root.Header.Set("Date", formatDate(msg.Date))
	for key, vals := range msg.ExtraHeader {
		for _, val := range vals {
			root.Header.Add(key, val)
		}
	}
	root.Header.Set("Subject", strings.TrimSpace(msg.Subject))

	var buf bytes.Buffer
	err = root.Encode(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

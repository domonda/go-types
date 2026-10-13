package email_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/mail"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/domonda/go-types/email"
	"github.com/domonda/go-types/nullable"
	"github.com/stretchr/testify/require"
)

func TestParseMessage(t *testing.T) {
	bytes, err := os.ReadFile("examplemessage_test.eml")
	require.NoError(t, err)

	msg, err := email.ParseMessage(bytes)
	require.NoError(t, err)

	equalsExamplemessage(t, msg, false)
}

func TestBuildRawMessage(t *testing.T) {
	// original
	bytes, err := os.ReadFile("examplemessage_test.eml")
	require.NoError(t, err)
	msg, err := email.ParseMessage(bytes)
	require.NoError(t, err)

	// build raw message from original
	bytes, err = msg.BuildRawMessage()
	require.NoError(t, err)

	// parse built raw message
	msg, err = email.ParseMessage(bytes)
	require.NoError(t, err)

	equalsExamplemessage(t, msg, true)
}

func equalsExamplemessage(t *testing.T, actual *email.Message, built bool) {
	actualAttachments := actual.Attachments
	actual.Attachments = nil // remove attachments for comparison, we'll only compare the hashes

	examplemessageDate, _ := time.Parse(time.RFC1123Z, "Thu, 22 Jan 2026 16:21:11 +0100")
	expected := &email.Message{
		MessageID:   nullable.TrimmedStringFrom("<94D8B6DE-E571-4F66-AAA7-2733E5A12434@denelop.com>"),
		From:        email.Address(`"John Doe" <john@doe.com>`),
		To:          email.AddressListJoinStrings("myinternal@invoicing.com"),
		DeliveredTo: email.NullableAddressListJoinStrings("denis+domonda@domonda.com", "somewhere+domonda@domonda.com"),
		ReplyTo:     email.NullableAddress(`"Jane Doe" <jane@doe.com>`),
		Subject:     "ok this is a test file check it out",
		Cc:          email.NullableAddressListJoinStrings(`"Mike Doe" <mike@doe.com>`),
		Date:        &examplemessageDate,
		Body:        "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque egestas eu arcu sit amet luctus. Phasellus ultrices velit et sem rhoncus, nec sollicitudin diam tristique. Pellentesque vehicula nec odio sit amet blandit. Praesent dictum consequat justo ut euismod. Aliquam varius rhoncus sem, ac semper elit eleifend et. Aliquam efficitur imperdiet ipsum, eget vestibulum tortor mollis ut. Nullam eleifend tellus eget massa iaculis, vitae facilisis elit malesuada. Morbi vel volutpat ex, sed ornare enim.\n\nOne\nTwo\nThree\nFour five\n\n￼\n￼￼\n￼￼\n\n", // contains unicode object replacement character U+FFFC, vscode might warn but it's correct
		BodyHTML:    nullable.TrimmedStringFrom(`<html aria-label="message body"><head><meta http-equiv="content-type" content="text/html; charset=us-ascii"></head><body style="overflow-wrap: break-word; -webkit-nbsp-mode: space; line-break: after-white-space;"><div><b>Lorem</b> ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque egestas eu arcu sit amet luctus. Phasellus ultrices velit<i> et sem</i> <span style="font-size: 36px;">rhoncus</span>, nec sollicitudin diam tristique. Pellentesque vehicula nec odio sit amet blandit. Praesent dictum consequat<b><i><u> justo ut euismod. Aliquam</u></i></b> <font color="#ffff9a">varius</font> rhoncu<strike>s sem, ac se</strike>mper elit eleifend et. Aliquam efficitur imperdiet ipsum, eget vestibulum tortor mollis ut. Nullam eleifend tellus eget massa iaculis, vitae facilisis elit malesuada. Morbi vel volutpat ex, sed ornare enim.</div><div><br></div><div><ul class="MailOutline"><li>One</li><li>Two</li><ol><li>Three</li><li>Four <b><i>five</i></b></li></ol></ul><div><b><i><br></i></b></div></div><div><b><i></i></b></div></body></html>`),
		ExtraHeader: email.Header{
			"Content-Type":    []string{"multipart/alternative; boundary=\"Apple-Mail=_766E5EE1-F3AA-4EA6-B1FB-3FE00B9843A2\""},
			"Mime-Version":    []string{"1.0 (Mac OS X Mail 16.0 \\(3864.300.41.1.7\\))"},
			"Received":        []string{"by XXXX:XXXX:XXXX:XXXX with SMTP id nb37csp577720ejc; Thu, 22 Jan 2026 07:21:29 -0800 (PST)"},
			"Return-Path":     []string{"<john@doe.com>"},
			"X-Mailer":        []string{"Apple Mail (2.3864.300.41.1.7)"},
			"X-Received":      []string{"by XXXX:XXXX:XXXX:XXXX with SMTP id ffacd0b85a97d-43569bbec67mr29502326f8f.34.1769095286652; Thu, 22 Jan 2026 07:21:26 -0800 (PST)"},
			"X-Original-From": []string{"NEBILY | Helena Törö <h.toeroe@nebily.com>"},
		},
	}

	// when building the message again, some headers are different
	if built {
		// content-type has an id that is newly generated
		expected.ExtraHeader["Content-Type"] = actual.ExtraHeader["Content-Type"]

		// mime-version is added by the builder
		expected.ExtraHeader["Mime-Version"] = append([]string{"1.0"}, expected.ExtraHeader["Mime-Version"]...)

		// body is carriage-return normalized by the builder
		expected.Body = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque egestas eu arcu sit amet luctus. Phasellus ultrices velit et sem rhoncus, nec sollicitudin diam tristique. Pellentesque vehicula nec odio sit amet blandit. Praesent dictum consequat justo ut euismod. Aliquam varius rhoncus sem, ac semper elit eleifend et. Aliquam efficitur imperdiet ipsum, eget vestibulum tortor mollis ut. Nullam eleifend tellus eget massa iaculis, vitae facilisis elit malesuada. Morbi vel volutpat ex, sed ornare enim.\r\n\r\nOne\r\nTwo\r\nThree\r\nFour five\r\n\r\n￼\r\n￼￼\r\n￼￼\r\n\r\n"
	}

	require.Equal(t, expected, actual)

	require.Len(t, actualAttachments, 5)

	// SHA-256 of each attachment's decoded content
	expectedHashes := map[string]string{
		"more invoices and stuff.zip":       "2002c7fc676e23732a28e806a8424b41ecf842f7d91f11c96499891a7d706399",
		"ReadMe.txt":                        "458f19173e7ed372d7c550a9b79f535f4a7265222a41ab8f75b1d817ed4283e8",
		"PastedGraphic-1.png":               "eeb05b27c3e19956cd6d46da06969aef01435568380095eb7d6196c330507411",
		"invoice_Yoseph Carroll_31061.tiff": "541c00ccbb8b1ad7f2ff898d2dd6b5896c38f181b55b6867aaa0d71296b47d92",
		"invoice_Aaron Bergman_36258.pdf":   "2e8206cd45c73701246757a641013aac483b4d58a9ee7ac3695c6f4b167c0101",
	}
	for _, attachment := range actualAttachments {
		hash := sha256.Sum256(attachment.Content)
		require.Equal(t, expectedHashes[attachment.Filename], hex.EncodeToString(hash[:]))
	}
}

func TestNewMessage(t *testing.T) {
	msg := email.NewMessage("me@example.com", "you@example.com", "Subject", "Body", nullable.TrimmedStringFrom("<p>Body</p>"))
	require.Equal(t, email.Address("me@example.com"), msg.From)
	require.Equal(t, email.AddressList("you@example.com"), msg.To)
	require.Equal(t, "Subject", msg.Subject)
	require.Equal(t, "Body", msg.Body)
	require.Equal(t, nullable.TrimmedString("<p>Body</p>"), msg.BodyHTML)
	// ExtraHeader must be initialized so headers can be added without a nil map panic.
	require.NotPanics(t, func() { msg.ExtraHeader.Set("X-Test", "value") })
}

// TestMessage_Recipients protects that every recipient gets the message
// exactly once: addresses from To, Cc and Bcc are normalized to lower case
// bare addresses and deduplicated across the fields, and an unparsable
// entry must not prevent sending to the valid recipients.
func TestMessage_Recipients(t *testing.T) {
	msg := &email.Message{
		To:  `John <John@Example.com>, jane@example.com`,
		Cc:  `"J. Doe" <john@example.com>, not an address, cc@example.com`,
		Bcc: `JANE@example.com, bcc@example.com`,
	}
	require.Equal(t,
		[]string{"john@example.com", "jane@example.com", "cc@example.com", "bcc@example.com"},
		msg.Recipients(),
	)

	// A quoted local part must be compared in its formatted, quoted form,
	// else the same mailbox from To, Cc and Bcc gets the message 3 times.
	msg = &email.Message{
		To:  `"a b"@example.com`,
		Cc:  `X <"A B"@Example.com>`,
		Bcc: `"a b"@example.com, c@example.com`,
	}
	require.Equal(t, []string{`"a b"@example.com`, "c@example.com"}, msg.Recipients())

	require.Empty(t, (&email.Message{}).Recipients(), "a message without recipients")
}

func TestMessage_ReferencesMessageIDs(t *testing.T) {
	tests := []struct {
		name       string
		references nullable.TrimmedString
		want       []string
	}{
		{name: "NULL", references: "", want: nil},
		{name: "single", references: "<a@example.com>", want: []string{"<a@example.com>"}},
		// Empty entries from stray separators must not become message IDs.
		{name: "comma separated", references: "<a@example.com>, ,<b@example.com>,", want: []string{"<a@example.com>", "<b@example.com>"}},
		// RFC 5322 3.6.4 separates the message IDs with whitespace (CFWS),
		// which is what mail clients send, including folded header lines.
		// Without splitting on whitespace the whole thread history would
		// be returned as one unusable message ID.
		{name: "space separated", references: "<a@example.com> <b@example.com>", want: []string{"<a@example.com>", "<b@example.com>"}},
		{name: "folded", references: "<a@example.com>\r\n <b@example.com>\t<c@example.com>", want: []string{"<a@example.com>", "<b@example.com>", "<c@example.com>"}},
		{name: "mixed separators", references: "<a@example.com>, <b@example.com> <c@example.com>", want: []string{"<a@example.com>", "<b@example.com>", "<c@example.com>"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &email.Message{References: tt.references}
			require.Equal(t, tt.want, msg.ReferencesMessageIDs())
		})
	}
}

// TestMessage_ReplyToAddress protects that replies go to the Reply-To
// address requested by the sender and only fall back to From without it.
func TestMessage_ReplyToAddress(t *testing.T) {
	msg := &email.Message{From: "sender@example.com"}
	require.Equal(t, email.Address("sender@example.com"), msg.ReplyToAddress())

	msg.ReplyTo = "Support <support@example.com>"
	require.Equal(t, email.Address("Support <support@example.com>"), msg.ReplyToAddress())
}

// TestMessage_IsAutoSubmitted protects auto-reply loop prevention per RFC 3834:
// any Auto-Submitted value except "no" marks a message as automatic.
func TestMessage_IsAutoSubmitted(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{value: "", want: false},
		{value: "no", want: false},
		{value: "auto-replied", want: true},
		{value: "auto-generated", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			msg := &email.Message{ExtraHeader: make(email.Header)}
			if tt.value != "" {
				msg.ExtraHeader.Set("Auto-Submitted", tt.value)
			}
			require.Equal(t, tt.want, msg.IsAutoSubmitted())
		})
	}
}

// TestMessage_AutoResponseSuppress protects that no automatic reply is sent
// when the sender (typically Exchange) asks to suppress auto replies.
func TestMessage_AutoResponseSuppress(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   bool
	}{
		{name: "no header", values: nil, want: false},
		{name: "All", values: []string{"All"}, want: true},
		{name: "DR", values: []string{"DR"}, want: true},
		{name: "AutoReply", values: []string{"AutoReply"}, want: true},
		// OOF only suppresses out of office replies.
		{name: "OOF", values: []string{"OOF"}, want: false},
		{name: "one of multiple headers", values: []string{"OOF", "AutoReply"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &email.Message{ExtraHeader: make(email.Header)}
			for _, value := range tt.values {
				msg.ExtraHeader.Add("X-Auto-Response-Suppress", value)
			}
			require.Equal(t, tt.want, msg.AutoResponseSuppress())
		})
	}
}

// TestMessage_ExtraHeaderAccessors protects that the header helpers work on
// parsed messages, where header names are case-insensitive and stored in
// ExtraHeader in canonical form.
func TestMessage_ExtraHeaderAccessors(t *testing.T) {
	msg, err := email.ParseMessage(rawMessage(
		`From: newsletter@example.com`,
		`To: me@example.com`,
		`Subject: News`,
		`Date: Thu, 22 Jan 2026 16:21:11 +0100`,
		`AUTO-SUBMITTED: auto-generated`,
		`x-auto-response-suppress: All`,
		`List-ID: <news.example.com>`,
		`Feedback-ID: 123:campaign:example`,
		`Content-Type: text/plain; charset=utf-8`,
		``,
		`Body`,
	))
	require.NoError(t, err)
	require.True(t, msg.IsAutoSubmitted())
	require.True(t, msg.AutoResponseSuppress())
	require.Equal(t, "<news.example.com>", msg.ListID())
	require.Equal(t, "123:campaign:example", msg.FeedbackID())

	empty := &email.Message{}
	require.False(t, empty.IsAutoSubmitted(), "nil ExtraHeader must be handled")
	require.False(t, empty.AutoResponseSuppress())
	require.Empty(t, empty.ListID())
	require.Empty(t, empty.FeedbackID())
}

func TestMessage_String(t *testing.T) {
	// The string is used for logging and must identify the message
	// without dumping the body.
	msg := &email.Message{
		Subject:    "Invoice 42",
		From:       "sender@example.com",
		MessageID:  "<id@example.com>",
		ProviderID: "provider-id",
		Body:       "SECRET BODY",
	}
	s := msg.String()
	require.Contains(t, s, "Invoice 42")
	require.Contains(t, s, "sender@example.com")
	require.Contains(t, s, "<id@example.com>")
	require.Contains(t, s, "provider-id")
	require.NotContains(t, s, "SECRET BODY")
}

func TestMessage_AddAttachment(t *testing.T) {
	msg := &email.Message{}
	msg.AddAttachment("1", "invoice.pdf", []byte("%PDF-1.7"))
	msg.AddAttachment("2", "readme.txt", []byte("Hello"))
	require.Len(t, msg.Attachments, 2)
	require.Equal(t, "invoice.pdf", msg.Attachments[0].Filename)
	require.Equal(t, "application/pdf", msg.Attachments[0].ContentType)
	require.Equal(t, "readme.txt", msg.Attachments[1].Filename)
}

func TestParseMessage_Empty(t *testing.T) {
	msg, err := email.ParseMessage(nil)
	require.Error(t, err)
	require.Nil(t, msg)
}

// TestParseMessage_JSON protects that a Message serialized as JSON
// (as stored or passed between services) is parsed back unchanged
// instead of being misinterpreted as MIME message.
func TestParseMessage_JSON(t *testing.T) {
	date := time.Date(2026, 1, 22, 15, 21, 11, 0, time.UTC)
	want := &email.Message{
		ProviderID:     "provider-id",
		ProviderLabels: []string{"INBOX", "UNREAD"},
		MessageID:      "<id@example.com>",
		InReplyTo:      "<parent@example.com>",
		References:     "<parent@example.com>",
		Date:           &date,
		From:           `"Doe, John" <john@example.com>`,
		ReplyTo:        "reply@example.com",
		To:             "a@example.com, b@example.com",
		Cc:             "cc@example.com",
		ExtraHeader:    email.Header{"X-Custom": {"value"}},
		Subject:        "Subject",
		Body:           "Body",
		BodyHTML:       "<p>Body</p>",
		Attachments: []*email.Attachment{
			{PartID: "1", ContentType: "application/pdf", Filename: "invoice.pdf", Content: []byte("%PDF-1.7")},
		},
	}
	data, err := json.Marshal(want)
	require.NoError(t, err)

	got, err := email.ParseMessage(data)
	require.NoError(t, err)
	require.Equal(t, want, got)

	_, err = email.ParseMessage([]byte(`{"subject": 1}`))
	require.Error(t, err, "invalid JSON must be reported, not parsed as MIME")
}

// TestParseMessage_JSONWhitespace protects that JSON is detected independent
// of surrounding whitespace: json.Encoder terminates its output with a
// newline and stored files or HTTP bodies often carry extra whitespace.
// Such data must be unmarshalled and not be misparsed as MIME message.
func TestParseMessage_JSONWhitespace(t *testing.T) {
	want := &email.Message{
		MessageID: "<id@example.com>",
		From:      "a@example.com",
		To:        "b@example.com",
		Subject:   "Subject",
		Body:      "Body",
	}
	var encoded strings.Builder
	require.NoError(t, json.NewEncoder(&encoded).Encode(want))
	compact, err := json.Marshal(want)
	require.NoError(t, err)

	for name, data := range map[string]string{
		"json.Encoder output":    encoded.String(),
		"surrounding whitespace": " \r\n\t" + string(compact) + "\r\n ",
	} {
		t.Run(name, func(t *testing.T) {
			got, err := email.ParseMessage([]byte(data))
			require.NoError(t, err)
			require.Equal(t, want, got)
		})
	}
}

// replyTextTemplate and replyHTMLTemplate quote the original message
// the way a typical mail client does.
const (
	replyTextTemplate = `{{.ReplyText}}

On {{.Date}} {{.From}} wrote:
{{range .BodyLines}}> {{.}}
{{end}}`

	replyHTMLTemplate = `{{.ReplyHTML}}<blockquote>{{.BodyHTML}}</blockquote>`
)

func newOriginalMessage() *email.Message {
	date := time.Date(2026, 1, 22, 16, 21, 11, 0, time.FixedZone("CET", 3600))
	return &email.Message{
		ProviderID:  "provider-id",
		MessageID:   "<original@example.com>",
		Date:        &date,
		From:        "sender@example.com",
		To:          "me@example.com",
		Subject:     "Invoice",
		Body:        "Line 1\nLine 2",
		ExtraHeader: make(email.Header),
		Attachments: []*email.Attachment{{PartID: "1", Filename: "invoice.pdf", Content: []byte("%PDF-1.7")}},
	}
}

// TestMessage_NewReplyMessage_Threading protects the headers that make mail
// clients show the reply in the same thread and deliver it to the right
// address.
func TestMessage_NewReplyMessage_Threading(t *testing.T) {
	original := newOriginalMessage()

	re, err := original.NewReplyMessage("me@example.com", "Thanks", "", false, replyTextTemplate, replyHTMLTemplate)
	require.NoError(t, err)
	require.Equal(t, email.Address("me@example.com"), re.From)
	require.Equal(t, email.AddressList("sender@example.com"), re.To, "without Reply-To the reply goes to From")
	require.Equal(t, "Re: Invoice", re.Subject)
	require.Equal(t, original.MessageID, re.InReplyTo)
	require.Equal(t, original.MessageID, re.References)
	require.Equal(t, original.ProviderID, re.InReplyToProviderID)
	require.NotNil(t, re.ExtraHeader)
	require.Nil(t, re.Attachments, "attachments are only kept on request")

	original.ReplyTo = `"Support" <support@example.com>`
	re, err = original.NewReplyMessage("me@example.com", "Thanks", "", true, replyTextTemplate, replyHTMLTemplate)
	require.NoError(t, err)
	require.Equal(t, email.AddressList(`"Support" <support@example.com>`), re.To, "Reply-To must be preferred over From")
	require.Equal(t, original.Attachments, re.Attachments)
}

// TestMessage_NewReplyMessage_Subject protects RFC 5322 3.6.5: a reply
// subject carries only one "Re: " prefix. Replying to a reply must not
// stack prefixes to "Re: Re: ...", which clutters the subject with every
// round trip of the conversation.
func TestMessage_NewReplyMessage_Subject(t *testing.T) {
	tests := []struct {
		subject string
		want    string
	}{
		{subject: "Invoice", want: "Re: Invoice"},
		{subject: "", want: "Re: "},
		{subject: "Re: Invoice", want: "Re: Invoice"},
		// Mail clients vary the case and spacing of the prefix.
		{subject: "RE: Invoice", want: "RE: Invoice"},
		{subject: "re:Invoice", want: "re:Invoice"},
		{subject: "  Re: Invoice", want: "  Re: Invoice"},
		// Only the "Re:" prefix counts, not a word starting with "Re".
		{subject: "Reminder", want: "Re: Reminder"},
		{subject: "Re Invoice", want: "Re: Re Invoice"},
	}
	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			original := newOriginalMessage()
			original.Subject = tt.subject
			re, err := original.NewReplyMessage("me@example.com", "Thanks", "", false, replyTextTemplate, replyHTMLTemplate)
			require.NoError(t, err)
			require.Equal(t, tt.want, re.Subject)
		})
	}
}

// TestMessage_NewReplyMessage_References protects RFC 5322 3.6.4: the
// References of a reply are the parent's References followed by the
// parent's Message-ID. Mail clients use the full chain to thread a
// conversation, with only the parent's Message-ID a reply to a reply
// loses its connection to the start of the thread. A parent without
// References but with an In-Reply-To of a single message ID gets that ID
// in place of its References.
func TestMessage_NewReplyMessage_References(t *testing.T) {
	tests := []struct {
		name       string
		messageID  nullable.TrimmedString
		inReplyTo  nullable.TrimmedString
		references nullable.TrimmedString
		want       nullable.TrimmedString
	}{
		{name: "first reply", messageID: "<c@example.com>", want: "<c@example.com>"},
		{name: "reply to reply", messageID: "<c@example.com>", references: "<a@example.com> <b@example.com>", want: "<a@example.com> <b@example.com> <c@example.com>"},
		// The result uses the RFC 5322 whitespace separator.
		{name: "comma separated parent", messageID: "<c@example.com>", references: "<a@example.com>, <b@example.com>", want: "<a@example.com> <b@example.com> <c@example.com>"},
		{name: "parent without Message-ID", references: "<a@example.com> <b@example.com>", want: "<a@example.com> <b@example.com>"},
		{name: "parent without IDs", want: ""},
		{name: "parent with In-Reply-To only", messageID: "<b@example.com>", inReplyTo: "<a@example.com>", want: "<a@example.com> <b@example.com>"},
		{name: "parent with In-Reply-To only, without Message-ID", inReplyTo: "<a@example.com>", want: "<a@example.com>"},
		{name: "References win over In-Reply-To", messageID: "<c@example.com>", inReplyTo: "<b@example.com>", references: "<a@example.com> <b@example.com>", want: "<a@example.com> <b@example.com> <c@example.com>"},
		// RFC 5322 3.6.4 only uses an In-Reply-To with a single message ID
		{name: "parent with multiple In-Reply-To IDs", messageID: "<c@example.com>", inReplyTo: "<a@example.com> <b@example.com>", want: "<c@example.com>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := newOriginalMessage()
			original.MessageID = tt.messageID
			original.InReplyTo = tt.inReplyTo
			original.References = tt.references
			re, err := original.NewReplyMessage("me@example.com", "Thanks", "", false, replyTextTemplate, replyHTMLTemplate)
			require.NoError(t, err)
			require.Equal(t, tt.want, re.References)
			require.Equal(t, tt.messageID, re.InReplyTo)
		})
	}
}

// TestMessage_NewReplyMessage_Body protects how the reply body is rendered:
// the templates quote the original message, the reply text is inserted
// literally (it is user input and must never be executed as template),
// and an HTML body is only created if there is HTML to render.
func TestMessage_NewReplyMessage_Body(t *testing.T) {
	t.Run("plaintext only", func(t *testing.T) {
		original := newOriginalMessage()
		re, err := original.NewReplyMessage("me@example.com", "Thanks {{.Subject}}", "", false, replyTextTemplate, replyHTMLTemplate)
		require.NoError(t, err)
		require.Equal(t,
			"Thanks {{.Subject}}\n\nOn Thu, 22 Jan 2026 16:21:11 +0100 sender@example.com wrote:\n> Line 1\n> Line 2\n",
			re.Body,
		)
		require.True(t, re.BodyHTML.IsNull(), "no HTML body without HTML in the original or the reply")
	})

	t.Run("original has HTML", func(t *testing.T) {
		original := newOriginalMessage()
		original.BodyHTML = "<p>Original</p>"
		re, err := original.NewReplyMessage("me@example.com", "Thanks\nErik", "", false, replyTextTemplate, replyHTMLTemplate)
		require.NoError(t, err)
		// The plaintext reply is converted to HTML keeping its line breaks.
		require.Equal(t, nullable.TrimmedString("Thanks<br>Erik<blockquote><p>Original</p></blockquote>"), re.BodyHTML)
	})

	t.Run("reply has HTML", func(t *testing.T) {
		original := newOriginalMessage()
		re, err := original.NewReplyMessage("me@example.com", "Thanks", "<b>Thanks</b>", false, replyTextTemplate, replyHTMLTemplate)
		require.NoError(t, err)
		// The plaintext original is quoted keeping its line breaks.
		require.Equal(t, nullable.TrimmedString("<b>Thanks</b><blockquote>Line 1<br>Line 2</blockquote>"), re.BodyHTML)
	})
}

func TestMessage_NewReplyMessage_TemplateErrors(t *testing.T) {
	// Broken templates must be reported instead of sending a broken reply.
	tests := []struct {
		name      string
		textTempl string
		htmlTempl string
	}{
		{name: "invalid text template", textTempl: "{{.ReplyText", htmlTempl: replyHTMLTemplate},
		{name: "failing text template", textTempl: "{{.NoSuchField}}", htmlTempl: replyHTMLTemplate},
		{name: "invalid HTML template", textTempl: replyTextTemplate, htmlTempl: "{{.ReplyHTML"},
		{name: "failing HTML template", textTempl: replyTextTemplate, htmlTempl: "{{.NoSuchField}}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := newOriginalMessage().NewReplyMessage("me@example.com", "Thanks", "<b>Thanks</b>", false, tt.textTempl, tt.htmlTempl)
			require.Error(t, err)
			require.Nil(t, re)
		})
	}
}

// TestBuildRawMessage_RoundTrip protects that all header fields and body
// variants survive BuildRawMessage and parsing the result again.
func TestBuildRawMessage_RoundTrip(t *testing.T) {
	date := time.Date(2026, 1, 22, 16, 21, 11, 0, time.FixedZone("CET", 3600))
	msg := &email.Message{
		MessageID:   "<id@example.com>",
		InReplyTo:   "<parent@example.com>",
		References:  "<parent@example.com>",
		Date:        &date,
		From:        `"John Doe" <john@example.com>`,
		ReplyTo:     "reply@example.com",
		To:          "a@example.com, b@example.com",
		DeliveredTo: "d1@example.com, d2@example.com",
		Cc:          "cc@example.com",
		Bcc:         "bcc@example.com",
		ExtraHeader: email.Header{"X-Custom": {"value"}},
		Subject:     "  Subject with spaces  ",
		Body:        "Body",
	}
	raw, err := msg.BuildRawMessage()
	require.NoError(t, err)

	parsed, err := email.ParseMessage(raw)
	require.NoError(t, err)
	require.Equal(t, msg.MessageID, parsed.MessageID)
	require.Equal(t, msg.InReplyTo, parsed.InReplyTo)
	require.Equal(t, msg.References, parsed.References)
	require.True(t, date.Equal(*parsed.Date), "Date %s != %s", parsed.Date, date)
	require.Equal(t, msg.From, parsed.From)
	require.Equal(t, msg.ReplyTo, parsed.ReplyTo)
	require.Equal(t, msg.To, parsed.To)
	require.Equal(t, msg.DeliveredTo, parsed.DeliveredTo)
	require.Equal(t, msg.Cc, parsed.Cc)
	require.Equal(t, msg.Bcc, parsed.Bcc)
	require.Equal(t, "value", parsed.ExtraHeader.Get("X-Custom"))
	require.Equal(t, "Subject with spaces", parsed.Subject, "the subject is trimmed")
	require.Equal(t, "Body", parsed.Body)
	require.True(t, parsed.BodyHTML.IsNull())
}

// TestBuildRawMessage_RoundTripEncodedAddresses protects that the
// package reads its own messages back: names that need RFC 2047 encoding,
// quoted local parts and non-ASCII address parts, for which the MIME
// encoder encodes the whole header value, so that names must not be
// encoded before and the parser must decode such a value as a whole.
func TestBuildRawMessage_RoundTripEncodedAddresses(t *testing.T) {
	for _, msg := range []*email.Message{
		{
			From: `"Jürgen Müller" <j@example.com>`,
			To:   `"N" <"'abc"@example.com>, "Ivan" <"иван"@example.com>, "=?utf-8?q?x?=" <x@example.com>`,
			Body: "Body",
		},
		{
			From: `"Jürgen" <jürgen@müller.de>`,
			To:   `"Jürgen Müller" <j@example.com>, "Ivan" <jürgen@müller.de>`,
			Cc:   `jürgen@müller.de`,
			Body: "Body",
		},
	} {
		raw, err := msg.BuildRawMessage()
		require.NoError(t, err)
		parsed, err := email.ParseMessage(raw)
		require.NoError(t, err)

		wantFrom, err := msg.From.Normalized()
		require.NoError(t, err)
		require.Equal(t, wantFrom, parsed.From, "From of %s", raw)
		wantTo, err := msg.To.Normalized()
		require.NoError(t, err)
		require.Equal(t, wantTo, parsed.To, "To of %s", raw)
		wantCc, err := msg.Cc.Normalized()
		require.NoError(t, err)
		require.Equal(t, wantCc, parsed.Cc, "Cc of %s", raw)
		require.NotContains(t, string(raw), "=3D=3Futf-8", "a name must not be encoded twice")
	}
}

func TestBuildRawMessage_Bodies(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		bodyHTML    nullable.TrimmedString
		attachments []*email.Attachment
	}{
		{name: "empty"},
		{name: "text", body: "Text body"},
		{name: "HTML", bodyHTML: "<p>HTML body</p>"},
		{name: "text and HTML", body: "Text body", bodyHTML: "<p>HTML body</p>"},
		{
			name:     "text, HTML and attachments",
			body:     "Text body",
			bodyHTML: "<p>HTML body</p>",
			attachments: []*email.Attachment{
				{ContentType: "application/pdf", Filename: "invoice.pdf", Content: []byte("%PDF-1.7")},
				{ContentType: "text/plain", Filename: "readme.txt", Content: []byte("Hello")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &email.Message{
				From:        "a@example.com",
				To:          "b@example.com",
				Body:        tt.body,
				BodyHTML:    tt.bodyHTML,
				Attachments: tt.attachments,
			}
			before := time.Now().Truncate(time.Second)
			raw, err := msg.BuildRawMessage()
			require.NoError(t, err)

			parsed, err := email.ParseMessage(raw)
			require.NoError(t, err)
			// A missing date is set to the build time because RFC 5322 requires it.
			require.NotNil(t, parsed.Date)
			require.False(t, parsed.Date.Before(before))

			if tt.body != "" || tt.bodyHTML == "" {
				require.Equal(t, tt.body, parsed.Body)
			}
			require.Equal(t, tt.bodyHTML, parsed.BodyHTML)
			require.Len(t, parsed.Attachments, len(tt.attachments))
			for i, want := range tt.attachments {
				got := parsed.Attachments[i]
				require.Equal(t, want.Filename, got.Filename)
				require.Equal(t, want.ContentType, got.ContentType)
				require.Equal(t, want.Content, got.Content)
			}
		})
	}
}

// TestBuildRawMessage_RecipientHeadersOnce protects RFC 5322 3.6.3:
// the To, Cc and Bcc header fields may each occur at most once and list
// all their addresses comma separated. Overwriting the header per address
// silently sends the message to fewer recipients than intended, and a
// repeated header field makes receivers keep only one of the lists.
func TestBuildRawMessage_RecipientHeadersOnce(t *testing.T) {
	msg := &email.Message{
		From: "from@example.com",
		To:   `a@example.com, "Doe, John" <john@example.com>, c@example.com`,
		Cc:   `cc1@example.com, cc2@example.com`,
		Bcc:  `bcc1@example.com, "Bcc Two" <bcc2@example.com>`,
		Body: "Body",
	}
	raw, err := msg.BuildRawMessage()
	require.NoError(t, err)

	header, err := mail.ReadMessage(strings.NewReader(string(raw)))
	require.NoError(t, err)
	for _, key := range []string{"To", "Cc", "Bcc"} {
		require.Len(t, header.Header[key], 1, "exactly one %s header field", key)
	}

	parsed, err := email.ParseMIMEMessage(strings.NewReader(string(raw)))
	require.NoError(t, err)
	require.Equal(t, msg.To, parsed.To)
	require.Equal(t, msg.Cc, parsed.Cc)
	require.Equal(t, msg.Bcc, parsed.Bcc)
}

// TestBuildRawMessage_NonASCIIFromName protects that the addresses stay
// parseable when their display name needs RFC 2047 encoding. Written raw,
// the whole header value including the <address> became one encoded-word
// and RFC 5322 parsers could not find the address, so replies went
// nowhere. The normalized Address keeps the name as UTF-8, so the
// encoding has to happen when the header is written.
func TestBuildRawMessage_NonASCIIFromName(t *testing.T) {
	msg := &email.Message{
		From:        "Jörg Müller <joerg@example.com>",
		ReplyTo:     "Zoë Bäcker <zoe@example.com>",
		To:          `"Müller, Jürgen" <juergen@example.com>, to@example.com`,
		DeliveredTo: "Zoë <zoe@example.com>",
		Cc:          "Søren <soeren@example.com>",
		Bcc:         "Bjørn <bjoern@example.com>",
		Body:        "Body",
	}
	raw, err := msg.BuildRawMessage()
	require.NoError(t, err)

	parsed, err := mail.ReadMessage(strings.NewReader(string(raw)))
	require.NoError(t, err)
	for key, values := range parsed.Header {
		for _, value := range values {
			for _, r := range value {
				require.Less(t, r, rune(128), "header %s must be ASCII: %s", key, value)
			}
		}
	}
	from, err := mail.ParseAddress(parsed.Header.Get("From"))
	require.NoError(t, err)
	require.Equal(t, "joerg@example.com", from.Address)
	require.Equal(t, "Jörg Müller", from.Name)
	replyTo, err := mail.ParseAddress(parsed.Header.Get("Reply-To"))
	require.NoError(t, err)
	require.Equal(t, "zoe@example.com", replyTo.Address)
	require.Equal(t, "Zoë Bäcker", replyTo.Name)
	to, err := mail.ParseAddressList(parsed.Header.Get("To"))
	require.NoError(t, err)
	require.Len(t, to, 2)
	require.Equal(t, "juergen@example.com", to[0].Address)
	require.Equal(t, "Müller, Jürgen", to[0].Name)
	cc, err := mail.ParseAddress(parsed.Header.Get("Cc"))
	require.NoError(t, err)
	require.Equal(t, "Søren", cc.Name)

	// Parsed back, the addresses are normalized with UTF-8 names
	msgParsed, err := email.ParseMIMEMessage(strings.NewReader(string(raw)))
	require.NoError(t, err)
	require.Equal(t, email.Address(`"Jörg Müller" <joerg@example.com>`), msgParsed.From)
	require.Equal(t, email.NullableAddress(`"Zoë Bäcker" <zoe@example.com>`), msgParsed.ReplyTo)
	require.Equal(t, email.AddressList(`"Müller, Jürgen" <juergen@example.com>, to@example.com`), msgParsed.To)
	require.Equal(t, email.NullableAddressList(`"Zoë" <zoe@example.com>`), msgParsed.DeliveredTo)
	require.Equal(t, email.NullableAddressList(`"Søren" <soeren@example.com>`), msgParsed.Cc)
	require.Equal(t, email.NullableAddressList(`"Bjørn" <bjoern@example.com>`), msgParsed.Bcc)
}

// TestBuildRawMessage_QuotedLocalPart protects that the headers name the
// mailbox that was validated: a quoted local part with special characters
// written without its quotes reads as another address, here at evil.com,
// and Recipients, which sends the message, must keep the quotes too.
func TestBuildRawMessage_QuotedLocalPart(t *testing.T) {
	const addr = `"victim@evil.com>"@example.com`
	msg := &email.Message{From: addr, To: addr, Body: "Body"}
	raw, err := msg.BuildRawMessage()
	require.NoError(t, err)
	require.Contains(t, string(raw), "From: "+addr+"\r\n")
	require.Contains(t, string(raw), "To: "+addr+"\r\n")
	require.Equal(t, []string{addr}, msg.Recipients())

	parsed, err := email.ParseMIMEMessage(strings.NewReader(string(raw)))
	require.NoError(t, err)
	require.Equal(t, email.Address(addr), parsed.From)
	require.Equal(t, []string{addr}, parsed.Recipients())
}

// TestBuildRawMessage_NilAttachment protects that a JSON null in the
// attachments array, which encoding/json decodes as a nil pointer,
// returns an error instead of crashing the caller with a nil dereference.
func TestBuildRawMessage_NilAttachment(t *testing.T) {
	var msg email.Message
	require.NoError(t, json.Unmarshal([]byte(`{"from":"from@example.com","to":"to@example.com","attachments":[null]}`), &msg))
	_, err := msg.BuildRawMessage()
	require.Error(t, err)
}

// TestBuildRawMessage_InlineParts protects that inline parts like images
// referenced from the HTML body via "cid:" stay inline parts with their
// Content-ID. Sent as regular attachments, the HTML body would show broken
// images and mail clients would list the images as extra attached files.
func TestBuildRawMessage_InlineParts(t *testing.T) {
	logo := &email.Attachment{
		ContentID:   "logo@example.com",
		ContentType: "image/png",
		Inline:      true,
		Filename:    "logo.png",
		Content:     []byte("\x89PNG\r\n\x1a\n"),
	}
	invoice := &email.Attachment{
		ContentType: "application/pdf",
		Filename:    "invoice.pdf",
		Content:     []byte("%PDF-1.7"),
	}
	msg := &email.Message{
		From:        "from@example.com",
		To:          "to@example.com",
		Body:        "Body",
		BodyHTML:    `<p>Body</p><img src="cid:logo@example.com">`,
		Attachments: []*email.Attachment{logo, invoice},
	}
	raw, err := msg.BuildRawMessage()
	require.NoError(t, err)
	require.Contains(t, string(raw), "Content-Disposition: inline; filename=logo.png")
	require.Contains(t, string(raw), "Content-Id: <logo@example.com>")

	parsed, err := email.ParseMIMEMessage(strings.NewReader(string(raw)))
	require.NoError(t, err)
	require.Equal(t, msg.Body, parsed.Body)
	require.Equal(t, msg.BodyHTML, parsed.BodyHTML)
	require.Len(t, parsed.Attachments, 2)
	for _, want := range msg.Attachments {
		i := slices.IndexFunc(parsed.Attachments, func(a *email.Attachment) bool { return a.Filename == want.Filename })
		require.NotEqual(t, -1, i, "attachment %s not found", want.Filename)
		got := parsed.Attachments[i]
		require.Equal(t, want.Inline, got.Inline, "Inline of %s", want.Filename)
		require.Equal(t, want.ContentID, got.ContentID, "ContentID of %s", want.Filename)
		require.Equal(t, want.ContentType, got.ContentType, "ContentType of %s", want.Filename)
		require.Equal(t, want.Content, got.Content, "Content of %s", want.Filename)
	}
}

// TestBuildRawMessage_InvalidRecipients protects that a message with
// unparsable recipients is not built, because silently dropping a
// recipient would send the message to fewer people than intended.
// An unparsable From or Reply-To is no recipient, but written into the
// header as given it leaves the receiver no address to reply to,
// so it is an error too instead of a message that can't be answered.
func TestBuildRawMessage_InvalidRecipients(t *testing.T) {
	tests := []struct {
		name string
		msg  email.Message
	}{
		{name: "To", msg: email.Message{To: "a@example.com, not an address"}},
		{name: "Delivered-To", msg: email.Message{To: "a@example.com", DeliveredTo: "not an address"}},
		{name: "Cc", msg: email.Message{To: "a@example.com", Cc: "not an address"}},
		{name: "Bcc", msg: email.Message{To: "a@example.com", Bcc: "not an address"}},
		{name: "From", msg: email.Message{From: "not an address", To: "a@example.com"}},
		{name: "Reply-To", msg: email.Message{To: "a@example.com", ReplyTo: "not an address"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.msg.From == "" {
				tt.msg.From = "from@example.com"
			}
			raw, err := tt.msg.BuildRawMessage()
			require.Error(t, err)
			require.Nil(t, raw)
		})
	}
}

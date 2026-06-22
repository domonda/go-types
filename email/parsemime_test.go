package email_test

import (
	"strings"
	"testing"

	"github.com/domonda/go-types/email"
	"github.com/stretchr/testify/require"
)

// rawMessage joins the passed header/body lines with CRLF as required by RFC 822.
func rawMessage(lines ...string) []byte {
	return []byte(strings.Join(lines, "\r\n"))
}

// TestParseMIMEMessage_NonStdlibCharset is a regression test for the bug where
// a header using a charset outside Go's stdlib mime support (us-ascii, utf-8,
// iso-8859-1) caused the whole message to be dropped with
// "mime: unhandled charset ...".
//
// It exercises the full ParseMIMEMessage pipeline. Note that enmime decodes
// the addressed/structured headers (From, To, Cc, Subject, ...) itself via the
// same x/net charset reader, so those assertions verify end-to-end correctness.
// The extra-header loop in ParseMIMEMessage, however, decodes the *raw*
// envelope.Root.Header values with our mimeHeaderDecoder, so the X-Original-From
// assertion is the one that specifically guards the CharsetReader fix
// (it fails with the plain stdlib mime.WordDecoder).
func TestParseMIMEMessage_NonStdlibCharset(t *testing.T) {
	raw := rawMessage(
		`From: =?iso-8859-2?Q?=A3ukasz?= <lukasz@example.pl>`,   // 0xA3 = Ł in ISO 8859-2
		`To: =?windows-1250?Q?Wa=B3=EAsa?= <walesa@example.pl>`, // 0xB3 = ł, 0xEA = ę in Windows-1250
		`Subject: =?iso-8859-2?Q?=A3ukasz?=`,
		`Date: Thu, 22 Jan 2026 16:21:11 +0100`,
		`X-Original-From: =?iso-8859-2?Q?=A3ukasz?= <lukasz@example.pl>`,
		`X-Bogus-Charset: =?made-up-9999?Q?abc?=`, // unsupported charset must not drop the message
		`Content-Type: text/plain; charset=utf-8`,
		``,
		`Body text.`,
	)

	msg, err := email.ParseMessage(raw)
	require.NoError(t, err, "message with a non-stdlib charset header must not be dropped")

	// From header path, decoded from ISO 8859-2.
	fromName, err := msg.From.NamePart()
	require.NoError(t, err)
	require.Equal(t, "Łukasz", fromName)
	fromAddr, err := msg.From.AddressPartString()
	require.NoError(t, err)
	require.Equal(t, "lukasz@example.pl", fromAddr)

	// To header path, decoded from Windows-1250.
	toAddrs, err := msg.To.Parse()
	require.NoError(t, err)
	require.Len(t, toAddrs, 1)
	require.Equal(t, "Wałęsa", toAddrs[0].Name)
	require.Equal(t, "walesa@example.pl", toAddrs[0].Address)

	// Subject decoded from ISO 8859-2.
	require.Equal(t, "Łukasz", msg.Subject)

	// Extra-header decode loop reads the raw header and decodes it with our
	// mimeHeaderDecoder — this assertion fails without the CharsetReader fix.
	require.Equal(t, []string{`Łukasz <lukasz@example.pl>`}, msg.ExtraHeader["X-Original-From"])

	// An unsupported charset must degrade gracefully: the extra-header loop
	// ignores the decode error and keeps the raw value instead of failing.
	require.Equal(t, []string{`=?made-up-9999?Q?abc?=`}, msg.ExtraHeader["X-Bogus-Charset"])
}

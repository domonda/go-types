package email_test

import (
	"strings"
	"testing"
	"testing/iotest"

	"github.com/domonda/go-errs"

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
// the Subject itself via the same x/net charset reader. The address headers
// (From, To, Cc, ...) and the extra-header loop in ParseMIMEMessage, however,
// decode the *raw* envelope.Root.Header values with our mimeHeaderDecoder,
// so the From, To and X-Original-From assertions guard the CharsetReader fix
// (they fail with the plain stdlib mime.WordDecoder).
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

// TestParseMIMEMessage_EncodedNameIsOneAddress protects that an RFC 2047
// encoded display name is decoded only as a name: decoding the whole header
// before splitting it turned an encoded comma and @ in a name into another
// recipient, which net/mail and mail clients don't see.
func TestParseMIMEMessage_EncodedNameIsOneAddress(t *testing.T) {
	msg, err := email.ParseMessage(rawMessage(
		`From: =?utf-8?q?attacker=40evil=2Ecom=2C?= <sender@example.com>`,
		`To: =?utf-8?q?victim=40evil=2Ecom=2C?= <ok@good.com>, =?UTF-8?Q?M=C3=BCller=2C_J=C3=BCrgen?= <j@example.de>`,
		`Cc: =?utf-8?q?cc=40evil=2Ecom=2C?= <cc@good.com>`,
		`Date: Thu, 22 Jan 2026 16:21:11 +0100`,
		`Content-Type: text/plain; charset=utf-8`,
		``,
		`Body`,
	))
	require.NoError(t, err)
	require.Equal(t, email.Address(`"attacker@evil.com," <sender@example.com>`), msg.From)
	require.Equal(t, email.AddressList(`"victim@evil.com," <ok@good.com>, "Müller, Jürgen" <j@example.de>`), msg.To)
	require.Equal(t, email.NullableAddressList(`"cc@evil.com," <cc@good.com>`), msg.Cc)
	require.Equal(t, []string{"ok@good.com", "j@example.de", "cc@good.com"}, msg.Recipients())
}

// TestParseMIMEMessage_AddressHeadersOfOtherSenders protects that reading
// the raw address headers loses no sender or recipient that the decoded
// headers gave: a name in a charset that can't be decoded drops only the
// name, IANA charset labels like cp850 decode, a value encoded as a whole
// is decoded as a whole, and a name written as comment is decoded too.
//
// A comma and @ written unencoded in an encoded-word of a name are not
// part of the name, because net/mail and mail servers read the comma as
// separator and =?utf-8?q?a@b as another recipient. Reading the name as
// one hid that recipient, so the To header with it is an error now, while
// the addresses that can be parsed are kept, like for other errors.
// An encoded name holding an encoded comma and @ is one name.
func TestParseMIMEMessage_AddressHeadersOfOtherSenders(t *testing.T) {
	msg, err := email.ParseMessage(rawMessage(
		`From: =?utf-8?B?TWF4IE11c3Rlcm1hbm4gPG1heEBleGFtcGxlLmNvbT4=?=`,
		`Reply-To: mueller@example.com (=?utf-8?q?M=C3=BCller?=)`,
		`To: =?utf-8?q?a@b,?= <hidden@example.com>, =?utf-8?q?c?= <shown@example.com>`,
		`Cc: =?cp850?q?M=81ller?= <cc1@example.com>, =?utf-7?q?x?= <cc2@example.com>, cc3@example.com`,
		`Date: Thu, 22 Jan 2026 16:21:11 +0100`,
		`Content-Type: text/plain; charset=utf-8`,
		``,
		`Body`,
	))
	require.ErrorContains(t, err, "can't parse email header 'To'")
	require.Equal(t, email.Address(`"Max Mustermann" <max@example.com>`), msg.From)
	require.Equal(t, email.NullableAddress(`"Müller" <mueller@example.com>`), msg.ReplyTo)
	require.Equal(t, email.AddressList(`"?=" <hidden@example.com>, "c" <shown@example.com>`), msg.To)
	require.Equal(t, email.NullableAddressList(`"Müller" <cc1@example.com>, cc2@example.com, cc3@example.com`), msg.Cc)

	// The encoded-word of a name without unencoded commas and @ is still
	// one name, like the encoded comma and @ of the Müller, Jürgen name
	msg, err = email.ParseMessage(rawMessage(
		`From: sender@example.com`,
		`To: =?utf-8?q?a=40b=2C?= <hidden@example.com>, =?UTF-8?Q?M=C3=BCller=2C_J=C3=BCrgen?= <j@example.de>`,
		`Date: Thu, 22 Jan 2026 16:21:11 +0100`,
		`Content-Type: text/plain; charset=utf-8`,
		``,
		`Body`,
	))
	require.NoError(t, err)
	require.Equal(t, email.AddressList(`"a@b," <hidden@example.com>, "Müller, Jürgen" <j@example.de>`), msg.To)
}

// TestParseMIMEMessage_PartialResult verifies that a message with usable data
// is returned together with the joined header parsing errors instead of being
// dropped entirely when individual headers (Date, From, To) can't be parsed.
func TestParseMIMEMessage_PartialResult(t *testing.T) {
	raw := rawMessage(
		`From: not an email address`,
		`To: good@example.com, this is broken`,
		`Subject: Partial parse`,
		`Date: definitely not a date`,
		`Content-Type: text/plain; charset=utf-8`,
		``,
		`Body text.`,
	)

	msg, err := email.ParseMessage(raw)
	require.Error(t, err, "unparseable Date/From/To headers must surface a joined error")
	require.NotNil(t, msg, "a message with usable data must still be returned")

	// Usable data is preserved instead of dropping the whole message.
	require.Equal(t, "Partial parse", msg.Subject)
	require.Equal(t, "Body text.", strings.TrimSpace(msg.Body))

	// Unparseable single-value headers are left unset.
	require.Empty(t, string(msg.From))
	require.Nil(t, msg.Date)

	// The valid address from the To list is kept even though the list also
	// contained an unparseable entry.
	require.Equal(t, email.AddressList("good@example.com"), msg.To)
}

// TestParseMIMEMessage_InvalidReplyTo protects the documented leniency for
// Reply-To: unparsable values seen in the wild must not fail the message,
// replies then fall back to the From address.
func TestParseMIMEMessage_InvalidReplyTo(t *testing.T) {
	msg, err := email.ParseMessage(rawMessage(
		`From: sender@example.com`,
		`Reply-To: not an address`,
		`To: me@example.com`,
		`Date: Thu, 22 Jan 2026 16:21:11 +0100`,
		`Content-Type: text/plain; charset=utf-8`,
		``,
		`Body`,
	))
	require.NoError(t, err)
	require.True(t, msg.ReplyTo.IsNull())
	require.Equal(t, email.Address("sender@example.com"), msg.ReplyToAddress())
}

func TestParseMIMEMessage_ReadError(t *testing.T) {
	// Read errors must be returned instead of a message with missing data.
	msg, err := email.ParseMIMEMessage(iotest.ErrReader(errs.New("read failed")))
	require.Error(t, err)
	require.Nil(t, msg)
}

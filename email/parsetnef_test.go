package email

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TNEF attribute IDs and MAPI property tags used by the test fixtures,
// see [MS-OXTNEF] and github.com/teamwork/tnef.
const (
	tnefAttachRendData = 0x9002 // starts a new attachment
	tnefAttachTitle    = 0x8010 // attachment file name
	tnefAttachData     = 0x800f // attachment content
	tnefMAPIProps      = 0x9003 // message level MAPI properties

	mapiTypeString = 0x001e // PT_STRING8
	mapiBody       = 0x1000 // PR_BODY
	mapiBodyHTML   = 0x1013 // PR_BODY_HTML
)

// tnefMessage returns a minimal TNEF stream consisting of
// the TNEF signature, a legacy key and the passed attributes.
func tnefMessage(attributes ...[]byte) []byte {
	b := []byte{0x78, 0x9f, 0x3e, 0x22} // signature 0x223e9f78 little endian
	b = append(b, 0x00, 0x00)           // legacy key
	for _, a := range attributes {
		b = append(b, a...)
	}
	return b
}

// tnefAttribute encodes a TNEF attribute as
// level(1) id(2) type(2) length(4) data checksum(2), little endian.
// Type and checksum are not evaluated by the decoder and left zero.
func tnefAttribute(level byte, id uint16, data []byte) []byte {
	b := []byte{level}
	b = binary.LittleEndian.AppendUint16(b, id)
	b = binary.LittleEndian.AppendUint16(b, 0)
	b = binary.LittleEndian.AppendUint32(b, uint32(len(data)))
	b = append(b, data...)
	return append(b, 0x00, 0x00)
}

// tnefAttachment returns the attachment level attributes
// for one attachment with the passed filename and content.
func tnefAttachment(filename string, content []byte) []byte {
	b := tnefAttribute(2, tnefAttachRendData, make([]byte, 14))
	b = append(b, tnefAttribute(2, tnefAttachTitle, append([]byte(filename), 0))...)
	return append(b, tnefAttribute(2, tnefAttachData, content)...)
}

// tnefStringProps returns a message level MAPI properties attribute
// with the passed PT_STRING8 properties, keyed by property ID.
func tnefStringProps(props map[uint16]string) []byte {
	data := binary.LittleEndian.AppendUint32(nil, uint32(len(props)))
	for id, value := range props {
		data = binary.LittleEndian.AppendUint16(data, mapiTypeString)
		data = binary.LittleEndian.AppendUint16(data, id)
		data = binary.LittleEndian.AppendUint32(data, 1) // value count
		data = binary.LittleEndian.AppendUint32(data, uint32(len(value)))
		data = append(data, value...)
		for len(data)%4 != 0 {
			data = append(data, 0) // values are padded to 4 bytes
		}
	}
	return tnefAttribute(1, tnefMAPIProps, data)
}

// TestParseTNEFMessageBytes protects the extraction of the content of
// winmail.dat style messages sent by Outlook: body and attachments must be
// available like for MIME messages, with unique part IDs per attachment.
func TestParseTNEFMessageBytes(t *testing.T) {
	data := tnefMessage(
		tnefStringProps(map[uint16]string{
			mapiBody:     "Please find the invoice attached.",
			mapiBodyHTML: "<p>Please find the invoice attached.</p>",
		}),
		tnefAttachment("invoice.pdf", []byte("%PDF-1.7")),
		tnefAttachment("readme.txt", []byte("Hello")),
	)

	msg, err := ParseTNEFMessageBytes(data)
	require.NoError(t, err)
	assert.Equal(t, "Please find the invoice attached.", msg.Body)
	assert.Equal(t, "<p>Please find the invoice attached.</p>", string(msg.BodyHTML))
	assert.NotNil(t, msg.ExtraHeader, "ExtraHeader must be usable without a nil map check")
	require.Len(t, msg.Attachments, 2)
	assert.Equal(t, &Attachment{PartID: "TNEF0", Filename: "invoice.pdf", Content: []byte("%PDF-1.7")}, msg.Attachments[0])
	assert.Equal(t, &Attachment{PartID: "TNEF1", Filename: "readme.txt", Content: []byte("Hello")}, msg.Attachments[1])
}

// TestParseTNEFMessageBytes_HTMLOnly protects that an Outlook message
// without a usable plain text body gets its Body converted from the HTML
// body, while BodyHTML keeps the original HTML. Consumers that only read
// the plain text Body would otherwise get an empty message, and the
// formatted HTML body would be lost.
func TestParseTNEFMessageBytes_HTMLOnly(t *testing.T) {
	const html = "<p>Please find the <b>invoice</b> attached.</p>"
	tests := []struct {
		name  string
		props map[uint16]string
	}{
		{name: "no plain body", props: map[uint16]string{mapiBodyHTML: html}},
		// A single byte plain body is a placeholder, not the message text.
		{name: "1 byte plain body", props: map[uint16]string{mapiBody: "\n", mapiBodyHTML: html}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := ParseTNEFMessageBytes(tnefMessage(tnefStringProps(tt.props)))
			require.NoError(t, err)
			assert.Equal(t, "Please find the *invoice* attached.", msg.Body)
			assert.Equal(t, html, string(msg.BodyHTML))
		})
	}
}

func TestParseTNEFMessageBytes_NoContent(t *testing.T) {
	// A TNEF stream with just the header is an empty but valid message.
	msg, err := ParseTNEFMessageBytes(tnefMessage())
	require.NoError(t, err)
	assert.Empty(t, msg.Body)
	assert.True(t, msg.BodyHTML.IsNull())
	assert.Empty(t, msg.Attachments)
}

// TestParseTNEFMessageBytes_NotTNEF protects that non TNEF data is rejected
// with an error, because ParseMessage relies on that error to fall back to
// parsing the data as MIME message.
func TestParseTNEFMessageBytes_NotTNEF(t *testing.T) {
	for name, data := range map[string][]byte{
		"nil":         nil,
		"too short":   {0x78, 0x9f, 0x3e},
		"MIME":        []byte("From: a@example.com\r\n\r\nBody"),
		"no TNEF tag": {0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	} {
		t.Run(name, func(t *testing.T) {
			msg, err := ParseTNEFMessageBytes(data)
			assert.Error(t, err)
			assert.Nil(t, msg)
		})
	}
}

func TestParseMessage_DetectsTNEF(t *testing.T) {
	// ParseMessage must recognize TNEF data by its signature
	// instead of trying to parse it as MIME message.
	msg, err := ParseMessage(tnefMessage(
		tnefStringProps(map[uint16]string{mapiBody: "TNEF body"}),
		tnefAttachment("a.txt", []byte("A")),
	))
	require.NoError(t, err)
	assert.Equal(t, "TNEF body", msg.Body)
	require.Len(t, msg.Attachments, 1)
	assert.Equal(t, "a.txt", msg.Attachments[0].Filename)
}

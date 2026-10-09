package email

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAttachment_MarshalJSON(t *testing.T) {
	j, err := json.Marshal(Attachment{
		PartID:      "PartID",
		ContentID:   "ContentID",
		ContentType: "ContentType",
		Filename:    "FileName",
		Content:     []byte("FileData"),
	})
	require.NoError(t, err, "json.Marshal")
	require.Equal(t, `{"PartID":"PartID","ContentID":"ContentID","ContentType":"ContentType","Filename":"FileName","Content":"RmlsZURhdGE="}`, string(j))
}

// TestNewAttachment protects that attachments built from raw bytes get a
// content type detected from the content (so they are rendered correctly
// by mail clients) and a unique Content-ID (so inline references from an
// HTML body can't point to the wrong attachment).
func TestNewAttachment(t *testing.T) {
	pdf := NewAttachment("1", "invoice.pdf", []byte("%PDF-1.7\n..."))
	require.Equal(t, "1", pdf.PartID)
	require.Equal(t, "invoice.pdf", pdf.Filename)
	require.Equal(t, []byte("%PDF-1.7\n..."), pdf.Content)
	require.Equal(t, "application/pdf", pdf.ContentType)
	require.False(t, pdf.Inline)
	require.False(t, pdf.OtherPart)
	require.NotEmpty(t, pdf.ContentID)

	txt := NewAttachment("2", "readme.txt", []byte("Hello"))
	require.Equal(t, "text/plain; charset=utf-8", txt.ContentType)
	require.NotEqual(t, pdf.ContentID, txt.ContentID, "every attachment needs its own Content-ID")
}

func TestAttachment_String(t *testing.T) {
	// The string is used for logging, so it must identify the attachment
	// and its size without dumping the possibly large binary content.
	a := &Attachment{PartID: "1.2", Filename: "invoice.pdf", Content: []byte("SECRET-CONTENT")}
	require.Equal(t, "Attachment{ID: `1.2`, File: `invoice.pdf`, Size: 14}", a.String())
}

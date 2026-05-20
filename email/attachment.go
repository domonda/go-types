package email

import (
	"fmt"
	"net/http"

	"github.com/domonda/go-types/uu"
)

// Attachment of an Email.
// Attachment implements fs.FileReader
type Attachment struct {
	// PartID is the MIME part identifier of the attachment.
	PartID string `json:",omitempty"`
	// ContentID is the value of the MIME Content-ID header,
	// used to reference inline attachments from HTML bodies.
	ContentID string `json:",omitempty"`
	// ContentType is the MIME media type of the attachment content.
	ContentType string `json:",omitempty"`
	// Inline indicates that the attachment is an inline part
	// referenced from the message body.
	Inline bool `json:",omitempty"`
	// OtherPart indicates a MIME part that is neither
	// a regular attachment nor an inline part.
	OtherPart bool `json:",omitempty"`
	// Filename is the name of the attachment file.
	Filename string
	// Content holds the raw bytes of the attachment.
	Content []byte
}

// NewAttachment returns a new Attachment with the passed partID, filename,
// and content. The ContentID is set to a random UUID and the ContentType
// is detected from the content.
func NewAttachment(partID, filename string, content []byte) *Attachment {
	return &Attachment{
		PartID:      partID,
		ContentID:   uu.IDv4().Hex(),
		ContentType: http.DetectContentType(content),
		Filename:    filename,
		Content:     content,
	}
}

// String implements the fmt.Stringer interface.
func (a *Attachment) String() string {
	return fmt.Sprintf("Attachment{ID: `%s`, File: `%s`, Size: %d}", a.PartID, a.Filename, len(a.Content))
}

package email

import (
	"fmt"
	"net/http"

	"github.com/domonda/go-types/uu"
)

// Attachment of an Email.
// Attachment implements fs.FileReader
type Attachment struct {
	PartID      string `json:",omitempty"`
	ContentID   string `json:",omitempty"`
	ContentType string `json:",omitempty"`
	Inline      bool   `json:",omitempty"`
	Filename    string
	Content     []byte
}

func NewAttachment(partID, filename string, content []byte) *Attachment {
	return &Attachment{
		PartID:      partID,
		ContentID:   uu.IDv4().Hex(),
		ContentType: http.DetectContentType(content),
		Filename:    filename,
		Content:     content,
	}
}

func (a *Attachment) String() string {
	return fmt.Sprintf("Attachment{ID: `%s`, File: `%s`, Size: %d}", a.PartID, a.Filename, len(a.Content))
}

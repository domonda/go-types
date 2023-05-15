package email

import (
	"context"
	"fmt"
	"net/http"

	"github.com/domonda/go-types/uu"
	"github.com/ungerik/go-fs"
)

// Attachment implements fs.FileReader
var _ fs.FileReader = new(Attachment)

// Attachment of an Email.
// Attachment implements fs.FileReader
type Attachment struct {
	PartID      string `json:"partID,omitempty"`
	ContentID   string `json:"contentID,omitempty"`
	ContentType string `json:"contentType,omitempty"`
	Inline      bool   `json:"inline,omitempty"`

	fs.MemFile
}

func NewAttachment(partID, filename string, content []byte) *Attachment {
	return &Attachment{
		PartID:      partID,
		ContentID:   uu.IDv4().Hex(),
		ContentType: http.DetectContentType(content),
		MemFile: fs.MemFile{
			FileName: filename,
			FileData: content,
		},
	}
}

func NewAttachmentReadFile(ctx context.Context, partID string, file fs.FileReader) (*Attachment, error) {
	data, err := file.ReadAllContext(ctx)
	if err != nil {
		return nil, err
	}
	return NewAttachment(partID, file.Name(), data), nil
}

func (a *Attachment) String() string {
	return fmt.Sprintf("Attachment{ID: `%s`, File: `%s`, Size: %d}", a.PartID, a.FileName, len(a.FileData))
}

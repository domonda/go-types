package email

import (
	"context"
	"fmt"
	"net/http"

	"github.com/domonda/go-types/uu"
	"github.com/ungerik/go-fs"
)

type Attachment struct {
	ContentID   string     `json:"contentID,omitempty"`
	ContentType string     `json:"contentType,omitempty"`
	File        fs.MemFile `json:"file"`
}

func (a *Attachment) FileReader() fs.FileReader { return &a.File }

func (a *Attachment) String() string {
	return fmt.Sprintf("Attachment{ID: `%s`, File: `%s`, Size: %d}", a.ContentID, a.File.FileName, len(a.File.FileData))
}

func NewAttachment(filename string, content []byte) *Attachment {
	return &Attachment{
		ContentID:   uu.IDv4().Hex(),
		ContentType: http.DetectContentType(content),
		File: fs.MemFile{
			FileName: filename,
			FileData: content,
		},
	}
}

func NewAttachmentReadFile(ctx context.Context, file fs.FileReader) (*Attachment, error) {
	data, err := file.ReadAllContext(ctx)
	if err != nil {
		return nil, err
	}
	return NewAttachment(file.Name(), data), nil
}

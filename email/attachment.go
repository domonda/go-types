package email

import (
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

func NewAttachmentReadFile(file fs.FileReader) (*Attachment, error) {
	data, err := file.ReadAll()
	if err != nil {
		return nil, err
	}
	return NewAttachment(file.Name(), data), nil
}

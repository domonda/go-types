package email

import (
	"bytes"
	"context"
	"fmt"

	"github.com/jaytaylor/html2text"
	"github.com/teamwork/tnef"
	"github.com/ungerik/go-fs"

	"github.com/domonda/go-errs"
	"github.com/domonda/go-types/nullable"
)

func ParseTNEFMessageBytes(messageBytes []byte) (msg *Message, err error) {
	defer errs.WrapWithFuncParams(&err, messageBytes)

	if len(messageBytes) < 4 {
		return nil, errs.New("message too short")
	}

	t, err := tnef.Decode(messageBytes)
	if err != nil {
		return nil, err
	}

	msg = &Message{
		Body:        string(t.Body),
		BodyHTML:    nullable.TrimmedStringFrom(string(t.BodyHTML)),
		ExtraHeader: make(Header),
	}
	if len(t.Body) <= 1 && len(t.BodyHTML) > len(t.Body) {
		html, err := html2text.FromReader(bytes.NewReader(t.BodyHTML))
		if err != nil {
			return nil, err
		}
		msg.BodyHTML.Set(html)
	}
	for i, attachment := range t.Attachments {
		msg.Attachments = append(msg.Attachments, &Attachment{
			PartID:  fmt.Sprintf("TNEF%d", i),
			Inline:  false, // ??
			MemFile: fs.MemFile{FileName: attachment.Title, FileData: attachment.Data},
		})
	}

	return msg, nil
}

func ParseTNEFMessageFile(ctx context.Context, file fs.FileReader) (msg *Message, err error) {
	defer errs.WrapWithFuncParams(&err, ctx, file)

	msgBytes, err := file.ReadAllContext(ctx)
	if err != nil {
		return nil, err
	}
	return ParseTNEFMessageBytes(msgBytes)
}

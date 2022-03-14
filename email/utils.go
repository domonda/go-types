package email

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"strings"
	"time"

	xhtml "golang.org/x/net/html"
	xurls "mvdan.cc/xurls/v2"

	"github.com/domonda/go-types/strutil"
)

func plaintextToHTML(text string) template.HTML {
	//template.HTML("<pre>" + text + "</pre>")
	return template.HTML(strings.ReplaceAll(text, "\n", "<br>")) //#nosec G203 -- not escaped HTML OK
}

var parseDateLayouts = []string{
	"02 Jan 2006 15:04:05 -0700",
	"Mon, 02 Jan 2006 15:04:05 -0700",
	"Mon, 02 Jan 2006 15:04:05 -0700 (MST)",
	"2 Jan 2006 15:04:05 -0700",
	"Mon, 2 Jan 2006 15:04:05 -0700",
	"Mon, 2 Jan 2006 15:04:05 -0700 (MST)",
	"Mon,  2 Jan 2006 15:04:05 MST",
}

func parseDate(date string) (*time.Time, error) {
	for _, layout := range parseDateLayouts {
		t, err := time.Parse(layout, date)
		if err == nil {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("can't parse email Date %q", date)
}

func formatDate(t *time.Time) string {
	if t == nil || t.IsZero() {
		return time.Now().Format(time.RFC1123Z)
	}
	return t.Format(time.RFC1123Z)
}

// ExtractBodyURLs returns all HTTP URLs from the message body
func ExtractBodyURLs(msg *Message) []string {
	regex := xurls.Strict()
	urls := make(strutil.StringSet)
	for _, url := range regex.FindAllString(msg.Body, -1) {
		if strings.HasPrefix(url, "http") {
			urls.Add(url)
		}
	}
	for _, url := range regex.FindAllString(string(msg.BodyHTML), -1) {
		if strings.HasPrefix(url, "http") {
			urls.Add(url)
		}
	}
	return urls.Sorted()
}

// func HTMLEmbedImages(html string, attachments []fs.FileReader) (inlinedHTML string, unusedAttachments []fs.FileReader) {

// 	panic("todo")
// }

// func embedAttachments(source string, attachments []*nmail.Attachment) string {
// 	for _, attachment := range attachments {
// 		for _, lookupTerm := range []string{attachment.ContentID, attachment.FileName} {
// 			// get base64 encoded image data for embedding
// 			data := getImageAsBase64(attachment.MimeType, attachment.Content)
// 			newContent := fmt.Sprintf(`src="%s"`, string(data))
// 			// since we use the filename in the regex we need quote it
// 			// special characters in filename should not affect regex query
// 			idQuoted := regexp.QuoteMeta(lookupTerm)
// 			regEx := fmt.Sprintf(`src=\"cid:%s.*"`, idQuoted)
// 			// compile regex and replace content
// 			r := regexp.MustCompile(regEx)
// 			source = string(r.ReplaceAll([]byte(source), []byte(newContent)))
// 		}
// 	}
// 	return source
// }

// HTMLToPlaintext converts HTML to plaintext
// by concaternating the content of text nodes
// with the passed delimiter between them.
// Whitespace is trimmed from the text nodes and
// nodes consisting only of whitespace are ignored.
// In case of an HTML parsing error all parsed text
// up until the error will be returned.
// Empty html will result in empty plaintext.
func HTMLToPlaintext(html []byte, delimiter string) (string, error) {
	var b strings.Builder
	tokenizer := xhtml.NewTokenizer(bytes.NewReader(html))
	for tt := tokenizer.Next(); tt != xhtml.ErrorToken; tt = tokenizer.Next() {
		if tt != xhtml.TextToken {
			continue
		}
		text := bytes.TrimSpace(tokenizer.Text())
		if len(text) == 0 {
			continue
		}
		if b.Len() > 0 {
			b.WriteString(delimiter)
		}
		b.Write(text)
	}
	if tokenizer.Err() != io.EOF {
		return b.String(), tokenizer.Err()
	}
	return b.String(), nil
}

package email_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/domonda/go-types/email"
	"github.com/domonda/go-types/nullable"
	"github.com/stretchr/testify/require"
)

func TestParseMessage(t *testing.T) {
	bytes, err := os.ReadFile("examplemessage_test.eml")
	require.NoError(t, err)

	msg, err := email.ParseMessage(bytes)
	require.NoError(t, err)

	equalsExamplemessage(t, msg, false)
}

func TestBuildRawMessage(t *testing.T) {
	// original
	bytes, err := os.ReadFile("examplemessage_test.eml")
	require.NoError(t, err)
	msg, err := email.ParseMessage(bytes)
	require.NoError(t, err)

	// build raw message from original
	bytes, err = msg.BuildRawMessage()
	require.NoError(t, err)

	// parse built raw message
	msg, err = email.ParseMessage(bytes)
	require.NoError(t, err)

	equalsExamplemessage(t, msg, true)
}

func equalsExamplemessage(t *testing.T, actual *email.Message, built bool) {
	actualAttachments := actual.Attachments
	actual.Attachments = nil // remove attachments for comparison, we'll only compare the hashes

	examplemessageDate, _ := time.Parse(time.RFC1123Z, "Thu, 22 Jan 2026 16:21:11 +0100")
	expected := &email.Message{
		MessageID:   nullable.TrimmedStringFrom("<94D8B6DE-E571-4F66-AAA7-2733E5A12434@denelop.com>"),
		From:        email.Address(`"John Doe" <john@doe.com>`),
		To:          email.AddressListJoinStrings("myinternal@invoicing.com"),
		DeliveredTo: email.NullableAddressListJoinStrings("denis+domonda@domonda.com", "somewhere+domonda@domonda.com"),
		ReplyTo:     email.NullableAddress(`"Jane Doe" <jane@doe.com>`),
		Subject:     "ok this is a test file check it out",
		Cc:          email.NullableAddressListJoinStrings(`"Mike Doe" <mike@doe.com>`),
		Date:        &examplemessageDate,
		Body:        "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque egestas eu arcu sit amet luctus. Phasellus ultrices velit et sem rhoncus, nec sollicitudin diam tristique. Pellentesque vehicula nec odio sit amet blandit. Praesent dictum consequat justo ut euismod. Aliquam varius rhoncus sem, ac semper elit eleifend et. Aliquam efficitur imperdiet ipsum, eget vestibulum tortor mollis ut. Nullam eleifend tellus eget massa iaculis, vitae facilisis elit malesuada. Morbi vel volutpat ex, sed ornare enim.\n\nOne\nTwo\nThree\nFour five\n\n￼\n￼￼\n￼￼\n\n", // contains unicode object replacement character U+FFFC, vscode might warn but it's correct
		BodyHTML:    nullable.TrimmedStringFrom(`<html aria-label="message body"><head><meta http-equiv="content-type" content="text/html; charset=us-ascii"></head><body style="overflow-wrap: break-word; -webkit-nbsp-mode: space; line-break: after-white-space;"><div><b>Lorem</b> ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque egestas eu arcu sit amet luctus. Phasellus ultrices velit<i> et sem</i> <span style="font-size: 36px;">rhoncus</span>, nec sollicitudin diam tristique. Pellentesque vehicula nec odio sit amet blandit. Praesent dictum consequat<b><i><u> justo ut euismod. Aliquam</u></i></b> <font color="#ffff9a">varius</font> rhoncu<strike>s sem, ac se</strike>mper elit eleifend et. Aliquam efficitur imperdiet ipsum, eget vestibulum tortor mollis ut. Nullam eleifend tellus eget massa iaculis, vitae facilisis elit malesuada. Morbi vel volutpat ex, sed ornare enim.</div><div><br></div><div><ul class="MailOutline"><li>One</li><li>Two</li><ol><li>Three</li><li>Four <b><i>five</i></b></li></ol></ul><div><b><i><br></i></b></div></div><div><b><i></i></b></div></body></html>`),
		ExtraHeader: email.Header{
			"Content-Type":    []string{"multipart/alternative; boundary=\"Apple-Mail=_766E5EE1-F3AA-4EA6-B1FB-3FE00B9843A2\""},
			"Mime-Version":    []string{"1.0 (Mac OS X Mail 16.0 \\(3864.300.41.1.7\\))"},
			"Received":        []string{"by XXXX:XXXX:XXXX:XXXX with SMTP id nb37csp577720ejc; Thu, 22 Jan 2026 07:21:29 -0800 (PST)"},
			"Return-Path":     []string{"<john@doe.com>"},
			"X-Mailer":        []string{"Apple Mail (2.3864.300.41.1.7)"},
			"X-Received":      []string{"by XXXX:XXXX:XXXX:XXXX with SMTP id ffacd0b85a97d-43569bbec67mr29502326f8f.34.1769095286652; Thu, 22 Jan 2026 07:21:26 -0800 (PST)"},
			"X-Original-From": []string{"NEBILY | Helena Törö <h.toeroe@nebily.com>"},
		},
	}

	// when building the message again, some headers are different
	if built {
		// content-type has an id that is newly generated
		expected.ExtraHeader["Content-Type"] = actual.ExtraHeader["Content-Type"]

		// mime-version is added by the builder
		expected.ExtraHeader["Mime-Version"] = append([]string{"1.0"}, expected.ExtraHeader["Mime-Version"]...)

		// body is carriage-return normalized by the builder
		expected.Body = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque egestas eu arcu sit amet luctus. Phasellus ultrices velit et sem rhoncus, nec sollicitudin diam tristique. Pellentesque vehicula nec odio sit amet blandit. Praesent dictum consequat justo ut euismod. Aliquam varius rhoncus sem, ac semper elit eleifend et. Aliquam efficitur imperdiet ipsum, eget vestibulum tortor mollis ut. Nullam eleifend tellus eget massa iaculis, vitae facilisis elit malesuada. Morbi vel volutpat ex, sed ornare enim.\r\n\r\nOne\r\nTwo\r\nThree\r\nFour five\r\n\r\n￼\r\n￼￼\r\n￼￼\r\n\r\n"
	}

	require.Equal(t, expected, actual)

	require.Len(t, actualAttachments, 5)

	// SHA-256 of each attachment's decoded content
	expectedHashes := map[string]string{
		"more invoices and stuff.zip":       "2002c7fc676e23732a28e806a8424b41ecf842f7d91f11c96499891a7d706399",
		"ReadMe.txt":                        "458f19173e7ed372d7c550a9b79f535f4a7265222a41ab8f75b1d817ed4283e8",
		"PastedGraphic-1.png":               "eeb05b27c3e19956cd6d46da06969aef01435568380095eb7d6196c330507411",
		"invoice_Yoseph Carroll_31061.tiff": "541c00ccbb8b1ad7f2ff898d2dd6b5896c38f181b55b6867aaa0d71296b47d92",
		"invoice_Aaron Bergman_36258.pdf":   "2e8206cd45c73701246757a641013aac483b4d58a9ee7ac3695c6f4b167c0101",
	}
	for _, attachment := range actualAttachments {
		hash := sha256.Sum256(attachment.Content)
		require.Equal(t, expectedHashes[attachment.Filename], hex.EncodeToString(hash[:]))
	}
}

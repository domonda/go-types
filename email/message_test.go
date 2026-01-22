package email_test

import (
	"testing"
	"time"

	"github.com/domonda/go-types/email"
	"github.com/domonda/go-types/nullable"
	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-fs"
)

func TestParseMessage(t *testing.T) {
	ctx := t.Context()

	eml := fs.File("examplemessage_test.eml")

	bytes, err := eml.ReadAllContext(ctx)
	require.NoError(t, err)

	msg, err := email.ParseMessage(bytes)
	require.NoError(t, err)

	equalsExamplemessage(t, msg, false)
}

func TestBuildRawMessage(t *testing.T) {
	ctx := t.Context()

	// original
	eml := fs.File("examplemessage_test.eml")
	bytes, err := eml.ReadAllContext(ctx)
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
		DeliveredTo: email.NullableAddress("denis+domonda@domonda.com"),
		ReplyTo:     email.NullableAddress(`"Jane Doe" <jane@doe.com>`),
		Subject:     "ok this is a test file check it out",
		Cc:          email.NullableAddressListJoinStrings(`"Mike Doe" <mike@doe.com>`),
		Date:        &examplemessageDate,
		Body:        "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque egestas eu arcu sit amet luctus. Phasellus ultrices velit et sem rhoncus, nec sollicitudin diam tristique. Pellentesque vehicula nec odio sit amet blandit. Praesent dictum consequat justo ut euismod. Aliquam varius rhoncus sem, ac semper elit eleifend et. Aliquam efficitur imperdiet ipsum, eget vestibulum tortor mollis ut. Nullam eleifend tellus eget massa iaculis, vitae facilisis elit malesuada. Morbi vel volutpat ex, sed ornare enim.\n\nOne\nTwo\nThree\nFour five\n\n￼\n￼￼\n￼￼\n\n", // contains unicode object replacement character U+FFFC, vscode might warn but it's correct
		BodyHTML:    nullable.TrimmedStringFrom(`<html aria-label="message body"><head><meta http-equiv="content-type" content="text/html; charset=us-ascii"></head><body style="overflow-wrap: break-word; -webkit-nbsp-mode: space; line-break: after-white-space;"><div><b>Lorem</b> ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque egestas eu arcu sit amet luctus. Phasellus ultrices velit<i> et sem</i> <span style="font-size: 36px;">rhoncus</span>, nec sollicitudin diam tristique. Pellentesque vehicula nec odio sit amet blandit. Praesent dictum consequat<b><i><u> justo ut euismod. Aliquam</u></i></b> <font color="#ffff9a">varius</font> rhoncu<strike>s sem, ac se</strike>mper elit eleifend et. Aliquam efficitur imperdiet ipsum, eget vestibulum tortor mollis ut. Nullam eleifend tellus eget massa iaculis, vitae facilisis elit malesuada. Morbi vel volutpat ex, sed ornare enim.</div><div><br></div><div><ul class="MailOutline"><li>One</li><li>Two</li><ol><li>Three</li><li>Four <b><i>five</i></b></li></ol></ul><div><b><i><br></i></b></div></div><div><b><i></i></b></div></body></html>`),
		ExtraHeader: email.Header{
			"Content-Type": []string{"multipart/alternative; boundary=\"Apple-Mail=_766E5EE1-F3AA-4EA6-B1FB-3FE00B9843A2\""},
			"Delivered-To": []string{"denis+domonda@domonda.com"},
			"Mime-Version": []string{"1.0 (Mac OS X Mail 16.0 \\(3864.300.41.1.7\\))"},
			"Received":     []string{"by XXXX:XXXX:XXXX:XXXX with SMTP id nb37csp577720ejc; Thu, 22 Jan 2026 07:21:29 -0800 (PST)"},
			"Return-Path":  []string{"<john@doe.com>"},
			"X-Mailer":     []string{"Apple Mail (2.3864.300.41.1.7)"},
			"X-Received":   []string{"by XXXX:XXXX:XXXX:XXXX with SMTP id ffacd0b85a97d-43569bbec67mr29502326f8f.34.1769095286652; Thu, 22 Jan 2026 07:21:26 -0800 (PST)"},
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

	expectedHashes := map[string]string{
		"more invoices and stuff.zip":       "f4d4dd017109789c8eeadf4fc7fcd31cb2406d534d94e6fd85b52fe1f8f610f6",
		"ReadMe.txt":                        "c8d257fa43d1fd263ce130342d0d3f22d28ad67f3a8ec3e3f9c5e32b8c38a2de",
		"PastedGraphic-1.png":               "9e8e28af259bab8ce625fcdba75699906638fdcaf126f17292894a972070e44c",
		"invoice_Yoseph Carroll_31061.tiff": "86d60a0faea42001861cf0239e92f78b80aea6233417410973ae500a8a30b671",
		"invoice_Aaron Bergman_36258.pdf":   "3440581a3ea3c7506dff48ed7d830770d79efa55de57f019e6e4b49df6b6d03e",
	}
	for _, attachment := range actualAttachments {
		att := fs.NewMemFile(attachment.Filename, attachment.Content)

		hash, err := att.ContentHash()
		require.NoError(t, err)

		require.Equal(t, expectedHashes[attachment.Filename], hash)
	}
}

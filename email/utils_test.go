package email

import (
	"html/template"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/domonda/go-types/strutil"
)

func TestHTMLToPlaintext(t *testing.T) {
	type args struct {
		html      []byte
		delimiter string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "nil",
			args: args{html: nil, delimiter: `|`},
			want: ``,
		},
		{
			name: "empty",
			args: args{html: []byte(``), delimiter: `|`},
			want: ``,
		},
		{
			name: "flat body",
			args: args{html: []byte(`<body>Hello &lt; World!</body>`), delimiter: `|`},
			want: `Hello < World!`,
		},
		{
			name: "divs",
			args: args{html: []byte(`<body><div>Hello</div><div>World!</div></body>`), delimiter: `|`},
			want: `Hello|World!`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HTMLToPlaintext(tt.args.html, tt.args.delimiter)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTMLToPlaintext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("HTMLToPlaintext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseDate(t *testing.T) {
	cetLocation, err := time.LoadLocation("CET")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		date    string
		want    time.Time
		wantErr bool
	}{
		{
			name: "Mon, 10 Feb 2020 11:32:11 +0000",
			date: "Mon, 10 Feb 2020 11:32:11 +0000",
			want: time.Date(2020, 2, 10, 11, 32, 11, 0, time.UTC),
		},
		{
			name: "Mon, 3 Feb 2020 11:32:11 +0000",
			date: "Mon, 3 Feb 2020 11:32:11 +0000",
			want: time.Date(2020, 2, 3, 11, 32, 11, 0, time.UTC),
		},
		{
			name: "Tue, 21 Jul 2020 09:39:53 +0000 (UTC)",
			date: "Tue, 21 Jul 2020 09:39:53 +0000 (UTC)",
			want: time.Date(2020, 7, 21, 9, 39, 53, 0, time.UTC),
		},
		{
			name: "Tue, 7 Jul 2020 09:39:53 +0000 (UTC)",
			date: "Tue, 7 Jul 2020 09:39:53 +0000 (UTC)",
			want: time.Date(2020, 7, 7, 9, 39, 53, 0, time.UTC),
		},
		{
			name: "Wed, 3 Nov 2021 10:13:07 +0100 (CET)",
			date: "Wed, 3 Nov 2021 10:13:07 +0100 (CET)",
			want: time.Date(2021, 11, 3, 10, 13, 7, 0, cetLocation),
		},
		{
			name: "8 Nov 2021 12:47:23 +0100",
			date: "8 Nov 2021 12:47:23 +0100",
			want: time.Date(2021, 11, 8, 12, 47, 23, 0, cetLocation),
		},
		{
			name: "Mon, 3 Jan 2022 11:00:51 GMT",
			date: "Mon, 3 Jan 2022 11:00:51 GMT",
			want: time.Date(2022, 1, 3, 11, 0, 51, 0, time.UTC),
		},
		{
			name: "Mon,  3 Jan 2022 11:00:51 GMT",
			date: "Mon,  3 Jan 2022 11:00:51 GMT",
			want: time.Date(2022, 1, 3, 11, 0, 51, 0, time.UTC),
		},
		{
			name: "Thu, 19 Feb 2026 06:12:35 +0100 (Mitteleuropaeische Zeit)",
			date: "Thu, 19 Feb 2026 06:12:35 +0100 (Mitteleuropaeische Zeit)",
			want: time.Date(2026, 2, 19, 06, 12, 35, 0, cetLocation),
		},
		{
			name: "Mon, 23 Feb 2026 11:42 +0100",
			date: "Mon, 23 Feb 2026 11:42 +0100",
			want: time.Date(2026, 2, 23, 11, 42, 0, 0, cetLocation),
		},
		{
			name: "27 Apr 2026 11:16 -0000",
			date: "27 Apr 2026 11:16 -0000",
			want: time.Date(2026, 4, 27, 11, 16, 0, 0, time.UTC),
		},
		{
			name: "8 Apr 2026 11:16 -0000",
			date: "8 Apr 2026 11:16 -0000",
			want: time.Date(2026, 4, 8, 11, 16, 0, 0, time.UTC),
		},
		{
			name: "24 Feb 26 18:52:49 GMT",
			date: "24 Feb 26 18:52:49 GMT",
			want: time.Date(2026, 2, 24, 18, 52, 49, 0, time.UTC),
		},
		{
			name: "27 Feb 2026 13:46:10 UT",
			date: "27 Feb 2026 13:46:10 UT",
			want: time.Date(2026, 2, 27, 13, 46, 10, 0, time.UTC),
		},
		{
			name: "01 Mar 2026 13:29:40 UT",
			date: "01 Mar 2026 13:29:40 UT",
			want: time.Date(2026, 3, 1, 13, 29, 40, 0, time.UTC),
		},
		{
			name: "1 Mar 2026 13:29:40 UT",
			date: "1 Mar 2026 13:29:40 UT",
			want: time.Date(2026, 3, 1, 13, 29, 40, 0, time.UTC),
		},
		{
			name: "Mon, 01 Mar 2026 13:29:40 UT",
			date: "Mon, 01 Mar 2026 13:29:40 UT",
			want: time.Date(2026, 3, 1, 13, 29, 40, 0, time.UTC),
		},
		{
			name: "Mon, 1 Mar 2026 13:29:40 UT",
			date: "Mon, 1 Mar 2026 13:29:40 UT",
			want: time.Date(2026, 3, 1, 13, 29, 40, 0, time.UTC),
		},
		{
			name: "24 Feb 26 18:52:49 UT",
			date: "24 Feb 26 18:52:49 UT",
			want: time.Date(2026, 2, 24, 18, 52, 49, 0, time.UTC),
		},
		// Invalid:
		{
			name:    "empty",
			date:    "",
			wantErr: true,
		},
		{
			name:    "error",
			date:    "error",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDate(tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !got.Equal(tt.want) {
				t.Errorf("parseDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestExtractBodyURLs covers the only function in utils.go that this branch
// changed (strutil.StringSet.Add was removed in favour of Insert). The set
// semantics it relies on are what make the swap safe, so they are asserted
// here rather than assumed: both bodies feed one set, a URL repeated within
// or across them appears once, and everything the regexp finds that is not
// an http(s) URL is dropped.
//
// Value: protects=ExtractBodyURLs unions the http(s) URLs of Body and
// BodyHTML into one deduplicated StringSet and drops non-http schemes;
// fails_when=the Insert replacing the removed Add stops deduplicating, one of
// the two bodies stops being scanned, or the "http" prefix filter is dropped;
// why_new=ExtractBodyURLs has no test at all, so the Add -> Insert swap is
// unprotected;
// seam=none
func TestExtractBodyURLs(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		bodyHTML string
		want     strutil.StringSet
	}{
		{
			name: "empty message",
			want: strutil.NewStringSet(),
		},
		{
			name: "plaintext body only",
			body: "see http://a.example.com and https://b.example.com/x?y=1",
			want: strutil.NewStringSet("http://a.example.com", "https://b.example.com/x?y=1"),
		},
		{
			name:     "HTML body only",
			bodyHTML: `<a href="https://c.example.com/path">c</a>`,
			want:     strutil.NewStringSet("https://c.example.com/path"),
		},
		{
			name:     "both bodies are unioned",
			body:     "http://a.example.com",
			bodyHTML: `<a href="https://c.example.com">c</a>`,
			want:     strutil.NewStringSet("http://a.example.com", "https://c.example.com"),
		},
		{
			name:     "duplicates collapse within and across the bodies",
			body:     "http://a.example.com twice http://a.example.com",
			bodyHTML: `<a href="http://a.example.com">again</a>`,
			want:     strutil.NewStringSet("http://a.example.com"),
		},
		{
			name:     "non-http schemes are dropped",
			body:     "write to mailto:a@example.com or fetch ftp://files.example.com/f",
			bodyHTML: `<a href="mailto:b@example.com">b</a> <a href="https://d.example.com">d</a>`,
			want:     strutil.NewStringSet("https://d.example.com"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &Message{Body: tt.body}
			msg.BodyHTML.Set(tt.bodyHTML)
			if got := ExtractBodyURLs(msg); !got.Equal(tt.want) {
				t.Errorf("ExtractBodyURLs() = %v, want %v", got.Sorted(), tt.want.Sorted())
			}
		})
	}
}

func Test_plaintextToHTML(t *testing.T) {
	// Line breaks of a quoted plaintext body must survive in the HTML reply.
	assert.Equal(t, template.HTML("Hello<br>World<br>"), plaintextToHTML("Hello\nWorld\n"))
	assert.Equal(t, template.HTML(""), plaintextToHTML(""))
}

// Test_formatDate protects that every built message gets a Date header
// (RFC 5322 requires it) that parseDate can read back, because
// BuildRawMessage writes it with formatDate and ParseMIMEMessage
// reads it with parseDate.
func Test_formatDate(t *testing.T) {
	t.Run("round-trip", func(t *testing.T) {
		date := time.Date(2026, 2, 19, 6, 12, 35, 0, time.FixedZone("CET", 3600))
		formatted := formatDate(&date)
		assert.Equal(t, "Thu, 19 Feb 2026 06:12:35 +0100", formatted)

		parsed, err := parseDate(formatted)
		require.NoError(t, err)
		assert.True(t, parsed.Equal(date), "parseDate(formatDate(%s)) = %s", date, parsed)
	})

	for name, date := range map[string]*time.Time{"nil": nil, "zero": {}} {
		t.Run(name+" uses current time", func(t *testing.T) {
			before := time.Now().Truncate(time.Second)
			parsed, err := parseDate(formatDate(date))
			require.NoError(t, err)
			after := time.Now()
			assert.False(t, parsed.Before(before), "%s is before %s", parsed, before)
			assert.False(t, parsed.After(after), "%s is after %s", parsed, after)
		})
	}
}

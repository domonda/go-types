package email

import (
	"testing"
	"time"
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

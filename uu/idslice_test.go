package uu

import (
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDSlice(t *testing.T) {
	expected := IDSliceMustFromStrings(
		"ec449f0f-e10c-4edb-8b59-0e6c896fdca5",
		"2d6a2c10-e4a6-45a3-a705-8115214a3778",
		"f3e52e97-e976-4a4c-a602-294310bcf935",
		"cc5873e6-286d-48cd-ae88-bda3a1e986e3",
	)

	jsonStr := `["ec449f0f-e10c-4edb-8b59-0e6c896fdca5","2d6a2c10-e4a6-45a3-a705-8115214a3778","f3e52e97-e976-4a4c-a602-294310bcf935","cc5873e6-286d-48cd-ae88-bda3a1e986e3"]`

	j, err := json.Marshal(expected)
	assert.NoError(t, err)
	assert.Equal(t, jsonStr, string(j))

	var parsed IDSlice
	err = json.Unmarshal([]byte(jsonStr), &parsed)
	assert.NoError(t, err)
	assert.Equal(t, expected, parsed)

	err = json.Unmarshal([]byte(`null`), &parsed)
	assert.NoError(t, err)
	assert.Nil(t, parsed)

	j, err = json.Marshal(nil)
	assert.NoError(t, err)
	assert.Equal(t, `null`, string(j))

	parsed = nil
	err = json.Unmarshal([]byte(`[]`), &parsed)
	assert.NoError(t, err)
	assert.Equal(t, IDSlice{}, parsed)

	j, err = json.Marshal(make(IDSlice, 0))
	assert.NoError(t, err)
	assert.Equal(t, `[]`, string(j))

	tests := []string{
		jsonStr,
		`"ec449f0f-e10c-4edb-8b59-0e6c896fdca5","2d6a2c10-e4a6-45a3-a705-8115214a3778","f3e52e97-e976-4a4c-a602-294310bcf935","cc5873e6-286d-48cd-ae88-bda3a1e986e3"`,
		`ec449f0f-e10c-4edb-8b59-0e6c896fdca5,2d6a2c10-e4a6-45a3-a705-8115214a3778,f3e52e97-e976-4a4c-a602-294310bcf935,cc5873e6-286d-48cd-ae88-bda3a1e986e3`,
		`[ec449f0f-e10c-4edb-8b59-0e6c896fdca5,2d6a2c10-e4a6-45a3-a705-8115214a3778,f3e52e97-e976-4a4c-a602-294310bcf935,cc5873e6-286d-48cd-ae88-bda3a1e986e3]`,
	}
	for _, str := range tests {
		t.Run(str, func(t *testing.T) {
			got, err := IDSliceFromString(str)
			assert.NoError(t, err)
			assert.Equal(t, expected, got)
		})
	}

	_, err = IDSliceFromString("")
	assert.Error(t, err)

	got, err := IDSliceFromStrings(nil)
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestIDSlice_Value(t *testing.T) {
	tests := []struct {
		name    string
		s       IDSlice
		want    driver.Value
		wantErr bool
	}{
		{name: "nil", s: nil, want: nil},
		{name: "len0", s: IDSlice{}, want: driver.Value(`{}`)},
		{name: "len1", s: IDSliceMustFromStrings("3004417b-25ca-441c-924f-102e571e5b5b"), want: driver.Value(`{"3004417b-25ca-441c-924f-102e571e5b5b"}`)},
		{name: "len2", s: IDSliceMustFromStrings("4a6ae04c-8718-4cea-929e-0d8071d328c7", "52d75836-03e0-4b38-8405-bbaa0f392d12"), want: driver.Value(`{"4a6ae04c-8718-4cea-929e-0d8071d328c7","52d75836-03e0-4b38-8405-bbaa0f392d12"}`)},
		{name: "len3", s: IDSliceMustFromStrings("2de0c5e3-d660-4ced-b929-f3a28a42849c", "b445f20b-d964-4fde-a360-39dc60b2157d", "2646cba5-4bc6-454f-bdd0-b869a2650f7e"), want: driver.Value(`{"2de0c5e3-d660-4ced-b929-f3a28a42849c","b445f20b-d964-4fde-a360-39dc60b2157d","2646cba5-4bc6-454f-bdd0-b869a2650f7e"}`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Value()
			if (err != nil) != tt.wantErr {
				t.Errorf("IDSlice.Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IDSlice.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIDSlice_Scan(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    IDSlice
		wantErr bool
	}{
		{name: "nil", value: nil, want: nil, wantErr: false},
		{name: "empty", value: ``, want: nil, wantErr: false},
		{name: "xxx", value: `xxx`, want: nil, wantErr: true},
		{
			name:  "21e70a0c-7f0e-44d8-8ae5-48992b89d0c5",
			value: `{"21e70a0c-7f0e-44d8-8ae5-48992b89d0c5"}`,
			want:  IDSlice{IDMust("21e70a0c-7f0e-44d8-8ae5-48992b89d0c5")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s IDSlice
			err := s.Scan(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("IDSlice.Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(s, tt.want) {
				t.Errorf("IDSlice.Scan() = %v, want %v", s, tt.want)
			}
		})
	}
}

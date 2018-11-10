package uu

import (
	"encoding/json"
	"testing"
)

func TestNullIDValueNil(t *testing.T) {
	u := NullID{}

	val, err := u.Value()
	if err != nil {
		t.Errorf("Error getting UUID value: %s", err)
	}

	if val != nil {
		t.Errorf("Wrong value returned, should be nil: %s", val)
	}
}

func TestNullIDScanValid(t *testing.T) {
	u := ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	s1 := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"

	u1 := NullID{}
	err := u1.Scan(s1)
	if err != nil {
		t.Errorf("Error unmarshaling NullID: %s", err)
	}

	if !u1.Valid {
		t.Errorf("NullID should be valid")
	}

	if u != u1.ID {
		t.Errorf("UUIDs should be equal: %s and %s", u, u1.ID)
	}
}

func TestNullIDScanNil(t *testing.T) {
	u := NullID{ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}, true}

	err := u.Scan(nil)
	if err != nil {
		t.Errorf("Error unmarshaling NullID: %s", err)
	}

	if u.Valid {
		t.Errorf("NullID should not be valid")
	}

	if u.ID != IDNil {
		t.Errorf("NullID value should be equal to Nil: %v", u)
	}
}

func TestNullID_MarshalUnmarshalJSON(t *testing.T) {
	u := NullID{ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}, true}
	var u2 NullID

	data, err := json.Marshal(&u)
	if err != nil {
		t.Errorf("Error JSON marshaling NullID: %s", err)
	}
	err = json.Unmarshal(data, &u2)
	if err != nil {
		t.Errorf("Error JSON unmarshaling NullID: %s", err)
	}
	if u2 != u {
		t.Errorf("JSON marshalling and unmarshalling produced a different UUID")
	}

	u.ID = IDNil
	u.Valid = false

	data, err = json.Marshal(&u)
	if err != nil {
		t.Errorf("Error JSON marshaling NullID: %s", err)
	}
	err = json.Unmarshal(data, &u2)
	if err != nil {
		t.Errorf("Error JSON unmarshaling NullID: %s", err)
	}
	if u2 != u {
		t.Errorf("JSON marshalling and unmarshalling produced a different UUID")
	}
}

func TestNullID_MarshalJSON(t *testing.T) {
	var testStruct struct {
		U ID     `json:"u"`
		N NullID `json:"n"`
	}
	data, err := json.Marshal(&testStruct)
	if err != nil {
		t.Errorf("Error JSON marshaling: %s", err)
	}
	if string(data) != `{"u":"00000000-0000-0000-0000-000000000000","n":null}` {
		t.Errorf("Marshalled wrong JSON: %s", string(data))
	}

	testStruct.U = ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	testStruct.N.SetValid(ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8})

	data, err = json.Marshal(&testStruct)
	if err != nil {
		t.Errorf("Error JSON marshaling: %s", err)
	}
	if string(data) != `{"u":"6ba7b810-9dad-11d1-80b4-00c04fd430c8","n":"6ba7b810-9dad-11d1-80b4-00c04fd430c8"}` {
		t.Errorf("Marshalled wrong JSON: %s", string(data))
	}
}

func TestNullID_UnmarshalJSON(t *testing.T) {
	type testStruct struct {
		U ID     `json:"u"`
		N NullID `json:"n"`
	}
	var out *testStruct
	err := json.Unmarshal([]byte(`{"u":"00000000-0000-0000-0000-000000000000","n":null}`), &out)
	if err != nil {
		t.Errorf("Error JSON unmarshaling: %s", err)
	}
	if out == nil {
		t.Errorf("Error JSON unmarshaling")
	}
	if out.U != IDNil || out.N.Valid {
		t.Errorf("Error JSON unmarshaling")
	}

	out = nil
	err = json.Unmarshal([]byte(`{"u":"6ba7b810-9dad-11d1-80b4-00c04fd430c8","n":"6ba7b810-9dad-11d1-80b4-00c04fd430c8"}`), &out)
	if err != nil {
		t.Errorf("Error JSON unmarshaling: %s", err)
	}
	if out == nil {
		t.Errorf("Error JSON unmarshaling")
	}
	ref := ID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	if out.U != ref || !out.N.Valid || out.N.ID != ref {
		t.Errorf("Error JSON unmarshaling")
	}
}

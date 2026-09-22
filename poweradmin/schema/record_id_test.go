// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT

package schema

import (
	"encoding/json"
	"testing"
)

func TestRecordIDUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    RecordID
		wantErr bool
	}{
		{name: "number", in: `42`, want: "42"},
		{name: "large number", in: `9007199254740993`, want: "9007199254740993"},
		{name: "numeric string", in: `"42"`, want: "42"},
		{name: "non-numeric string", in: `"rec-1"`, want: "rec-1"},
		{name: "null", in: `null`, want: ""},
		{name: "bool", in: `true`, wantErr: true},
		{name: "object", in: `{}`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got RecordID
			err := json.Unmarshal([]byte(tt.in), &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRecordUnmarshalNumericID(t *testing.T) {
	var r Record
	if err := json.Unmarshal([]byte(`{"id":1234,"name":"www","type":"A","content":"192.0.2.1","ttl":300,"priority":null}`), &r); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if r.ID != "1234" {
		t.Errorf("ID = %q, want 1234", r.ID)
	}
}

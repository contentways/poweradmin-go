// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT

package schema

import (
	"encoding/json"
	"testing"
)

func TestZoneCreateRequestOwnerEncoding(t *testing.T) {
	tests := []struct {
		name string
		req  ZoneCreateRequest
		want string
	}{
		{"omitted", ZoneCreateRequest{Name: "a.com", Type: "MASTER"}, `{"name":"a.com","type":"MASTER"}`},
		{"explicit null", ZoneCreateRequest{Name: "a.com", Type: "MASTER", OwnerUserID: OwnerUserIDNull(), GroupIDs: []int{2}},
			`{"name":"a.com","type":"MASTER","owner_user_id":null,"group_ids":[2]}`},
		{"user id", ZoneCreateRequest{Name: "a.com", Type: "MASTER", OwnerUserID: OwnerUserIDValue(7)},
			`{"name":"a.com","type":"MASTER","owner_user_id":7}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.req)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

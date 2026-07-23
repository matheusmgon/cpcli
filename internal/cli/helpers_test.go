package cli

import (
	"reflect"
	"testing"
)

func TestParseFields(t *testing.T) {
	cases := []struct {
		name    string
		fields  []string
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name:   "plain string value",
			fields: []string{"ip-address=1.2.3.4"},
			want:   map[string]interface{}{"ip-address": "1.2.3.4"},
		},
		{
			name:   "json array value",
			fields: []string{`members=["host1","host2"]`},
			want:   map[string]interface{}{"members": []interface{}{"host1", "host2"}},
		},
		{
			name:   "json bool and number values",
			fields: []string{"read-only=true", "port=8080"},
			want:   map[string]interface{}{"read-only": true, "port": float64(8080)},
		},
		{
			name:    "missing equals sign",
			fields:  []string{"not-a-kv-pair"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseFields(tc.fields)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestNameOrUIDPayload(t *testing.T) {
	t.Run("uid takes precedence over name", func(t *testing.T) {
		got, err := nameOrUIDPayload("some-name", "some-uid")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := map[string]interface{}{"uid": "some-uid"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %#v, want %#v", got, want)
		}
	})

	t.Run("falls back to name", func(t *testing.T) {
		got, err := nameOrUIDPayload("some-name", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := map[string]interface{}{"name": "some-name"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %#v, want %#v", got, want)
		}
	})

	t.Run("errors when neither is set", func(t *testing.T) {
		if _, err := nameOrUIDPayload("", ""); err == nil {
			t.Fatal("expected an error when neither --name nor --uid is set")
		}
	})
}

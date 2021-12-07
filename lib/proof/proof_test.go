package proof

import (
	"testing"
)

const uuid = "10000000-0000-0000-0000000000000000"

func TestQuery(t *testing.T) {
	type fields struct {
		Overwrite   bool
		AllProofs   bool
		HideMissing bool
	}
	no := fields{false, false, false}
	tests := []struct {
		name    string
		id      string
		fields  fields
		wantErr bool
	}{
		{"empty", "", no, true},
		{"missing", "1", no, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := Request{
				Overwrite:   tt.fields.Overwrite,
				AllProofs:   tt.fields.AllProofs,
				HideMissing: tt.fields.HideMissing,
			}
			if err := request.Query(tt.id); (err != nil) != tt.wantErr {
				t.Errorf("Request.Query() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sqlSelect(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want int
	}{
		{"empty", "", 141},
		{"id", "1", 154},
		{"uuid", uuid, 141},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sqlSelect(tt.id); len(got) != tt.want {
				t.Errorf("sqlSelect() = %v, want %v", len(got), tt.want)
			}
		})
	}
}

package proof

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/Defacto2/df2/lib/proof/internal/stat"
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

func TestTotal(t *testing.T) {
	zero := stat.Proof{
		Total: 0,
	}
	ten := stat.Proof{
		Total: 10,
	}
	type args struct {
		s       *stat.Proof
		request Request
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{}, ""},
		{"no req", args{&zero, Request{}}, ""},
		{
			"zero",
			args{&zero, Request{byID: "1"}},
			"file record id '1' does not exist or is not a release proof",
		},
		{"ten", args{&ten, Request{byID: "1"}}, "Total records 10"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := strings.TrimSpace(Total(tt.args.s, tt.args.request)); got != tt.want {
				t.Errorf("Total() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequest_Skip(t *testing.T) {
	type fields struct {
		Overwrite   bool
		AllProofs   bool
		HideMissing bool
		byID        string
	}
	tests := []struct {
		name   string
		fields fields
		values []sql.RawBytes
		want   bool
	}{
		{"empty", fields{}, nil, true},
		{"false", fields{byID: "1", Overwrite: true}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := Request{
				Overwrite:   tt.fields.Overwrite,
				AllProofs:   tt.fields.AllProofs,
				HideMissing: tt.fields.HideMissing,
				byID:        tt.fields.byID,
			}
			if got := request.Skip(tt.values); got != tt.want {
				t.Errorf("Request.Skip() = %v, want %v", got, tt.want)
			}
		})
	}
}

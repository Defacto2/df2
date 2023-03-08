package request_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/groups/internal/request"
)

func TestParse(t *testing.T) {
	t.Parallel()
	tmp := filepath.Join(os.TempDir(), "request.htm")
	type args struct {
		filename string
		templ    string
	}
	tests := []struct {
		name    string
		r       request.Flags
		args    args
		wantErr bool
	}{
		{"empty", request.Flags{Filter: "bbs"}, args{tmp, ""}, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// note: this test is slow
			t.Parallel()
			if err := tt.r.Parse(tt.args.filename, tt.args.templ); (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPrintr(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		r         request.Flags
		wantTotal bool
		wantErr   bool
	}{
		{"", request.Flags{}, true, false},
		{"bbs", request.Flags{Filter: "bbs"}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			t.Parallel()
			gotTotal, err := request.Print(tt.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Print() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (gotTotal > 0) != tt.wantTotal {
				t.Errorf("Print() = %v, want %v", gotTotal, tt.wantTotal)
			}
		})
	}
}

func TestInitialism(t *testing.T) {
	type args struct {
		group string
	}
	tests := []struct {
		name     string
		r        request.Flags
		args     args
		wantName string
		wantErr  bool
	}{
		{"empty", request.Flags{}, args{""}, "", false},
		{"none", request.Flags{Initialisms: false}, args{"Defacto2"}, "", false},
		{"Defacto2", request.Flags{Initialisms: true}, args{"Defacto2"}, "DF2", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, err := tt.r.Initialism(tt.args.group)
			if (err != nil) != tt.wantErr {
				t.Errorf("Initialism() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotName != tt.wantName {
				t.Errorf("Initialism() = %v, want %v", gotName, tt.wantName)
			}
		})
	}
}

func TestFiles(t *testing.T) {
	type args struct {
		group string
	}
	tests := []struct {
		name      string
		r         request.Flags
		args      args
		wantTotal bool
		wantErr   bool
	}{
		{"empty", request.Flags{}, args{""}, false, false},
		{"none", request.Flags{Counts: false}, args{"Defacto2"}, false, false},
		{"Defacto2", request.Flags{Counts: true}, args{"Defacto2"}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTotal, err := tt.r.Files(tt.args.group)
			if (err != nil) != tt.wantErr {
				t.Errorf("Request.files() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (gotTotal > 0) != tt.wantTotal {
				t.Errorf("Request.files() = %v, want %v", gotTotal, tt.wantTotal)
			}
		})
	}
}

package archive

import (
	"testing"
)

func TestExtract(t *testing.T) {
	type args struct {
		archive  string
		filename string
		uuid     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{}, true},
		{"error", args{testDir("demozoo/test.zip"), "test.zip", "xxxx"}, true},
		{"zip", args{testDir("demozoo/test.zip"), "test.zip", "6ba7b814-9dad-11d1-80b4-00c04fd430c8"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Extract(tt.args.archive, tt.args.filename, tt.args.uuid); (err != nil) != tt.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

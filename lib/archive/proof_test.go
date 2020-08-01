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
		{"7z", args{testDir("demozoo/test.7z"), "test.7z", "6ba7b814-9dad-11d1-80b4-00c04fd430c8"}, true},
		{"tar", args{testDir("demozoo/test.tar"), "test.tar", "6ba7b814-9dad-11d1-80b4-00c04fd430c8"}, false},
		{"tar.gz", args{testDir("demozoo/test.tar.gz"), "test.tar.gz", "6ba7b814-9dad-11d1-80b4-00c04fd430c8"}, false},
		{"tar.bz2", args{testDir("demozoo/test.tar.bz2"), "test.tar.bz2", "6ba7b814-9dad-11d1-80b4-00c04fd430c8"}, false},
		{"sz (snappy)", args{testDir("demozoo/test.tar.sz"), "test.tar.sz", "6ba7b814-9dad-11d1-80b4-00c04fd430c8"}, false},
		{"xz", args{testDir("demozoo/test.tar.xz"), "test.tar.xz", "6ba7b814-9dad-11d1-80b4-00c04fd430c8"}, false},
		{"zip", args{testDir("demozoo/test.zip"), "test.zip", "6ba7b814-9dad-11d1-80b4-00c04fd430c8"}, false},
		{"zip.bz2", args{testDir("demozoo/test.bz2.zip"), "test.bz2.zip", "6ba7b814-9dad-11d1-80b4-00c04fd430c8"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Extract(tt.args.archive, tt.args.filename, tt.args.uuid); (err != nil) != tt.wantErr {
				t.Errorf("Extract(%s) error = %v, wantErr %v", tt.args.archive, err, tt.wantErr)
			}
		})
	}
}

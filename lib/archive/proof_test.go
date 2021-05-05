package archive

import (
	"testing"

	"github.com/spf13/viper"
)

func TestExtract(t *testing.T) {
	const uuid = "6ba7b814-9dad-11d1-80b4-00c04fd430c8"
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
		{"7z", args{testDir("demozoo/test.7z"), "test.7z", uuid}, true},
		{"tar", args{testDir("demozoo/test.tar"), "test.tar", uuid}, false},
		{"tar.gz", args{testDir("demozoo/test.tar.gz"), "test.tar.gz", uuid}, false},
		{"tar.bz2", args{testDir("demozoo/test.tar.bz2"), "test.tar.bz2", uuid}, false},
		{"sz (snappy)", args{testDir("demozoo/test.tar.sz"), "test.tar.sz", uuid}, false},
		{"xz", args{testDir("demozoo/test.tar.xz"), "test.tar.xz", uuid}, false},
		{"zip", args{testDir("demozoo/test.zip"), "test.zip", uuid}, false},
		{"zip.bz2", args{testDir("demozoo/test.bz2.zip"), "test.bz2.zip", uuid}, false},
	}
	for _, tt := range tests {
		if viper.GetString("directory.root") == "" {
			return
		}
		t.Run(tt.name, func(t *testing.T) {
			if err := Extract(tt.args.archive, tt.args.filename, tt.args.uuid); (err != nil) != tt.wantErr {
				t.Errorf("Extract(%s) error = %v, wantErr %v", tt.args.archive, err, tt.wantErr)
			}
		})
	}
}

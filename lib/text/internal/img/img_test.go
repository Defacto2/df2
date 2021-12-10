package img_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/lib/text/internal/img"
	"github.com/spf13/viper"
)

const (
	root    = "../../../../tests"
	dir     = "../../../../tests/text"
	imgs    = "../../../../tests/images"
	uuidDir = "../../../../tests/uuid/"
)

// config the directories expected by ansilove.
func config(t *testing.T) string {
	d, err := filepath.Abs(uuidDir)
	if err != nil {
		t.Error(err)
	}
	viper.Set("directory.000", d)
	viper.Set("directory.400", d)
	return d
}

func TestMakePng(t *testing.T) {
	dst, err := filepath.Abs(dir)
	if err != nil {
		t.Error(err)
	}
	src := filepath.Join(dst, "test.txt")
	png, err := filepath.Abs(filepath.Join(root, "text.png"))
	if err != nil {
		t.Error(err)
	}

	type args struct {
		src   string
		dest  string
		amiga bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{"", "", false}, true},
		{"missing src", args{"", dst, false}, true},
		{"missing dst", args{src, "", false}, true},
		{"invalid src", args{src + "invalidate", dst, false}, true},
		{"text", args{src, dst, false}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := img.MakePng(tt.args.src, tt.args.dest, tt.args.amiga)
			if (err != nil) != tt.wantErr {
				t.Errorf("makePng() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			defer os.Remove(png)
		})
	}
}

func Test_generate(t *testing.T) {
	d, err := filepath.Abs(dir)
	if err != nil {
		t.Error(err)
	}
	gif := filepath.Join(d, "test.gif")
	txt := filepath.Join(d, "test.txt")
	config(t)
	ud, err := filepath.Abs(uuidDir)
	if err != nil {
		t.Error(err)
	}
	type args struct {
		name  string
		id    string
		amiga bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"test", args{"", "", false}, true},
		{"missing", args{"abce", "1", false}, true},
		{"gif", args{gif, "1", false}, true},
		{"okay txt", args{txt, "1", false}, false},
		{"okay amiga", args{txt, "1", true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := img.Generate(tt.args.name, tt.args.id, tt.args.amiga)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			os.Remove(filepath.Join(ud, tt.args.id+".png"))
			os.Remove(filepath.Join(ud, tt.args.id+".webp"))
		})
	}
}

func createLongFile(name string) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()
	for i := 1; i < 600; i++ {
		if _, err := f.WriteString(fmt.Sprintf("line %d\n", i)); err != nil {
			return err
		}
	}
	return nil
}

func TestReduce(t *testing.T) {
	long := filepath.Join(dir, "tooLong.txt")
	if err := createLongFile(long); err != nil {
		t.Error(err)
	}
	type args struct {
		src  string
		uuid string
	}
	tests := []struct {
		name     string
		args     args
		wantSize int64
		wantErr  bool
	}{
		{"empty", args{}, 0, true},
		{"okay", args{src: long}, 4000, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := img.Reduce(tt.args.src, tt.args.uuid)
			if (err != nil) != tt.wantErr {
				t.Errorf("Reduce() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == "" {
				return
			}
			defer os.Remove(got)
			defer os.Remove(long)
			if gotStat, err := os.Stat(got); err != nil {
				t.Errorf("Reduce() stat error: %v", err)
				return
			} else if gotStat.Size() < tt.wantSize {
				t.Errorf("Reduce() created a trimmed file of %dB, want at least %dB", gotStat.Size(), tt.wantSize)
			}
		})
	}
}

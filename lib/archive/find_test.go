package archive

import "testing"

//nolint: revive
func TestFindNFO(t *testing.T) {
	var empty []string
	const (
		ff2 = "hi.nfo"
		ff3 = "random.txt"
	)
	type args struct {
		name  string
		files []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{"", empty}, ""},
		{"empty zip", args{"hi.zip", empty}, ""},
		{"1 file", args{"hi.zip", []string{fileDiz}}, fileDiz},
		{"2 files", args{"hi.zip", []string{fileDiz, ff2}}, "hi.nfo"},
		{"3 files", args{"hi.zip", []string{fileDiz, ff2, ff3}}, "hi.nfo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FindNFO(tt.args.name, tt.args.files...); got != tt.want {
				t.Errorf("FindNFO() = %v, want %v", got, tt.want)
			}
		})
	}
}

package assets

import (
	"reflect"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func Test_ignoreList(t *testing.T) {
	var want struct{}
	if got := ignoreList("")["blank.png"]; !reflect.DeepEqual(got, want) {
		t.Errorf("ignoreList() = %v, want %v", got, want)
	}

}

func Test_targets(t *testing.T) {
	type args struct {
		target string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"", args{"all"}, 6},
		{"", args{"image"}, 3},
		{"error", args{""}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := targets(tt.args.target); got != tt.want {
				t.Errorf("targets() = %v, want %v", got, tt.want)
			}
		})
	}
}

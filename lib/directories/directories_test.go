package directories

import (
	"testing"
)

// func TestInit(t *testing.T) {
// 	t.Run("", func(t *testing.T) {
// 		if got := Init(false); !reflect.DeepEqual(got.Img000, "/") {
// 			t.Errorf("Init() = %q, want %q", got.Img000, "/")
// 		}
// 	})
// }

// func TestFiles(t *testing.T) {
// 	t.Run("", func(t *testing.T) {
// 		if got := Files("defacto2"); !reflect.DeepEqual(got.Img000, "/") {
// 			t.Errorf("Init() = %q, want %q", got.Img000, "/")
// 		}
// 	})
// }

func Test_randStringBytes(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"", args{0}, 0},
		{"", args{1}, 1},
		{"", args{10}, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := randStringBytes(tt.args.n); len(got) != tt.want {
				t.Errorf("randStringBytes() = %v, want %v", len(got), tt.want)
			}
		})
	}
}

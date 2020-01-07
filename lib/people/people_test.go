package people

import "testing"

func Test_roles(t *testing.T) {
	type args struct {
		r string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{""}, "wmca"},
		{"", args{"artists"}, "a"},
		{"", args{"a"}, "a"},
		{"", args{"all"}, "wmca"},
		{"error", args{"xxx"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := roles(tt.args.r); got != tt.want {
				t.Errorf("roles() = %v, want %v", got, tt.want)
			}
		})
	}
}

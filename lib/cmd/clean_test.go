package cmd

import "testing"

func Test_options(t *testing.T) {
	type args struct {
		a []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{}, "\noptions: "},
		{"targets", args{targets}, "\noptions: all,download,emulation,image"},
		{"test", args{[]string{"test"}}, "\noptions: test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := options(tt.args.a...); got != tt.want {
				t.Errorf("options() = %q, want %q", got, tt.want)
			}
		})
	}
}

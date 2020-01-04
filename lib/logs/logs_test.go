package logs

import "testing"

func Test_ProgressPct(t *testing.T) {
	type args struct {
		name  string
		count int
		total int
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{"", args{"", 1, 10}, float64(10)},
		{"", args{"", 10, 10}, float64(100)},
		{"", args{"", 0, 10}, float64(0)},
		{"", args{"", -1, 10}, float64(-10)},
		{"", args{"", 1, 99999}, float64(0.001000010000100001)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ProgressPct(tt.args.name, tt.args.count, tt.args.total); got != tt.want {
				t.Errorf("ProgressPct() = %v, want %v", got, tt.want)
			}
		})
	}
}

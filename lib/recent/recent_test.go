package recent

import "testing"

func TestList(t *testing.T) {
	type args struct {
		limit    uint
		compress bool
	}
	tests := []struct {
		name string
		args args
	}{
		{"standard", args{1, false}},
		{"compress", args{1, true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			List(tt.args.limit, tt.args.compress)
		})
	}
}

package recent

import (
	"testing"
)

func TestList(t *testing.T) {
	type args struct {
		limit    uint
		compress bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"standard", args{1, false}, false},
		{"compress", args{1, true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := List(tt.args.limit, tt.args.compress); (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

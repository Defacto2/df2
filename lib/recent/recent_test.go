package recent

import (
	"log"
	"testing"
)

func BenchmarkCreate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := List(10, true); err != nil {
			log.Print(err)
		}
	}
}

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

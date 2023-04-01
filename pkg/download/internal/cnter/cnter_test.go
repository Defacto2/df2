package cnter_test

import (
	"testing"

	"github.com/Defacto2/df2/pkg/download/internal/cnter"
)

func TestWrite(t *testing.T) {
	t.Parallel()
	type fields struct {
		Name    string
		Total   uint64
		Written uint64
	}
	type args struct {
		p []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{"empty", fields{"", 0, 0}, args{[]byte("")}, 0, false},
		{"hi", fields{"hi", 2, 2}, args{[]byte("hi")}, 2, false},
		{"some filler text", fields{"x", 2, 6}, args{[]byte("some filler text")}, 16, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			wc := &cnter.Writer{
				Name:    tt.fields.Name,
				Total:   tt.fields.Total,
				Written: tt.fields.Written,
			}
			got, err := wc.Write(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Write() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPercent(t *testing.T) {
	t.Parallel()
	type args struct {
		count uint64
		total uint64
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{"0", args{0, 0}, 0},
		{"1", args{1, 100}, 1},
		{"80", args{33, 41}, 80},
		{"100", args{100, 100}, 100},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := cnter.Percent(tt.args.count, tt.args.total); got != tt.want {
				t.Errorf("Percent() = %v, want %v", got, tt.want)
			}
		})
	}
}

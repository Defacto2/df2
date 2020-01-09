package database

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

func TestID(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"blank", args{""}, false},
		{"letters", args{"abcde"}, false},
		{"zeros", args{"00000"}, false},
		{"zeros", args{"00000876786"}, false},
		{"negative", args{"-1"}, false},
		{"valid 1", args{"1"}, true},
		{"valid 9", args{"99999"}, true},
		{"float", args{"1.0000"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ID(tt.args.id); got != tt.want {
				t.Errorf("ID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUUID(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"", args{"x"}, false},
		{"", args{"0000"}, false},
		{"", args{""}, false},
		{"zeros", args{"00000000-0000-0000-0000-000000000000"}, true},
		{"random", args{uuid.New().String()}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UUID(tt.args.id); got != tt.want {
				t.Errorf("UUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

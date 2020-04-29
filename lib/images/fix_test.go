package images

import (
	"testing"
)

func TestImg_check(t *testing.T) {
	tests := []struct {
		name string
		i    Img
		want bool
	}{
		{"not exist", Img{UUID: "false id"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.check(); got != tt.want {
				t.Errorf("Img.check() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImg_validExt(t *testing.T) {
	tests := []struct {
		name   string
		i      Img
		wantOk bool
	}{
		{"empty", Img{}, false},
		{"text", Img{Filename: "some.txt"}, false},
		{"png", Img{Filename: "some.png"}, true},
		{"jpeg", Img{Filename: "some other.jpeg"}, true},
		{"jpeg", Img{Filename: "some.other.jpeg"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOk := tt.i.validExt(); gotOk != tt.wantOk {
				t.Errorf("Img.validExt() = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestFix(t *testing.T) {
	tests := []struct {
		name     string
		simulate bool
		wantErr  bool
	}{
		{"name", true, false},
		{"name", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Fix(tt.simulate); (err != nil) != tt.wantErr {
				t.Errorf("Fix() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

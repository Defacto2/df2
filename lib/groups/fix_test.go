package groups

import "testing"

func Test_toClean(t *testing.T) {
	tests := []struct {
		name   string
		wantOk bool
	}{
		{"", false},
		{"Defacto2", false},
		{"defacto2", true},
		{"d-e-f-a-c-t-o-2", true},
		{"d_f", true},
		{"D2", false},
	}
	sim = false
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOk := toClean(tt.name); gotOk != tt.wantOk {
				t.Errorf("toClean() = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

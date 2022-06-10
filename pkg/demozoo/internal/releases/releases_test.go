package releases

import (
	"testing"
)

func TestSite(t *testing.T) {
	tests := []struct {
		title string
		want  string
	}{
		{"", ""},
		{"Hello world", ""},
		{"Pool Of Radiance BBS", "Pool Of Radiance BBS"},
		{"The Void BBS (1)", "Void BBS"},
		{"The Maximum Security FTP (2a)", "Maximum Security FTP"},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			if got := Site(tt.title); got != tt.want {
				t.Errorf("Site() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProductionV1_Released(t *testing.T) {
	tests := []struct {
		ReleaseDate string
		wantYear    int
		wantMonth   int
		wantDay     int
	}{
		{"", 0, 0, 0},
		{"1970", 1970, 0, 0},
		{"1970-01", 1970, 1, 0},
		{"1970-01-01", 1970, 1, 1},
		{"1970-11-31", 1970, 11, 31},
	}
	for _, tt := range tests {
		t.Run(tt.ReleaseDate, func(t *testing.T) {
			p := ProductionV1{
				ReleaseDate: tt.ReleaseDate,
			}
			gotYear, gotMonth, gotDay := p.Released()
			if gotYear != tt.wantYear {
				t.Errorf("ProductionV1.Released() gotYear = %v, want %v", gotYear, tt.wantYear)
			}
			if gotMonth != tt.wantMonth {
				t.Errorf("ProductionV1.Released() gotMonth = %v, want %v", gotMonth, tt.wantMonth)
			}
			if gotDay != tt.wantDay {
				t.Errorf("ProductionV1.Released() gotDay = %v, want %v", gotDay, tt.wantDay)
			}
		})
	}
}

package demozoo

import (
	"testing"
	"time"
)

func TestProduction_URL(t *testing.T) {
	type fields struct {
		ID         int64
		Timeout    time.Duration
		link       string
		StatusCode int
		Status     string
	}
	tests := []struct {
		name     string
		fields   fields
		wantLink string
		wantErr  bool
	}{
		{"empty", fields{}, "https://demozoo.org/api/v1/productions/0?format=json", false},
		{"1", fields{ID: 1}, "https://demozoo.org/api/v1/productions/1?format=json", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Production{
				ID:         tt.fields.ID,
				Timeout:    tt.fields.Timeout,
				link:       tt.fields.link,
				StatusCode: tt.fields.StatusCode,
				Status:     tt.fields.Status,
			}
			if err := p.URL(); (err != nil) != tt.wantErr {
				t.Errorf("Production.URL() error = %v, wantErr %v", err, tt.wantErr)
			}
			if p.link != tt.wantLink {
				t.Errorf("Production.URL() link = %v, want %v", p.link, tt.wantLink)
			}
		})
	}
}
